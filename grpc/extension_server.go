package grpc

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"

	sdk "github.com/DouDOU-start/airgate-sdk"
	pb "github.com/DouDOU-start/airgate-sdk/proto"
)

// ExtensionGRPCServer 将 ExtensionPlugin 包装为 gRPC 服务端
type ExtensionGRPCServer struct {
	pb.UnimplementedExtensionServiceServer
	Impl   sdk.ExtensionPlugin
	router *extensionRouter
}

// initRouter 初始化路由器并注册插件路由（由 Serve 调用）
func (s *ExtensionGRPCServer) initRouter() {
	s.router = newExtensionRouter()
	s.Impl.RegisterRoutes(s.router)
}

func (s *ExtensionGRPCServer) Migrate(_ context.Context, _ *pb.Empty) (*pb.Empty, error) {
	if err := s.Impl.Migrate(); err != nil {
		return nil, err
	}
	return &pb.Empty{}, nil
}

func (s *ExtensionGRPCServer) GetBackgroundTasks(_ context.Context, _ *pb.Empty) (*pb.BackgroundTasksResponse, error) {
	tasks := s.Impl.BackgroundTasks()
	resp := &pb.BackgroundTasksResponse{}
	for _, t := range tasks {
		resp.Tasks = append(resp.Tasks, &pb.BackgroundTaskProto{
			Name:       t.Name,
			IntervalMs: t.Interval.Milliseconds(),
		})
	}
	return resp, nil
}

func (s *ExtensionGRPCServer) HandleRequest(ctx context.Context, req *pb.HttpRequest) (*pb.HttpResponse, error) {
	if s.router == nil {
		return &pb.HttpResponse{
			StatusCode: 501,
			Body:       []byte(`{"error":"extension router not initialized"}`),
		}, nil
	}

	handler := s.router.match(req.Method, req.Path)
	if handler == nil {
		return &pb.HttpResponse{
			StatusCode: 404,
			Body:       []byte(`{"error":"route not found"}`),
		}, nil
	}

	// 将 gRPC 请求转为 http.Request + httptest.ResponseRecorder
	httpReq, err := pbRequestToHTTP(ctx, req)
	if err != nil {
		return &pb.HttpResponse{
			StatusCode: 500,
			Body:       []byte(`{"error":"failed to convert request"}`),
		}, nil
	}

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httpReq)

	return httpResponseToPB(recorder), nil
}

func (s *ExtensionGRPCServer) HandleStreamRequest(req *pb.HttpRequest, stream pb.ExtensionService_HandleStreamRequestServer) error {
	return stream.Send(&pb.HttpResponseChunk{
		Done:       true,
		StatusCode: 501,
		Data:       []byte(`{"error":"stream not implemented"}`),
	})
}

// pbRequestToHTTP 将 protobuf HttpRequest 转为 net/http.Request
func pbRequestToHTTP(ctx context.Context, req *pb.HttpRequest) (*http.Request, error) {
	url := req.Path
	if req.Query != "" {
		url += "?" + req.Query
	}

	httpReq, err := http.NewRequestWithContext(ctx, req.Method, url, bytes.NewReader(req.Body))
	if err != nil {
		return nil, err
	}

	for k, vals := range req.Headers {
		for _, v := range vals.Values {
			httpReq.Header.Add(k, v)
		}
	}

	httpReq.RemoteAddr = req.RemoteAddr
	return httpReq, nil
}

// httpResponseToPB 将 httptest.ResponseRecorder 转为 protobuf HttpResponse
func httpResponseToPB(rec *httptest.ResponseRecorder) *pb.HttpResponse {
	headers := make(map[string]*pb.HeaderValues)
	for k, v := range rec.Header() {
		headers[strings.ToLower(k)] = &pb.HeaderValues{Values: v}
	}
	return &pb.HttpResponse{
		StatusCode: int32(rec.Code),
		Headers:    headers,
		Body:       rec.Body.Bytes(),
	}
}
