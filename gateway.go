package sdk

import (
	"context"
	"net/http"
	"time"
)

// GatewayPlugin 网关插件接口
// Core 自动处理：账号调度、计费、限流、并发控制
// 插件负责：声明模型和路由、转发请求、验证凭证、查询额度
type GatewayPlugin interface {
	Plugin

	// Platform 返回业务平台标识（如 "openai"），与 accounts.platform 对应
	Platform() string
	// Models 返回支持的模型列表（含价格信息，Core 用于计费）
	Models() []ModelInfo
	// Routes 声明 API 端点（如 POST /v1/chat/completions），Core 据此自动注册路由
	Routes() []RouteDefinition
	// Forward 转发请求到上游，账号已由 Core 调度选好
	Forward(ctx context.Context, req *ForwardRequest) (*ForwardResult, error)
	// ValidateAccount 验证凭证有效性，Core 在添加/导入账号时调用
	ValidateAccount(ctx context.Context, credentials map[string]string) error
	// QueryQuota 查询账号额度，Core 定时巡检调用（不支持可返回 ErrNotSupported）
	QueryQuota(ctx context.Context, credentials map[string]string) (*QuotaInfo, error)
	// HandleWebSocket 处理 WebSocket 双向通信，连接结束后返回 ForwardResult（不支持可返回 ErrNotSupported）
	HandleWebSocket(ctx context.Context, conn WebSocketConn) (*ForwardResult, error)
}

// ForwardRequest 转发请求（Core → 插件）
type ForwardRequest struct {
	Account *Account            // Core 已调度好的上游账号
	Body    []byte              // 原始请求体
	Headers http.Header         // 原始请求头
	Model   string              // 请求的模型
	Stream  bool                // 是否流式
	Writer  http.ResponseWriter // 用于流式写入 SSE 响应
}

// ForwardResult 转发结果（插件 → Core，用于计费和账号状态管理）
type ForwardResult struct {
	StatusCode   int           // HTTP 状态码
	InputTokens  int           // 输入 token 数
	OutputTokens int           // 输出 token 数
	CacheTokens  int           // 缓存 token 数
	Model        string        // 实际使用的模型
	Duration     time.Duration // 请求耗时

	// 账号状态反馈（插件识别，Core 处置）
	AccountStatus string        // "" 正常 / "rate_limited" / "disabled" / "expired"
	RetryAfter    time.Duration // 限流时建议的等待时间
}
