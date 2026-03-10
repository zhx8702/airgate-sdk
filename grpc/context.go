package grpc

import (
	"log/slog"
	"strconv"
	"time"

	sdk "github.com/DouDOU-start/airgate-sdk"
)

// grpcPluginContext 通过 gRPC 传入的插件上下文（插件进程侧）
type grpcPluginContext struct {
	config sdk.PluginConfig
	logger *slog.Logger
}

func (c *grpcPluginContext) Logger() *slog.Logger {
	if c.logger == nil {
		return slog.Default()
	}
	return c.logger
}

func (c *grpcPluginContext) Config() sdk.PluginConfig {
	return c.config
}

// mapConfig 基于 map 的简易配置实现
type mapConfig struct {
	data map[string]string
}

func (c *mapConfig) GetString(key string) string {
	return c.data[key]
}

func (c *mapConfig) GetInt(key string) int {
	v, _ := strconv.Atoi(c.data[key])
	return v
}

func (c *mapConfig) GetBool(key string) bool {
	v, _ := strconv.ParseBool(c.data[key])
	return v
}

func (c *mapConfig) GetFloat64(key string) float64 {
	v, _ := strconv.ParseFloat(c.data[key], 64)
	return v
}

func (c *mapConfig) GetDuration(key string) time.Duration {
	v, _ := time.ParseDuration(c.data[key])
	return v
}

func (c *mapConfig) GetAll() map[string]string {
	return c.data
}
