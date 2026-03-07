package grpc

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	sdk "github.com/DouDOU-start/airgate-sdk"
	pb "github.com/DouDOU-start/airgate-sdk/proto"
)

// SimpleGatewayGRPCClient 将 gRPC 客户端包装为 SimpleGatewayPlugin 接口（核心侧使用）
type SimpleGatewayGRPCClient struct {
	plugin  pb.PluginServiceClient
	gateway pb.SimpleGatewayServiceClient

	// 缓存
	cachedInfo     *sdk.PluginInfo
	cachedPlatform string
	cachedModels   []sdk.ModelInfo
	cachedRoutes   []sdk.RouteDefinition
}

func (c *SimpleGatewayGRPCClient) Info() sdk.PluginInfo {
	if c.cachedInfo != nil {
		return *c.cachedInfo
	}
	pc := &PluginGRPCClient{client: c.plugin}
	info := pc.Info()
	c.cachedInfo = &info
	return info
}

func (c *SimpleGatewayGRPCClient) Init(ctx sdk.PluginContext) error {
	pc := &PluginGRPCClient{client: c.plugin}
	return pc.Init(ctx)
}

func (c *SimpleGatewayGRPCClient) Start(ctx context.Context) error {
	_, err := c.plugin.Start(ctx, &pb.Empty{})
	return err
}

func (c *SimpleGatewayGRPCClient) Stop(ctx context.Context) error {
	_, err := c.plugin.Stop(ctx, &pb.Empty{})
	return err
}

func (c *SimpleGatewayGRPCClient) Platform() string {
	if c.cachedPlatform != "" {
		return c.cachedPlatform
	}
	resp, err := c.gateway.GetPlatform(context.Background(), &pb.Empty{})
	if err != nil {
		return ""
	}
	c.cachedPlatform = resp.Value
	return resp.Value
}

func (c *SimpleGatewayGRPCClient) Models() []sdk.ModelInfo {
	if c.cachedModels != nil {
		return c.cachedModels
	}
	resp, err := c.gateway.GetModels(context.Background(), &pb.Empty{})
	if err != nil {
		return nil
	}
	models := make([]sdk.ModelInfo, len(resp.Models))
	for i, m := range resp.Models {
		models[i] = sdk.ModelInfo{
			ID:          m.Id,
			Name:        m.Name,
			MaxTokens:   int(m.MaxTokens),
			InputPrice:  m.InputPrice,
			OutputPrice: m.OutputPrice,
			CachePrice:  m.CachePrice,
		}
	}
	c.cachedModels = models
	return models
}

func (c *SimpleGatewayGRPCClient) Routes() []sdk.RouteDefinition {
	if c.cachedRoutes != nil {
		return c.cachedRoutes
	}
	resp, err := c.gateway.GetRoutes(context.Background(), &pb.Empty{})
	if err != nil {
		return nil
	}
	routes := make([]sdk.RouteDefinition, len(resp.Routes))
	for i, r := range resp.Routes {
		routes[i] = sdk.RouteDefinition{
			Method:      r.Method,
			Path:        r.Path,
			Description: r.Description,
		}
	}
	c.cachedRoutes = routes
	return routes
}

func (c *SimpleGatewayGRPCClient) Forward(ctx context.Context, req *sdk.ForwardRequest) (*sdk.ForwardResult, error) {
	// 序列化 credentials
	credsJSON, _ := json.Marshal(req.Account.Credentials)

	// 扁平化 headers
	headers := make(map[string]string)
	for k := range req.Headers {
		headers[k] = req.Headers.Get(k)
	}

	pbReq := &pb.ForwardRequest{
		AccountId:       req.Account.ID,
		CredentialsJson: credsJSON,
		ProxyUrl:        req.Account.ProxyURL,
		Body:            req.Body,
		Headers:         headers,
		Model:           req.Model,
		Stream:          req.Stream,
		RateMultiplier:  req.Account.RateMultiplier,
		MaxConcurrency:  int32(req.Account.MaxConcurrency),
	}

	// 流式请求
	if req.Stream && req.Writer != nil {
		return c.forwardStream(ctx, pbReq, req)
	}

	// 非流式请求
	resp, err := c.gateway.Forward(ctx, pbReq)
	if err != nil {
		return nil, fmt.Errorf("gRPC Forward 调用失败: %w", err)
	}

	return &sdk.ForwardResult{
		StatusCode:   int(resp.StatusCode),
		InputTokens:  int(resp.InputTokens),
		OutputTokens: int(resp.OutputTokens),
		CacheTokens:  int(resp.CacheTokens),
		Model:        resp.Model,
		Duration:     time.Duration(resp.DurationMs) * time.Millisecond,
	}, nil
}

func (c *SimpleGatewayGRPCClient) forwardStream(ctx context.Context, pbReq *pb.ForwardRequest, req *sdk.ForwardRequest) (*sdk.ForwardResult, error) {
	stream, err := c.gateway.ForwardStream(ctx, pbReq)
	if err != nil {
		return nil, fmt.Errorf("gRPC ForwardStream 调用失败: %w", err)
	}

	var finalResult *sdk.ForwardResult
	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("gRPC 流接收失败: %w", err)
		}

		// 写入数据到 ResponseWriter
		if len(chunk.Data) > 0 && req.Writer != nil {
			if _, err := req.Writer.Write(chunk.Data); err != nil {
				return nil, fmt.Errorf("写入响应失败: %w", err)
			}
			// 刷新流式数据
			if flusher, ok := req.Writer.(interface{ Flush() }); ok {
				flusher.Flush()
			}
		}

		// 最终结果
		if chunk.Done && chunk.FinalResult != nil {
			r := chunk.FinalResult
			finalResult = &sdk.ForwardResult{
				StatusCode:   int(r.StatusCode),
				InputTokens:  int(r.InputTokens),
				OutputTokens: int(r.OutputTokens),
				CacheTokens:  int(r.CacheTokens),
				Model:        r.Model,
				Duration:     time.Duration(r.DurationMs) * time.Millisecond,
			}
		}
	}

	if finalResult == nil {
		return nil, fmt.Errorf("未收到最终结果")
	}
	return finalResult, nil
}

// ValidateCredentials 验证凭证（实现 AccountValidator 接口）
func (c *SimpleGatewayGRPCClient) ValidateCredentials(ctx context.Context, credentials map[string]string) error {
	_, err := c.gateway.ValidateCredentials(ctx, &pb.CredentialsRequest{
		Credentials: credentials,
	})
	return err
}

// GetWebAssets 获取插件的前端静态资源
func (c *SimpleGatewayGRPCClient) GetWebAssets() (map[string][]byte, error) {
	resp, err := c.plugin.GetWebAssets(context.Background(), &pb.Empty{})
	if err != nil {
		return nil, err
	}
	if !resp.HasAssets {
		return nil, nil
	}
	assets := make(map[string][]byte, len(resp.Files))
	for _, f := range resp.Files {
		assets[f.Path] = f.Content
	}
	return assets, nil
}
