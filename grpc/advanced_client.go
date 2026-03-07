package grpc

import (
	"context"

	sdk "github.com/DouDOU-start/airgate-sdk"
	pb "github.com/DouDOU-start/airgate-sdk/proto"
)

// AdvancedGatewayGRPCClient 将 gRPC 客户端包装为 AdvancedGatewayPlugin 接口（核心侧使用）
type AdvancedGatewayGRPCClient struct {
	plugin   pb.PluginServiceClient
	advanced pb.AdvancedGatewayServiceClient

	cachedInfo     *sdk.PluginInfo
	cachedPlatform string
	cachedModels   []sdk.ModelInfo
}

func (c *AdvancedGatewayGRPCClient) Info() sdk.PluginInfo {
	if c.cachedInfo != nil {
		return *c.cachedInfo
	}
	pc := &PluginGRPCClient{client: c.plugin}
	info := pc.Info()
	c.cachedInfo = &info
	return info
}

func (c *AdvancedGatewayGRPCClient) Init(ctx sdk.PluginContext) error {
	pc := &PluginGRPCClient{client: c.plugin}
	return pc.Init(ctx)
}

func (c *AdvancedGatewayGRPCClient) Start(ctx context.Context) error {
	_, err := c.plugin.Start(ctx, &pb.Empty{})
	return err
}

func (c *AdvancedGatewayGRPCClient) Stop(ctx context.Context) error {
	_, err := c.plugin.Stop(ctx, &pb.Empty{})
	return err
}

func (c *AdvancedGatewayGRPCClient) Platform() string {
	if c.cachedPlatform != "" {
		return c.cachedPlatform
	}
	resp, err := c.advanced.GetPlatform(context.Background(), &pb.Empty{})
	if err != nil {
		return ""
	}
	c.cachedPlatform = resp.Value
	return resp.Value
}

func (c *AdvancedGatewayGRPCClient) Models() []sdk.ModelInfo {
	if c.cachedModels != nil {
		return c.cachedModels
	}
	resp, err := c.advanced.GetModels(context.Background(), &pb.Empty{})
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

func (c *AdvancedGatewayGRPCClient) RegisterRoutes(_ sdk.RouteRegistrar) {
	// Advanced 插件的路由在 gRPC 模式下由核心代理
	// 核心收到请求后通过 HandleRequest gRPC 调用转发给插件
}

func (c *AdvancedGatewayGRPCClient) AdvancedServices() sdk.AdvancedServiceNeeds {
	resp, err := c.advanced.GetAdvancedServiceNeeds(context.Background(), &pb.Empty{})
	if err != nil {
		return sdk.AdvancedServiceNeeds{}
	}
	return sdk.AdvancedServiceNeeds{
		NeedScheduler:   resp.NeedScheduler,
		NeedConcurrency: resp.NeedConcurrency,
		NeedRateLimit:   resp.NeedRateLimit,
		NeedBilling:     resp.NeedBilling,
	}
}

// HandleHTTPRequest 代理 HTTP 请求到插件（核心内部调用）
func (c *AdvancedGatewayGRPCClient) HandleHTTPRequest(ctx context.Context, req *pb.HttpRequest) (*pb.HttpResponse, error) {
	return c.advanced.HandleRequest(ctx, req)
}

// GetWebAssets 获取插件的前端静态资源
func (c *AdvancedGatewayGRPCClient) GetWebAssets() (map[string][]byte, error) {
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
