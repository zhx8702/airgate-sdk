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

// PaymentGRPCClient 将 gRPC 客户端包装为 PaymentPlugin 接口（核心侧使用）
type PaymentGRPCClient struct {
	plugin  pb.PluginServiceClient
	payment pb.PaymentServiceClient

	cachedInfo *sdk.PluginInfo
}

func (c *PaymentGRPCClient) Info() sdk.PluginInfo {
	if c.cachedInfo != nil {
		return *c.cachedInfo
	}
	pc := &PluginGRPCClient{client: c.plugin}
	info := pc.Info()
	c.cachedInfo = &info
	return info
}

func (c *PaymentGRPCClient) Init(ctx sdk.PluginContext) error {
	pc := &PluginGRPCClient{client: c.plugin}
	return pc.Init(ctx)
}

func (c *PaymentGRPCClient) Start(ctx context.Context) error {
	_, err := c.plugin.Start(ctx, &pb.Empty{})
	return err
}

func (c *PaymentGRPCClient) Stop(ctx context.Context) error {
	_, err := c.plugin.Stop(ctx, &pb.Empty{})
	return err
}

func (c *PaymentGRPCClient) CreateOrder(ctx context.Context, req sdk.CreateOrderRequest) (*sdk.PaymentOrder, error) {
	resp, err := c.payment.CreateOrder(ctx, &pb.CreateOrderRequest{
		UserId:  req.UserID,
		Amount:  req.Amount,
		Channel: req.Channel,
	})
	if err != nil {
		return nil, err
	}
	return &sdk.PaymentOrder{
		OrderID: resp.OrderId,
		PayURL:  resp.PayUrl,
		Status:  resp.Status,
		Amount:  resp.Amount,
	}, nil
}

func (c *PaymentGRPCClient) QueryOrder(ctx context.Context, orderID string) (*sdk.PaymentOrder, error) {
	resp, err := c.payment.QueryOrder(ctx, &pb.QueryOrderRequest{
		OrderId: orderID,
	})
	if err != nil {
		return nil, err
	}
	return &sdk.PaymentOrder{
		OrderID: resp.OrderId,
		PayURL:  resp.PayUrl,
		Status:  resp.Status,
		Amount:  resp.Amount,
	}, nil
}

func (c *PaymentGRPCClient) HandleCallback(w http.ResponseWriter, r *http.Request) error {
	// 将 http.Request 序列化为 protobuf
	body, _ := io.ReadAll(r.Body)
	headers := make(map[string]string)
	for k := range r.Header {
		headers[k] = r.Header.Get(k)
	}

	query := ""
	if r.URL != nil {
		query = r.URL.RawQuery
	}
	path := ""
	if r.URL != nil {
		path = r.URL.Path
	}

	resp, err := c.payment.HandleCallback(r.Context(), &pb.HttpRequest{
		Method:     r.Method,
		Path:       path,
		Query:      query,
		Headers:    headers,
		Body:       body,
		RemoteAddr: r.RemoteAddr,
	})
	if err != nil {
		return err
	}

	// 将 protobuf 响应写回 http.ResponseWriter
	for k, v := range resp.Headers {
		w.Header().Set(k, v)
	}
	w.WriteHeader(int(resp.StatusCode))
	_, _ = w.Write(resp.Body)
	return nil
}

// httpRequestFromProto 将 protobuf HttpRequest 转为 http.Request（工具函数）
func httpRequestFromProto(req *pb.HttpRequest) *http.Request {
	r := &http.Request{
		Method: req.Method,
		URL:    &url.URL{Path: req.Path, RawQuery: req.Query},
		Header: make(http.Header),
		Body:   io.NopCloser(bytes.NewReader(req.Body)),
	}
	for k, v := range req.Headers {
		r.Header.Set(k, v)
	}
	return r
}
