package sdk

import "time"

// PluginConfig 插件配置读取接口
type PluginConfig interface {
	// GetString 获取字符串配置项
	GetString(key string) string
	// GetInt 获取整数配置项
	GetInt(key string) int
	// GetBool 获取布尔配置项
	GetBool(key string) bool
	// GetFloat64 获取浮点数配置项
	GetFloat64(key string) float64
	// GetDuration 获取时间间隔配置项
	GetDuration(key string) time.Duration
	// GetAll 获取所有配置（JSONB 原始 map）
	GetAll() map[string]string
}
