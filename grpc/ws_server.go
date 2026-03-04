package grpc

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	sdk "github.com/DouDOU-start/airgate-sdk"
	pb "github.com/DouDOU-start/airgate-sdk/proto"
)

// HandleWebSocket 处理核心发来的 WebSocket 双向流
// 将 gRPC 双向流包装为 sdk.WebSocketConn，传给插件的 HandleWebSocket()
func (s *SimpleGatewayGRPCServer) HandleWebSocket(stream pb.SimpleGatewayService_HandleWebSocketServer) error {
	// 检查插件是否实现了 WebSocketHandler
	handler, ok := s.Impl.(sdk.WebSocketHandler)
	if !ok {
		return fmt.Errorf("插件未实现 WebSocketHandler 接口")
	}

	// 等待第一帧：CONNECT 帧，获取连接元信息
	firstFrame, err := stream.Recv()
	if err != nil {
		return fmt.Errorf("接收 CONNECT 帧失败: %w", err)
	}
	if firstFrame.Type != pb.WebSocketFrame_CONNECT {
		return fmt.Errorf("期望 CONNECT 帧，收到 %v", firstFrame.Type)
	}

	// 构建 WebSocketConn 适配器
	conn := &grpcWebSocketConn{
		stream:      stream,
		connectInfo: convertConnectInfo(firstFrame.ConnectInfo),
	}

	return handler.HandleWebSocket(stream.Context(), conn)
}

// grpcWebSocketConn 将 gRPC 双向流包装为 sdk.WebSocketConn
type grpcWebSocketConn struct {
	stream      pb.SimpleGatewayService_HandleWebSocketServer
	connectInfo *sdk.WebSocketConnectInfo
}

func (c *grpcWebSocketConn) ReadMessage() (int, []byte, error) {
	frame, err := c.stream.Recv()
	if err != nil {
		if err == io.EOF {
			return 0, nil, io.EOF
		}
		return 0, nil, err
	}

	switch frame.Type {
	case pb.WebSocketFrame_TEXT:
		return sdk.WSMessageText, frame.Data, nil
	case pb.WebSocketFrame_BINARY:
		return sdk.WSMessageBinary, frame.Data, nil
	case pb.WebSocketFrame_CLOSE:
		return 0, nil, io.EOF
	default:
		return sdk.WSMessageText, frame.Data, nil
	}
}

func (c *grpcWebSocketConn) WriteMessage(msgType int, data []byte) error {
	frameType := pb.WebSocketFrame_TEXT
	if msgType == sdk.WSMessageBinary {
		frameType = pb.WebSocketFrame_BINARY
	}
	return c.stream.Send(&pb.WebSocketFrame{
		Type: frameType,
		Data: data,
	})
}

func (c *grpcWebSocketConn) ConnectInfo() *sdk.WebSocketConnectInfo {
	return c.connectInfo
}

func (c *grpcWebSocketConn) Close(code int, reason string) error {
	return c.stream.Send(&pb.WebSocketFrame{
		Type:        pb.WebSocketFrame_CLOSE,
		CloseCode:   int32(code),
		CloseReason: reason,
	})
}

// convertConnectInfo 将 protobuf 连接信息转为 SDK 类型
func convertConnectInfo(info *pb.WebSocketConnectInfo) *sdk.WebSocketConnectInfo {
	if info == nil {
		return &sdk.WebSocketConnectInfo{}
	}

	headers := make(http.Header)
	for k, v := range info.Headers {
		headers.Set(k, v)
	}

	var creds map[string]string
	if len(info.CredentialsJson) > 0 {
		_ = json.Unmarshal(info.CredentialsJson, &creds)
	}

	return &sdk.WebSocketConnectInfo{
		Path:         info.Path,
		Query:        info.Query,
		Headers:      headers,
		RemoteAddr:   info.RemoteAddr,
		ConnectionID: info.ConnectionId,
		Account: &sdk.Account{
			ID:             info.AccountId,
			Credentials:    creds,
			ProxyURL:       info.ProxyUrl,
			RateMultiplier: info.RateMultiplier,
			MaxConcurrency: int(info.MaxConcurrency),
		},
	}
}
