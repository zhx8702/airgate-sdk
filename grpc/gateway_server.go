package grpc

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	sdk "github.com/DouDOU-start/airgate-sdk"
	pb "github.com/DouDOU-start/airgate-sdk/proto"
)

// GatewayGRPCServer 将 GatewayPlugin 包装为 gRPC 服务端
type GatewayGRPCServer struct {
	pb.UnimplementedGatewayServiceServer
	Impl sdk.GatewayPlugin
}

func (s *GatewayGRPCServer) GetPlatform(_ context.Context, _ *pb.Empty) (*pb.StringResponse, error) {
	return &pb.StringResponse{Value: s.Impl.Platform()}, nil
}

func (s *GatewayGRPCServer) GetModels(_ context.Context, _ *pb.Empty) (*pb.ModelsResponse, error) {
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

func (s *GatewayGRPCServer) GetRoutes(_ context.Context, _ *pb.Empty) (*pb.RoutesResponse, error) {
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

// buildAccount 从 proto ForwardRequest 构建 SDK Account
func buildAccount(req *pb.ForwardRequest) *sdk.Account {
	var creds map[string]string
	if len(req.CredentialsJson) > 0 {
		_ = json.Unmarshal(req.CredentialsJson, &creds)
	}
	return &sdk.Account{
		ID:          req.AccountId,
		Name:        req.AccountName,
		Platform:    req.AccountPlatform,
		Type:        req.AccountType,
		Credentials: creds,
		ProxyURL:    req.ProxyUrl,
	}
}

// toProtoResult 将 SDK ForwardResult 转为 proto ForwardResult
func toProtoResult(result *sdk.ForwardResult) *pb.ForwardResult {
	return &pb.ForwardResult{
		StatusCode:    int32(result.StatusCode),
		InputTokens:   int32(result.InputTokens),
		OutputTokens:  int32(result.OutputTokens),
		CacheTokens:   int32(result.CacheTokens),
		Model:         result.Model,
		DurationMs:    result.Duration.Milliseconds(),
		AccountStatus: result.AccountStatus,
		RetryAfterMs:  result.RetryAfter.Milliseconds(),
	}
}

func (s *GatewayGRPCServer) Forward(ctx context.Context, req *pb.ForwardRequest) (*pb.ForwardResult, error) {
	headers := make(http.Header)
	for k, v := range req.Headers {
		headers.Set(k, v)
	}

	fwdReq := &sdk.ForwardRequest{
		Account: buildAccount(req),
		Body:    req.Body,
		Headers: headers,
		Model:   req.Model,
		Stream:  req.Stream,
	}

	result, err := s.Impl.Forward(ctx, fwdReq)
	if err != nil {
		return nil, err
	}
	return toProtoResult(result), nil
}

func (s *GatewayGRPCServer) ForwardStream(req *pb.ForwardRequest, stream pb.GatewayService_ForwardStreamServer) error {
	headers := make(http.Header)
	for k, v := range req.Headers {
		headers.Set(k, v)
	}

	sw := &streamWriter{stream: stream}
	fwdReq := &sdk.ForwardRequest{
		Account: buildAccount(req),
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

	// 补充耗时
	if result.Duration == 0 {
		result.Duration = time.Since(startTime)
	}

	return stream.Send(&pb.ForwardChunk{
		Done:        true,
		FinalResult: toProtoResult(result),
	})
}

func (s *GatewayGRPCServer) ValidateAccount(ctx context.Context, req *pb.CredentialsRequest) (*pb.Empty, error) {
	if err := s.Impl.ValidateAccount(ctx, req.Credentials); err != nil {
		return nil, err
	}
	return &pb.Empty{}, nil
}

func (s *GatewayGRPCServer) QueryQuota(ctx context.Context, req *pb.CredentialsRequest) (*pb.QuotaInfoResponse, error) {
	info, err := s.Impl.QueryQuota(ctx, req.Credentials)
	if err != nil {
		return nil, err
	}
	return &pb.QuotaInfoResponse{
		Total:     info.Total,
		Used:      info.Used,
		Remaining: info.Remaining,
		Currency:  info.Currency,
		ExpiresAt: info.ExpiresAt,
		Extra:     info.Extra,
	}, nil
}

// streamWriter 将 gRPC 流包装为 http.ResponseWriter
type streamWriter struct {
	stream  pb.GatewayService_ForwardStreamServer
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
