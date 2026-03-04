package grpc

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"

	sdk "github.com/DouDOU-start/airgate-sdk"
	pb "github.com/DouDOU-start/airgate-sdk/proto"
)

// PaymentGRPCServer 将 PaymentPlugin 包装为 gRPC 服务端
type PaymentGRPCServer struct {
	pb.UnimplementedPaymentServiceServer
	Impl sdk.PaymentPlugin
}

func (s *PaymentGRPCServer) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.PaymentOrderResponse, error) {
	order, err := s.Impl.CreateOrder(ctx, sdk.CreateOrderRequest{
		UserID:  req.UserId,
		Amount:  req.Amount,
		Channel: req.Channel,
	})
	if err != nil {
		return nil, err
	}
	return &pb.PaymentOrderResponse{
		OrderId: order.OrderID,
		PayUrl:  order.PayURL,
		Status:  order.Status,
		Amount:  order.Amount,
	}, nil
}

func (s *PaymentGRPCServer) QueryOrder(ctx context.Context, req *pb.QueryOrderRequest) (*pb.PaymentOrderResponse, error) {
	order, err := s.Impl.QueryOrder(ctx, req.OrderId)
	if err != nil {
		return nil, err
	}
	return &pb.PaymentOrderResponse{
		OrderId: order.OrderID,
		PayUrl:  order.PayURL,
		Status:  order.Status,
		Amount:  order.Amount,
	}, nil
}

func (s *PaymentGRPCServer) HandleCallback(ctx context.Context, req *pb.HttpRequest) (*pb.HttpResponse, error) {
	// 将 protobuf HttpRequest 转换为 http.Request
	httpReq := &http.Request{
		Method: req.Method,
		URL:    &url.URL{Path: req.Path, RawQuery: req.Query},
		Header: make(http.Header),
		Body:   io.NopCloser(bytes.NewReader(req.Body)),
	}
	for k, v := range req.Headers {
		httpReq.Header.Set(k, v)
	}
	httpReq = httpReq.WithContext(ctx)

	// 使用 ResponseRecorder 捕获响应
	rec := &responseRecorder{
		headers: make(http.Header),
		code:    200,
	}

	if err := s.Impl.HandleCallback(rec, httpReq); err != nil {
		return nil, err
	}

	respHeaders := make(map[string]string)
	for k := range rec.headers {
		respHeaders[k] = rec.headers.Get(k)
	}

	return &pb.HttpResponse{
		StatusCode: int32(rec.code),
		Headers:    respHeaders,
		Body:       rec.body.Bytes(),
	}, nil
}

// responseRecorder 用于捕获 http.ResponseWriter 的输出
type responseRecorder struct {
	headers http.Header
	body    bytes.Buffer
	code    int
}

func (r *responseRecorder) Header() http.Header {
	return r.headers
}

func (r *responseRecorder) Write(data []byte) (int, error) {
	return r.body.Write(data)
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.code = statusCode
}
