package grpc

import (
	"context"

	sdk "github.com/DouDOU-start/airgate-sdk"
	pb "github.com/DouDOU-start/airgate-sdk/proto"
)

// StartOAuth 核心侧调用：发起 OAuth 授权
func (c *SimpleGatewayGRPCClient) StartOAuth(ctx context.Context, req *sdk.OAuthStartRequest) (*sdk.OAuthStartResponse, error) {
	resp, err := c.gateway.StartOAuth(ctx, &pb.StartOAuthRequest{
		CallbackUrl: req.CallbackURL,
		ProxyUrl:    req.ProxyURL,
	})
	if err != nil {
		return nil, err
	}

	return &sdk.OAuthStartResponse{
		AuthorizeURL: resp.AuthorizeUrl,
		State:        resp.State,
	}, nil
}

// HandleOAuthCallback 核心侧调用：处理 OAuth 回调
func (c *SimpleGatewayGRPCClient) HandleOAuthCallback(ctx context.Context, req *sdk.OAuthCallbackRequest) (*sdk.OAuthResult, error) {
	resp, err := c.gateway.HandleOAuthCallback(ctx, &pb.OAuthCallbackRequest{
		Code:     req.Code,
		State:    req.State,
		ProxyUrl: req.ProxyURL,
	})
	if err != nil {
		return nil, err
	}

	return &sdk.OAuthResult{
		AccountType: resp.AccountType,
		Credentials: resp.Credentials,
		AccountName: resp.AccountName,
	}, nil
}
