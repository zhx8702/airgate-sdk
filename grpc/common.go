package grpc

import (
	"context"
	"time"

	sdk "github.com/DouDOU-start/airgate-sdk"
	pb "github.com/DouDOU-start/airgate-sdk/proto"
)

// defaultGRPCTimeout gRPC 内部调用的默认超时时间
const defaultGRPCTimeout = 10 * time.Second

// withTimeout 创建带默认超时的 context
func withTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), defaultGRPCTimeout)
}

// pluginBase 封装所有 gRPC Client 共有的 Plugin 接口方法，
// 通过嵌入到各具体 Client 中消除重复代码
type pluginBase struct {
	plugin     pb.PluginServiceClient
	cachedInfo *sdk.PluginInfo
}

// Info 获取插件信息（带缓存）
func (b *pluginBase) Info() sdk.PluginInfo {
	if b.cachedInfo != nil {
		return *b.cachedInfo
	}
	ctx, cancel := withTimeout()
	defer cancel()
	resp, err := b.plugin.GetInfo(ctx, &pb.Empty{})
	if err != nil {
		return sdk.PluginInfo{}
	}

	info := sdk.PluginInfo{
		ID:          resp.Id,
		Name:        resp.Name,
		Version:     resp.Version,
		Description: resp.Description,
		Author:      resp.Author,
		Type:        sdk.PluginType(resp.Type),
	}

	for _, at := range resp.AccountTypes {
		accountType := sdk.AccountType{
			Key:         at.Key,
			Label:       at.Label,
			Description: at.Description,
		}
		for _, f := range at.Fields {
			accountType.Fields = append(accountType.Fields, sdk.CredentialField{
				Key:         f.Key,
				Label:       f.Label,
				Type:        f.Type,
				Required:    f.Required,
				Placeholder: f.Placeholder,
			})
		}
		info.AccountTypes = append(info.AccountTypes, accountType)
	}
	for _, p := range resp.FrontendPages {
		info.FrontendPages = append(info.FrontendPages, sdk.FrontendPage{
			Path:        p.Path,
			Title:       p.Title,
			Icon:        p.Icon,
			Description: p.Description,
		})
	}
	for _, w := range resp.FrontendWidgets {
		info.FrontendWidgets = append(info.FrontendWidgets, sdk.FrontendWidget{
			Slot:      w.Slot,
			EntryFile: w.EntryFile,
			Title:     w.Title,
		})
	}

	b.cachedInfo = &info
	return info
}

// Init 初始化插件
func (b *pluginBase) Init(ctx sdk.PluginContext) error {
	config := make(map[string]string)
	if ctx != nil && ctx.Config() != nil {
		config = ctx.Config().GetAll()
	}
	grpcCtx, cancel := withTimeout()
	defer cancel()
	_, err := b.plugin.Init(grpcCtx, &pb.InitRequest{
		Config: config,
	})
	return err
}

// Start 启动插件
func (b *pluginBase) Start(ctx context.Context) error {
	_, err := b.plugin.Start(ctx, &pb.Empty{})
	return err
}

// Stop 停止插件
func (b *pluginBase) Stop(ctx context.Context) error {
	_, err := b.plugin.Stop(ctx, &pb.Empty{})
	return err
}

// GetWebAssets 获取插件前端静态资源
func (b *pluginBase) GetWebAssets() (map[string][]byte, error) {
	ctx, cancel := withTimeout()
	defer cancel()
	resp, err := b.plugin.GetWebAssets(ctx, &pb.Empty{})
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

// convertModels 将 proto ModelInfoProto 列表转为 SDK ModelInfo 列表
func convertModels(pbModels []*pb.ModelInfoProto) []sdk.ModelInfo {
	models := make([]sdk.ModelInfo, len(pbModels))
	for i, m := range pbModels {
		models[i] = sdk.ModelInfo{
			ID:          m.Id,
			Name:        m.Name,
			MaxTokens:   int(m.MaxTokens),
			InputPrice:  m.InputPrice,
			OutputPrice: m.OutputPrice,
			CachePrice:  m.CachePrice,
		}
	}
	return models
}
