package grpc

import (
	"context"

	sdk "github.com/DouDOU-start/airgate-sdk"
	pb "github.com/DouDOU-start/airgate-sdk/proto"
)

// AdvancedGatewayGRPCServer 将 AdvancedGatewayPlugin 包装为 gRPC 服务端
type AdvancedGatewayGRPCServer struct {
	pb.UnimplementedAdvancedGatewayServiceServer
	Impl sdk.AdvancedGatewayPlugin
}

func (s *AdvancedGatewayGRPCServer) GetPlatform(_ context.Context, _ *pb.Empty) (*pb.StringResponse, error) {
	return &pb.StringResponse{Value: s.Impl.Platform()}, nil
}

func (s *AdvancedGatewayGRPCServer) GetModels(_ context.Context, _ *pb.Empty) (*pb.ModelsResponse, error) {
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

func (s *AdvancedGatewayGRPCServer) GetAdvancedServiceNeeds(_ context.Context, _ *pb.Empty) (*pb.AdvancedServiceNeedsResponse, error) {
	needs := s.Impl.AdvancedServices()
	return &pb.AdvancedServiceNeedsResponse{
		NeedScheduler:   needs.NeedScheduler,
		NeedConcurrency: needs.NeedConcurrency,
		NeedRateLimit:   needs.NeedRateLimit,
		NeedBilling:     needs.NeedBilling,
	}, nil
}

func (s *AdvancedGatewayGRPCServer) HandleRequest(ctx context.Context, req *pb.HttpRequest) (*pb.HttpResponse, error) {
	// Advanced 插件的 HTTP 请求处理通过 RegisterRoutes 注册的 handler 来处理
	// 这里是 gRPC 代理模式：核心收到 HTTP 请求 → 序列化为 protobuf → 发给插件 → 反序列化执行 → 返回结果
	// 具体的路由分发由插件内部处理
	return &pb.HttpResponse{
		StatusCode: 501,
		Body:       []byte(`{"error":"not implemented"}`),
	}, nil
}

func (s *AdvancedGatewayGRPCServer) HandleStreamRequest(req *pb.HttpRequest, stream pb.AdvancedGatewayService_HandleStreamRequestServer) error {
	// 流式 HTTP 请求处理（同上，gRPC 代理模式）
	return stream.Send(&pb.HttpResponseChunk{
		Done:       true,
		StatusCode: 501,
		Data:       []byte(`{"error":"not implemented"}`),
	})
}
