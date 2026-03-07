package sdk

import (
	"context"
	"net/http"
	"time"
)

// SimpleGatewayPlugin 简单网关插件
// 核心自动处理：账户调度、计费、限流、并发控制
// 插件只需实现请求转换和响应解析
type SimpleGatewayPlugin interface {
	Plugin
	// Platform 返回平台标识（如 "claude"），与 accounts.platform 对应
	Platform() string
	// Models 返回支持的模型列表（含价格信息，核心用于计费）
	Models() []ModelInfo
	// Routes 声明原生端点（如 POST /v1/messages），核心据此自动注册路由
	Routes() []RouteDefinition
	// Forward 转发请求到上游，账户已由核心调度选好
	Forward(ctx context.Context, req *ForwardRequest) (*ForwardResult, error)
}

// AdvancedGatewayPlugin 高级网关插件
// 插件完全控制调度和转发逻辑，通过 CoreServices 访问核心能力
type AdvancedGatewayPlugin interface {
	Plugin
	// Platform 返回平台标识
	Platform() string
	// Models 返回支持的模型列表
	Models() []ModelInfo
	// RegisterRoutes 自行注册路由和 handler
	RegisterRoutes(r RouteRegistrar)
	// AdvancedServices 声明需要的核心服务
	AdvancedServices() AdvancedServiceNeeds
}

// AdvancedServiceNeeds 声明 Advanced 插件需要的核心服务
type AdvancedServiceNeeds struct {
	NeedScheduler   bool `json:"need_scheduler"`
	NeedConcurrency bool `json:"need_concurrency"`
	NeedRateLimit   bool `json:"need_rate_limit"`
	NeedBilling     bool `json:"need_billing"`
}

// ForwardRequest 转发请求（核心 → 插件）
type ForwardRequest struct {
	Account *Account           // 核心已选好的账户
	Body    []byte             // 原始请求体
	Headers http.Header        // 原始请求头
	Model   string             // 请求的模型
	Stream  bool               // 是否流式
	Writer  http.ResponseWriter // 用于流式写入
}

// ForwardResult 转发结果（插件 → 核心，用于计费和记录）
type ForwardResult struct {
	StatusCode      int           // HTTP 状态码
	InputTokens     int           // 输入 token 数
	OutputTokens    int           // 输出 token 数
	CacheTokens     int           // 缓存 token 数
	Model           string        // 实际使用的模型
	Duration        time.Duration // 请求耗时
	FallbackErrBody []byte        // 降级模式下的上游错误体（供调用方判断是否需要降级）
}
