package grpc

import (
	"context"

	sdk "github.com/DouDOU-start/airgate-sdk"
	pb "github.com/DouDOU-start/airgate-sdk/proto"
	"github.com/DouDOU-start/airgate-sdk/shared"
	goplugin "github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"
)

// 确保所有 Plugin 类型都实现了 goplugin.GRPCPlugin 接口
var (
	_ goplugin.GRPCPlugin = (*SimpleGatewayGRPCPlugin)(nil)
	_ goplugin.GRPCPlugin = (*AdvancedGatewayGRPCPlugin)(nil)
	_ goplugin.GRPCPlugin = (*PaymentGRPCPlugin)(nil)
	_ goplugin.GRPCPlugin = (*ExtensionGRPCPlugin)(nil)
)

// SimpleGatewayGRPCPlugin 实现 hashicorp/go-plugin.GRPCPlugin 接口
// 插件侧用于注册 gRPC 服务，核心侧用于创建 gRPC 客户端
type SimpleGatewayGRPCPlugin struct {
	goplugin.Plugin
	// Impl 只在插件侧设置（gRPC server 端）
	Impl sdk.SimpleGatewayPlugin
}

func (p *SimpleGatewayGRPCPlugin) GRPCServer(_ *goplugin.GRPCBroker, s *grpc.Server) error {
	pb.RegisterPluginServiceServer(s, &PluginGRPCServer{Impl: p.Impl})
	pb.RegisterSimpleGatewayServiceServer(s, &SimpleGatewayGRPCServer{Impl: p.Impl})
	return nil
}

func (p *SimpleGatewayGRPCPlugin) GRPCClient(_ context.Context, _ *goplugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &SimpleGatewayGRPCClient{
		plugin:  pb.NewPluginServiceClient(c),
		gateway: pb.NewSimpleGatewayServiceClient(c),
	}, nil
}

// AdvancedGatewayGRPCPlugin 实现 Advanced 网关的 go-plugin 接口
type AdvancedGatewayGRPCPlugin struct {
	goplugin.Plugin
	Impl sdk.AdvancedGatewayPlugin
}

func (p *AdvancedGatewayGRPCPlugin) GRPCServer(_ *goplugin.GRPCBroker, s *grpc.Server) error {
	pb.RegisterPluginServiceServer(s, &PluginGRPCServer{Impl: p.Impl})
	pb.RegisterAdvancedGatewayServiceServer(s, &AdvancedGatewayGRPCServer{Impl: p.Impl})
	return nil
}

func (p *AdvancedGatewayGRPCPlugin) GRPCClient(_ context.Context, _ *goplugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &AdvancedGatewayGRPCClient{
		plugin:   pb.NewPluginServiceClient(c),
		advanced: pb.NewAdvancedGatewayServiceClient(c),
	}, nil
}

// PaymentGRPCPlugin 实现支付插件的 go-plugin 接口
type PaymentGRPCPlugin struct {
	goplugin.Plugin
	Impl sdk.PaymentPlugin
}

func (p *PaymentGRPCPlugin) GRPCServer(_ *goplugin.GRPCBroker, s *grpc.Server) error {
	pb.RegisterPluginServiceServer(s, &PluginGRPCServer{Impl: p.Impl})
	pb.RegisterPaymentServiceServer(s, &PaymentGRPCServer{Impl: p.Impl})
	return nil
}

func (p *PaymentGRPCPlugin) GRPCClient(_ context.Context, _ *goplugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &PaymentGRPCClient{
		plugin:  pb.NewPluginServiceClient(c),
		payment: pb.NewPaymentServiceClient(c),
	}, nil
}

// ExtensionGRPCPlugin 实现扩展插件的 go-plugin 接口
type ExtensionGRPCPlugin struct {
	goplugin.Plugin
	Impl sdk.ExtensionPlugin
}

func (p *ExtensionGRPCPlugin) GRPCServer(_ *goplugin.GRPCBroker, s *grpc.Server) error {
	pb.RegisterPluginServiceServer(s, &PluginGRPCServer{Impl: p.Impl})
	pb.RegisterExtensionServiceServer(s, &ExtensionGRPCServer{Impl: p.Impl})
	return nil
}

func (p *ExtensionGRPCPlugin) GRPCClient(_ context.Context, _ *goplugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &ExtensionGRPCClient{
		plugin:    pb.NewPluginServiceClient(c),
		extension: pb.NewExtensionServiceClient(c),
	}, nil
}

// Serve 便捷函数：启动插件 gRPC 服务（插件的 main.go 中调用）
func Serve(impl interface{}) {
	pluginMap := make(goplugin.PluginSet)

	switch p := impl.(type) {
	case sdk.SimpleGatewayPlugin:
		pluginMap[shared.PluginKeySimpleGateway] = &SimpleGatewayGRPCPlugin{Impl: p}
	case sdk.AdvancedGatewayPlugin:
		pluginMap[shared.PluginKeyAdvancedGateway] = &AdvancedGatewayGRPCPlugin{Impl: p}
	case sdk.PaymentPlugin:
		pluginMap[shared.PluginKeyPayment] = &PaymentGRPCPlugin{Impl: p}
	case sdk.ExtensionPlugin:
		pluginMap[shared.PluginKeyExtension] = &ExtensionGRPCPlugin{Impl: p}
	}

	goplugin.Serve(&goplugin.ServeConfig{
		HandshakeConfig: shared.Handshake,
		Plugins:         pluginMap,
		GRPCServer:      goplugin.DefaultGRPCServer,
	})
}

// ServeWithContext 带 context 的便捷函数
func ServeWithContext(_ context.Context, impl interface{}) {
	Serve(impl)
}
