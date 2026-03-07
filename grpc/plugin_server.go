package grpc

import (
	"context"

	sdk "github.com/DouDOU-start/airgate-sdk"
	pb "github.com/DouDOU-start/airgate-sdk/proto"
)

// PluginGRPCServer 将 sdk.Plugin 实现包装为 gRPC 服务端
type PluginGRPCServer struct {
	pb.UnimplementedPluginServiceServer
	Impl sdk.Plugin
}

func (s *PluginGRPCServer) GetInfo(_ context.Context, _ *pb.Empty) (*pb.PluginInfoResponse, error) {
	info := s.Impl.Info()
	resp := &pb.PluginInfoResponse{
		Id:          info.ID,
		Name:        info.Name,
		Version:     info.Version,
		Description: info.Description,
		Author:      info.Author,
		Type:        string(info.Type),
	}

	for _, f := range info.ConfigFields {
		resp.ConfigFields = append(resp.ConfigFields, &pb.ConfigFieldProto{
			Key:         f.Key,
			Type:        f.Type,
			Default:     f.Default,
			Description: f.Description,
			Required:    f.Required,
		})
	}
	for _, f := range info.CredentialFields {
		resp.CredentialFields = append(resp.CredentialFields, &pb.CredentialFieldProto{
			Key:         f.Key,
			Label:       f.Label,
			Type:        f.Type,
			Required:    f.Required,
			Placeholder: f.Placeholder,
		})
	}
	for _, p := range info.FrontendPages {
		resp.FrontendPages = append(resp.FrontendPages, &pb.FrontendPageProto{
			Path:        p.Path,
			Title:       p.Title,
			Icon:        p.Icon,
			Description: p.Description,
		})
	}
	for _, at := range info.AccountTypes {
		atProto := &pb.AccountTypeProto{
			Key:         at.Key,
			Label:       at.Label,
			Description: at.Description,
		}
		for _, f := range at.Fields {
			atProto.Fields = append(atProto.Fields, &pb.CredentialFieldProto{
				Key:         f.Key,
				Label:       f.Label,
				Type:        f.Type,
				Required:    f.Required,
				Placeholder: f.Placeholder,
			})
		}
		resp.AccountTypes = append(resp.AccountTypes, atProto)
	}

	return resp, nil
}

func (s *PluginGRPCServer) Init(ctx context.Context, req *pb.InitRequest) (*pb.Empty, error) {
	// 构建 PluginContext（简易实现，配置通过 gRPC 传入）
	pctx := &grpcPluginContext{
		config: &mapConfig{data: req.Config},
	}
	if err := s.Impl.Init(pctx); err != nil {
		return nil, err
	}
	return &pb.Empty{}, nil
}

func (s *PluginGRPCServer) Start(ctx context.Context, _ *pb.Empty) (*pb.Empty, error) {
	if err := s.Impl.Start(ctx); err != nil {
		return nil, err
	}
	return &pb.Empty{}, nil
}

func (s *PluginGRPCServer) Stop(ctx context.Context, _ *pb.Empty) (*pb.Empty, error) {
	if err := s.Impl.Stop(ctx); err != nil {
		return nil, err
	}
	return &pb.Empty{}, nil
}

// GetWebAssets 获取插件的前端静态资源
func (s *PluginGRPCServer) GetWebAssets(_ context.Context, _ *pb.Empty) (*pb.WebAssetsResponse, error) {
	provider, ok := s.Impl.(sdk.WebAssetsProvider)
	if !ok {
		return &pb.WebAssetsResponse{HasAssets: false}, nil
	}
	assets := provider.GetWebAssets()
	if len(assets) == 0 {
		return &pb.WebAssetsResponse{HasAssets: false}, nil
	}
	resp := &pb.WebAssetsResponse{HasAssets: true}
	for path, content := range assets {
		resp.Files = append(resp.Files, &pb.WebAssetFile{
			Path:    path,
			Content: content,
		})
	}
	return resp, nil
}
