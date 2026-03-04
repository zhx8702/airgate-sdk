package grpc

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	sdk "github.com/DouDOU-start/airgate-sdk"
	pb "github.com/DouDOU-start/airgate-sdk/proto"
)

// SimpleGatewayGRPCServer 将 SimpleGatewayPlugin 包装为 gRPC 服务端
type SimpleGatewayGRPCServer struct {
	pb.UnimplementedSimpleGatewayServiceServer
	Impl sdk.SimpleGatewayPlugin
}

func (s *SimpleGatewayGRPCServer) GetPlatform(_ context.Context, _ *pb.Empty) (*pb.StringResponse, error) {
	return &pb.StringResponse{Value: s.Impl.Platform()}, nil
}

func (s *SimpleGatewayGRPCServer) GetModels(_ context.Context, _ *pb.Empty) (*pb.ModelsResponse, error) {
	models := s.Impl.Models()
	resp := &pb.ModelsResponse{}
	for _, m := range models {
		resp.Models = append(resp.Models, &pb.ModelInfoProto{
			Id:          m.ID,
			Name:        m.Name,
			MaxTokens:   int32(m.MaxTokens),
			InputPrice:  m.InputPrice,
			OutputPrice: m.OutputPrice,
			CachePrice:  m.CachePrice,
		})
	}
	return resp, nil
}

func (s *SimpleGatewayGRPCServer) GetRoutes(_ context.Context, _ *pb.Empty) (*pb.RoutesResponse, error) {
	routes := s.Impl.Routes()
	resp := &pb.RoutesResponse{}
	for _, r := range routes {
		resp.Routes = append(resp.Routes, &pb.RouteDefinitionProto{
			Method:      r.Method,
			Path:        r.Path,
			Description: r.Description,
		})
	}
	return resp, nil
}

func (s *SimpleGatewayGRPCServer) Forward(ctx context.Context, req *pb.ForwardRequest) (*pb.ForwardResult, error) {
	// 反序列化 credentials
	var creds map[string]string
	if len(req.CredentialsJson) > 0 {
		_ = json.Unmarshal(req.CredentialsJson, &creds)
	}

	// 构建 SDK ForwardRequest
	headers := make(http.Header)
	for k, v := range req.Headers {
		headers.Set(k, v)
	}

	fwdReq := &sdk.ForwardRequest{
		Account: &sdk.Account{
			ID:             req.AccountId,
			Credentials:    creds,
			ProxyURL:       req.ProxyUrl,
			RateMultiplier: req.RateMultiplier,
			MaxConcurrency: int(req.MaxConcurrency),
		},
		Body:    req.Body,
		Headers: headers,
		Model:   req.Model,
		Stream:  req.Stream,
		// Writer 在非流式模式下不需要
	}

	result, err := s.Impl.Forward(ctx, fwdReq)
	if err != nil {
		return nil, err
	}

	return &pb.ForwardResult{
		StatusCode:   int32(result.StatusCode),
		InputTokens:  int32(result.InputTokens),
		OutputTokens: int32(result.OutputTokens),
		CacheTokens:  int32(result.CacheTokens),
		Model:        result.Model,
		DurationMs:   result.Duration.Milliseconds(),
	}, nil
}

func (s *SimpleGatewayGRPCServer) ForwardStream(req *pb.ForwardRequest, stream pb.SimpleGatewayService_ForwardStreamServer) error {
	// 反序列化 credentials
	var creds map[string]string
	if len(req.CredentialsJson) > 0 {
		_ = json.Unmarshal(req.CredentialsJson, &creds)
	}

	headers := make(http.Header)
	for k, v := range req.Headers {
		headers.Set(k, v)
	}

	// 创建一个流式写入器，将 HTTP ResponseWriter 写入到 gRPC 流
	sw := &streamWriter{stream: stream}

	fwdReq := &sdk.ForwardRequest{
		Account: &sdk.Account{
			ID:             req.AccountId,
			Credentials:    creds,
			ProxyURL:       req.ProxyUrl,
			RateMultiplier: req.RateMultiplier,
			MaxConcurrency: int(req.MaxConcurrency),
		},
		Body:    req.Body,
		Headers: headers,
		Model:   req.Model,
		Stream:  true,
		Writer:  sw,
	}

	startTime := time.Now()
	result, err := s.Impl.Forward(stream.Context(), fwdReq)
	if err != nil {
		return err
	}

	// 发送最终结果
	return stream.Send(&pb.ForwardChunk{
		Done: true,
		FinalResult: &pb.ForwardResult{
			StatusCode:   int32(result.StatusCode),
			InputTokens:  int32(result.InputTokens),
			OutputTokens: int32(result.OutputTokens),
			CacheTokens:  int32(result.CacheTokens),
			Model:        result.Model,
			DurationMs:   time.Since(startTime).Milliseconds(),
		},
	})
}

func (s *SimpleGatewayGRPCServer) ValidateCredentials(ctx context.Context, req *pb.CredentialsRequest) (*pb.Empty, error) {
	validator, ok := s.Impl.(sdk.AccountValidator)
	if !ok {
		return &pb.Empty{}, nil
	}
	if err := validator.ValidateCredentials(ctx, req.Credentials); err != nil {
		return nil, err
	}
	return &pb.Empty{}, nil
}

// streamWriter 将 gRPC 流包装为 http.ResponseWriter
type streamWriter struct {
	stream  pb.SimpleGatewayService_ForwardStreamServer
	headers http.Header
	code    int
}

func (w *streamWriter) Header() http.Header {
	if w.headers == nil {
		w.headers = make(http.Header)
	}
	return w.headers
}

func (w *streamWriter) Write(data []byte) (int, error) {
	err := w.stream.Send(&pb.ForwardChunk{
		Data: data,
	})
	if err != nil {
		return 0, err
	}
	return len(data), nil
}

func (w *streamWriter) WriteHeader(statusCode int) {
	w.code = statusCode
}
