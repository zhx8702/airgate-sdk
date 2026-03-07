package sdk

import (
	"context"
	"database/sql"
	"log/slog"
)

// Plugin 基础插件接口，所有插件必须实现
type Plugin interface {
	// Info 返回插件元信息
	Info() PluginInfo
	// Init 初始化插件，核心注入上下文
	Init(ctx PluginContext) error
	// Start 启动插件
	Start(ctx context.Context) error
	// Stop 停止插件
	Stop(ctx context.Context) error
}

// PluginType 插件类型
type PluginType string

const (
	PluginTypeGateway   PluginType = "gateway"
	PluginTypePayment   PluginType = "payment"
	PluginTypeExtension PluginType = "extension"
)

// PluginInfo 插件元信息
type PluginInfo struct {
	ID               string            `json:"id"`                // 唯一标识，如 "gateway-claude"
	Name             string            `json:"name"`              // 显示名称
	Version          string            `json:"version"`           // 语义化版本
	Description      string            `json:"description"`       // 简要描述
	Author           string            `json:"author"`            // 作者
	Type             PluginType        `json:"type"`              // gateway / payment / extension
	ConfigFields     []ConfigField     `json:"config_fields"`     // 插件配置项声明
	CredentialFields []CredentialField `json:"credential_fields"` // 凭证字段声明（向后兼容）
	AccountTypes     []AccountType     `json:"account_types"`     // 账号类型声明（多凭证模式时使用，优先于 CredentialFields）
	FrontendPages    []FrontendPage    `json:"frontend_pages"`    // 前端页面声明（extension 使用）
}

// PluginContext 核心注入给插件的上下文
type PluginContext interface {
	// Logger 返回日志记录器
	Logger() *slog.Logger
	// Config 返回插件配置
	Config() PluginConfig
	// DB 返回数据库连接（extension 插件建表使用）
	DB() *sql.DB
	// CoreServices 返回核心服务集合（Advanced/Extension 插件使用）
	CoreServices() CoreServices
}

// CoreServices 核心服务集合
type CoreServices interface {
	Scheduler() SchedulerService
	Concurrency() ConcurrencyService
	RateLimit() RateLimitService
	Billing() BillingService
}

// ConfigWatcher 可选接口，支持配置热更新的插件实现
type ConfigWatcher interface {
	OnConfigUpdate(config PluginConfig) error
}

// AccountValidator 可选接口，支持凭证验证的网关插件实现
type AccountValidator interface {
	ValidateCredentials(ctx context.Context, credentials map[string]string) error
}

// WebAssetsProvider 可选接口，插件实现此接口可提供前端静态资源
// 核心在启动插件时调用，将资源提取到本地供前端动态加载
type WebAssetsProvider interface {
	GetWebAssets() map[string][]byte
}
