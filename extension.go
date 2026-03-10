package sdk

import (
	"context"
	"time"
)

// ExtensionPlugin 通用扩展插件接口
// 可注册自定义路由、数据库迁移、后台任务
type ExtensionPlugin interface {
	Plugin
	// RegisterRoutes 注册扩展路由（管理 API）
	RegisterRoutes(r RouteRegistrar)
	// Migrate 数据库迁移（插件通过 Config 获取 DSN 自行建连）
	Migrate() error
	// BackgroundTasks 声明后台任务（Core 负责调度）
	BackgroundTasks() []BackgroundTask
}

// BackgroundTask 后台任务声明
type BackgroundTask struct {
	Name     string                          // 任务名
	Interval time.Duration                   // 执行间隔
	Handler  func(ctx context.Context) error // 任务处理函数
}
