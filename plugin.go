package sdk

import (
	"context"
	"errors"
	"log/slog"
)

// ErrNotSupported 插件不支持某项能力时返回此错误
var ErrNotSupported = errors.New("not supported")

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
	PluginTypeExtension PluginType = "extension"
)

// PluginInfo 插件元信息
type PluginInfo struct {
	ID              string           `json:"id"`               // 运行时唯一标识，Core 用于 API 路径、资源挂载、缓存键
	Name            string           `json:"name"`             // 展示名称
	Version         string           `json:"version"`          // 语义化版本
	Description     string           `json:"description"`      // 简要描述
	Author          string           `json:"author"`           // 作者
	Type            PluginType       `json:"type"`             // gateway / extension
	AccountTypes    []AccountType    `json:"account_types"`    // 账号类型声明
	FrontendPages   []FrontendPage   `json:"frontend_pages"`   // 前端页面声明
	FrontendWidgets []FrontendWidget `json:"frontend_widgets"` // 前端组件嵌入声明
}

// PluginContext 核心注入给插件的上下文
// 数据库连接通过 Config 传递 DSN（config.GetString("db_dsn")），插件自行建连
type PluginContext interface {
	// Logger 返回结构化日志记录器
	Logger() *slog.Logger
	// Config 返回插件配置
	Config() PluginConfig
}

// ConfigWatcher 可选接口，支持配置热更新的插件实现
type ConfigWatcher interface {
	OnConfigUpdate(config PluginConfig) error
}

// WebAssetsProvider 可选接口，插件实现此接口可提供前端静态资源
// 核心在启动插件时调用，将资源提取到本地供前端动态加载
type WebAssetsProvider interface {
	GetWebAssets() map[string][]byte
}
