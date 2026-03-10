package grpc

import (
	"context"

	sdk "github.com/DouDOU-start/airgate-sdk"
	pb "github.com/DouDOU-start/airgate-sdk/proto"
)

// ExtensionGRPCServer 将 ExtensionPlugin 包装为 gRPC 服务端
type ExtensionGRPCServer struct {
	pb.UnimplementedExtensionServiceServer
	Impl sdk.ExtensionPlugin
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
	// Extension 的 HTTP 请求通过 gRPC 代理处理
	return &pb.HttpResponse{
		StatusCode: 501,
		Body:       []byte(`{"error":"not implemented"}`),
	}, nil
}

func (s *ExtensionGRPCServer) HandleStreamRequest(req *pb.HttpRequest, stream pb.ExtensionService_HandleStreamRequestServer) error {
	return stream.Send(&pb.HttpResponseChunk{
		Done:       true,
		StatusCode: 501,
		Data:       []byte(`{"error":"not implemented"}`),
	})
}
