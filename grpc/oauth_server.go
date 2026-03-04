package grpc

import (
	"context"
	"fmt"

	sdk "github.com/DouDOU-start/airgate-sdk"
	pb "github.com/DouDOU-start/airgate-sdk/proto"
)

// StartOAuth 处理核心发来的 OAuth 授权请求
func (s *SimpleGatewayGRPCServer) StartOAuth(ctx context.Context, req *pb.StartOAuthRequest) (*pb.StartOAuthResponse, error) {
	handler, ok := s.Impl.(sdk.OAuthHandler)
	if !ok {
		return nil, fmt.Errorf("插件未实现 OAuthHandler 接口")
	}

	result, err := handler.StartOAuth(ctx, &sdk.OAuthStartRequest{
		CallbackURL: req.CallbackUrl,
		ProxyURL:    req.ProxyUrl,
	})
	if err != nil {
		return nil, err
	}

	return &pb.StartOAuthResponse{
		AuthorizeUrl: result.AuthorizeURL,
		State:        result.State,
	}, nil
}

// HandleOAuthCallback 处理核心发来的 OAuth 回调请求
func (s *SimpleGatewayGRPCServer) HandleOAuthCallback(ctx context.Context, req *pb.OAuthCallbackRequest) (*pb.OAuthCallbackResponse, error) {
	handler, ok := s.Impl.(sdk.OAuthHandler)
	if !ok {
		return nil, fmt.Errorf("插件未实现 OAuthHandler 接口")
	}

	result, err := handler.HandleOAuthCallback(ctx, &sdk.OAuthCallbackRequest{
		Code:     req.Code,
		State:    req.State,
		ProxyURL: req.ProxyUrl,
	})
	if err != nil {
		return nil, err
	}

	return &pb.OAuthCallbackResponse{
		AccountType: result.AccountType,
		Credentials: result.Credentials,
		AccountName: result.AccountName,
	}, nil
}
