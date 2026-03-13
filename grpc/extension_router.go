package grpc

import (
	"net/http"
	"strings"
	"sync"

	sdk "github.com/DouDOU-start/airgate-sdk"
)

// extensionRouter 实现 sdk.RouteRegistrar，用于在 gRPC 模式下将 HTTP 请求分发到注册的处理函数
type extensionRouter struct {
	mu       sync.RWMutex
	handlers map[string]http.HandlerFunc // "METHOD /path" → handler
	prefix   string
}

func newExtensionRouter() *extensionRouter {
	return &extensionRouter{
		handlers: make(map[string]http.HandlerFunc),
	}
}

func (r *extensionRouter) Handle(method, path string, handler http.HandlerFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()
	fullPath := r.prefix + path
	key := strings.ToUpper(method) + " " + fullPath
	r.handlers[key] = handler
}

func (r *extensionRouter) Group(prefix string) sdk.RouteRegistrar {
	return &extensionRouter{
		handlers: r.handlers,
		prefix:   r.prefix + prefix,
	}
}

// match 查找匹配的处理函数
func (r *extensionRouter) match(method, path string) http.HandlerFunc {
	r.mu.RLock()
	defer r.mu.RUnlock()
	key := strings.ToUpper(method) + " " + path
	return r.handlers[key]
}
