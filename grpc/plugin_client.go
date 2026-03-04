package grpc

import (
	"context"

	sdk "github.com/DouDOU-start/airgate-sdk"
	pb "github.com/DouDOU-start/airgate-sdk/proto"
)

// PluginGRPCClient 将 gRPC 客户端包装为 sdk.Plugin 接口（核心侧使用）
type PluginGRPCClient struct {
	client pb.PluginServiceClient
}

func (c *PluginGRPCClient) Info() sdk.PluginInfo {
	resp, err := c.client.GetInfo(context.Background(), &pb.Empty{})
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

	for _, f := range resp.ConfigFields {
		info.ConfigFields = append(info.ConfigFields, sdk.ConfigField{
			Key:         f.Key,
			Type:        f.Type,
			Default:     f.Default,
			Description: f.Description,
			Required:    f.Required,
		})
	}
	for _, f := range resp.CredentialFields {
		info.CredentialFields = append(info.CredentialFields, sdk.CredentialField{
			Key:         f.Key,
			Label:       f.Label,
			Type:        f.Type,
			Required:    f.Required,
			Placeholder: f.Placeholder,
		})
	}
	for _, p := range resp.FrontendPages {
		info.FrontendPages = append(info.FrontendPages, sdk.FrontendPage{
			Path:        p.Path,
			Title:       p.Title,
			Icon:        p.Icon,
			Description: p.Description,
		})
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

	return info
}

func (c *PluginGRPCClient) Init(ctx sdk.PluginContext) error {
	config := make(map[string]string)
	if ctx != nil && ctx.Config() != nil {
		config = ctx.Config().GetAll()
	}
	_, err := c.client.Init(context.Background(), &pb.InitRequest{
		Config: config,
	})
	return err
}

func (c *PluginGRPCClient) Start(ctx context.Context) error {
	_, err := c.client.Start(ctx, &pb.Empty{})
	return err
}

func (c *PluginGRPCClient) Stop(ctx context.Context) error {
	_, err := c.client.Stop(ctx, &pb.Empty{})
	return err
}
