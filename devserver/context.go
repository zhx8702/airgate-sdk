package devserver

import (
	"log/slog"
	"time"

	sdk "github.com/DouDOU-start/airgate-sdk"
)

// devPluginContext 开发模式的 PluginContext 实现
type devPluginContext struct {
	logger *slog.Logger
}

func (c *devPluginContext) Logger() *slog.Logger    { return c.logger }
func (c *devPluginContext) Config() sdk.PluginConfig { return &devPluginConfig{} }

// devPluginConfig 开发模式的空配置（返回零值）
type devPluginConfig struct{}

func (c *devPluginConfig) GetString(key string) string         { return "" }
func (c *devPluginConfig) GetInt(key string) int               { return 0 }
func (c *devPluginConfig) GetBool(key string) bool             { return false }
func (c *devPluginConfig) GetFloat64(key string) float64       { return 0 }
func (c *devPluginConfig) GetDuration(key string) time.Duration { return 0 }
func (c *devPluginConfig) GetAll() map[string]string           { return nil }
