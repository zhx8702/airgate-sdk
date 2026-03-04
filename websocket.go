package sdk

import (
	"context"
	"net/http"
)

// WebSocket 消息类型
const (
	WSMessageText   = 1
	WSMessageBinary = 2
)

// WebSocketHandler 可选接口，Simple 网关插件额外实现以支持 WebSocket
// 核心通过 type assertion 自动检测插件是否具备此能力
type WebSocketHandler interface {
	HandleWebSocket(ctx context.Context, conn WebSocketConn) error
}

// WebSocketConn 抽象的 WebSocket 连接（gRPC 双向流实现）
type WebSocketConn interface {
	// ReadMessage 读取客户端发来的消息
	ReadMessage() (messageType int, data []byte, err error)
	// WriteMessage 向客户端发送消息
	WriteMessage(messageType int, data []byte) error
	// ConnectInfo 返回连接建立时的元信息
	ConnectInfo() *WebSocketConnectInfo
	// Close 关闭连接
	Close(code int, reason string) error
}

// WebSocketConnectInfo WebSocket 连接元信息
type WebSocketConnectInfo struct {
	Path         string      // 请求路径
	Query        string      // 查询参数
	Headers      http.Header // 原始请求头
	RemoteAddr   string      // 客户端地址
	ConnectionID string      // 核心分配的连接 ID
	Account      *Account    // 核心调度的账户
}
