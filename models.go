package sdk

import "net/http"

// Account 上游 AI 账户（核心调度后传给插件）
type Account struct {
	ID             int64             `json:"id"`
	Name           string            `json:"name"`
	Platform       string            `json:"platform"`
	Credentials    map[string]string `json:"credentials"`     // JSONB 透传
	ProxyURL       string            `json:"proxy_url"`
	RateMultiplier float64           `json:"rate_multiplier"`
	MaxConcurrency int               `json:"max_concurrency"`
}

// ModelInfo 模型信息（插件声明，核心缓存用于计费）
type ModelInfo struct {
	ID          string  `json:"id"`           // 如 "claude-opus-4-20250514"
	Name        string  `json:"name"`         // 显示名 "Claude Opus 4"
	MaxTokens   int     `json:"max_tokens"`
	InputPrice  float64 `json:"input_price"`  // 每百万 input token 价格（USD）
	OutputPrice float64 `json:"output_price"` // 每百万 output token 价格（USD）
	CachePrice  float64 `json:"cache_price"`  // 每百万 cache token 价格（USD）
}

// RouteDefinition 路由声明（Simple 网关插件使用）
type RouteDefinition struct {
	Method      string `json:"method"`      // "GET", "POST" 等
	Path        string `json:"path"`        // 如 "/v1/messages"
	Description string `json:"description"`
}

// RouteRegistrar 路由注册器（Advanced/Extension 插件使用）
type RouteRegistrar interface {
	Handle(method, path string, handler http.HandlerFunc)
	Group(prefix string) RouteRegistrar
}

// CredentialField 凭证字段声明（前端据此渲染账号管理表单）
type CredentialField struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	Type        string `json:"type"`        // "text", "password", "textarea", "select"
	Required    bool   `json:"required"`
	Placeholder string `json:"placeholder"`
}

// AccountType 账号类型声明（插件声明多种凭证模式时使用）
// 前端根据此信息渲染账号类型选择器和对应的凭证表单
type AccountType struct {
	Key         string            `json:"key"`         // 类型标识，如 "apikey", "oauth"
	Label       string            `json:"label"`       // 显示名称
	Description string            `json:"description"` // 简要说明
	Fields      []CredentialField `json:"fields"`      // 该类型的凭证字段
}

// ConfigField 配置项声明（前端据此渲染插件配置表单）
type ConfigField struct {
	Key         string `json:"key"`
	Type        string `json:"type"`        // "string", "int", "bool", "duration", "select", "textarea"
	Default     string `json:"default"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
}

// FrontendPage 前端页面声明（extension 插件使用）
type FrontendPage struct {
	Path        string `json:"path"`
	Title       string `json:"title"`
	Icon        string `json:"icon"`
	Description string `json:"description"`
}

// UsageLog 使用记录（插件上报给核心）
type UsageLog struct {
	UserID           int64   `json:"user_id"`
	APIKeyID         int64   `json:"api_key_id"`
	AccountID        int64   `json:"account_id"`
	GroupID          int64   `json:"group_id"`
	Platform         string  `json:"platform"`
	Model            string  `json:"model"`
	InputTokens      int     `json:"input_tokens"`
	OutputTokens     int     `json:"output_tokens"`
	CacheTokens      int     `json:"cache_tokens"`
	InputCost        float64 `json:"input_cost"`
	OutputCost       float64 `json:"output_cost"`
	CacheCost        float64 `json:"cache_cost"`
	TotalCost        float64 `json:"total_cost"`
	ActualCost       float64 `json:"actual_cost"`
	Stream           bool    `json:"stream"`
	DurationMs       int64   `json:"duration_ms"`
	FirstTokenMs     int64   `json:"first_token_ms"`
	UserAgent        string  `json:"user_agent"`
	IPAddress        string  `json:"ip_address"`
}

// ScheduleRequest 调度请求
type ScheduleRequest struct {
	Platform  string `json:"platform"`
	Model     string `json:"model"`
	UserID    int64  `json:"user_id"`
	GroupID   int64  `json:"group_id"`
	SessionID string `json:"session_id"` // 粘性会话 ID
}

// AccountSelection 负载感知调度结果
type AccountSelection struct {
	Account  *Account `json:"account"`
	LoadRate float64  `json:"load_rate"` // 当前负载率
}
