package grpc

import (
	"context"
	"database/sql"
	"time"

	sdk "github.com/DouDOU-start/airgate-sdk"
	pb "github.com/DouDOU-start/airgate-sdk/proto"
)

// ExtensionGRPCClient 将 gRPC 客户端包装为 ExtensionPlugin 接口（核心侧使用）
type ExtensionGRPCClient struct {
	plugin    pb.PluginServiceClient
	extension pb.ExtensionServiceClient

	cachedInfo *sdk.PluginInfo
}

func (c *ExtensionGRPCClient) Info() sdk.PluginInfo {
	if c.cachedInfo != nil {
		return *c.cachedInfo
	}
	pc := &PluginGRPCClient{client: c.plugin}
	info := pc.Info()
	c.cachedInfo = &info
	return info
}

func (c *ExtensionGRPCClient) Init(ctx sdk.PluginContext) error {
	pc := &PluginGRPCClient{client: c.plugin}
	return pc.Init(ctx)
}

func (c *ExtensionGRPCClient) Start(ctx context.Context) error {
	_, err := c.plugin.Start(ctx, &pb.Empty{})
	return err
}

func (c *ExtensionGRPCClient) Stop(ctx context.Context) error {
	_, err := c.plugin.Stop(ctx, &pb.Empty{})
	return err
}

func (c *ExtensionGRPCClient) RegisterRoutes(_ sdk.RouteRegistrar) {
	// Extension 插件的路由在 gRPC 模式下由核心代理
}

func (c *ExtensionGRPCClient) Migrate(_ *sql.DB) error {
	_, err := c.extension.Migrate(context.Background(), &pb.Empty{})
	return err
}

func (c *ExtensionGRPCClient) BackgroundTasks() []sdk.BackgroundTask {
	resp, err := c.extension.GetBackgroundTasks(context.Background(), &pb.Empty{})
	if err != nil {
		return nil
	}

	tasks := make([]sdk.BackgroundTask, len(resp.Tasks))
	for i, t := range resp.Tasks {
		tasks[i] = sdk.BackgroundTask{
			Name:     t.Name,
			Interval: time.Duration(t.IntervalMs) * time.Millisecond,
			// Handler 在 gRPC 模式下无法直接传递函数
			// 核心通过 HandleRequest gRPC 调用触发后台任务
		}
	}
	return tasks
}

// HandleHTTPRequest 代理 HTTP 请求到插件（核心内部调用）
func (c *ExtensionGRPCClient) HandleHTTPRequest(ctx context.Context, req *pb.HttpRequest) (*pb.HttpResponse, error) {
	return c.extension.HandleRequest(ctx, req)
}

// GetWebAssets 获取插件的前端静态资源
func (c *ExtensionGRPCClient) GetWebAssets() (map[string][]byte, error) {
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
