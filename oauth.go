package sdk

import "context"

// OAuthHandler 可选接口，插件实现后支持 OAuth 浏览器授权
// 核心/devserver 通过类型断言检测：handler, ok := impl.(sdk.OAuthHandler)
type OAuthHandler interface {
	// StartOAuth 发起授权，返回浏览器应跳转的授权 URL
	// callbackURL 由核心/devserver 提供（如 http://localhost:8080/auth/callback）
	StartOAuth(ctx context.Context, req *OAuthStartRequest) (*OAuthStartResponse, error)

	// HandleOAuthCallback 处理授权回调，完成 token 交换，返回凭证
	HandleOAuthCallback(ctx context.Context, req *OAuthCallbackRequest) (*OAuthResult, error)
}

// OAuthStartRequest 发起 OAuth 授权请求
type OAuthStartRequest struct {
	CallbackURL string // 核心/devserver 提供的回调地址
	ProxyURL    string // HTTP 代理（可选）
}

// OAuthStartResponse 授权发起响应
type OAuthStartResponse struct {
	AuthorizeURL string // 浏览器应打开的授权 URL
	State        string // CSRF state（插件生成，核心暂存用于回调校验）
}

// OAuthCallbackRequest OAuth 回调请求
type OAuthCallbackRequest struct {
	Code     string // 授权码
	State    string // 回调中的 state
	ProxyURL string // HTTP 代理（可选）
}

// OAuthResult OAuth 授权结果
type OAuthResult struct {
	AccountType string            // 账号类型标识（如 "oauth"）
	Credentials map[string]string // 凭证（access_token, refresh_token 等）
	AccountName string            // 建议的账号名称
}
