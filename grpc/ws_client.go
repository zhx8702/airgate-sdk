package grpc

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	sdk "github.com/DouDOU-start/airgate-sdk"
	pb "github.com/DouDOU-start/airgate-sdk/proto"
)

// HandleWebSocket 核心侧调用：建立 gRPC 双向流，桥接入站 WebSocket 连接到插件
// wsConn 是核心侧已升级的 WebSocket 连接（由调用方自行管理）
// 返回值通过 readCh/writeCh 实现消息传递，调用方负责与实际 WS 连接的桥接
func (c *SimpleGatewayGRPCClient) HandleWebSocket(ctx context.Context, info *sdk.WebSocketConnectInfo) (sdk.WebSocketConn, error) {
	// 建立 gRPC 双向流
	stream, err := c.gateway.HandleWebSocket(ctx)
	if err != nil {
		return nil, fmt.Errorf("gRPC HandleWebSocket 调用失败: %w", err)
	}

	// 序列化账户凭证
	var credsJSON []byte
	var accountID int64
	var proxyURL string
	var rateMul float64
	var maxConc int32
	if info.Account != nil {
		credsJSON, _ = json.Marshal(info.Account.Credentials)
		accountID = info.Account.ID
		proxyURL = info.Account.ProxyURL
		rateMul = info.Account.RateMultiplier
		maxConc = int32(info.Account.MaxConcurrency)
	}

	// 扁平化 headers
	headers := make(map[string]string)
	for k := range info.Headers {
		headers[k] = info.Headers.Get(k)
	}

	// 发送 CONNECT 帧
	if err := stream.Send(&pb.WebSocketFrame{
		Type: pb.WebSocketFrame_CONNECT,
		ConnectInfo: &pb.WebSocketConnectInfo{
			Path:            info.Path,
			Query:           info.Query,
			Headers:         headers,
			RemoteAddr:      info.RemoteAddr,
			ConnectionId:    info.ConnectionID,
			AccountId:       accountID,
			CredentialsJson: credsJSON,
			ProxyUrl:        proxyURL,
			RateMultiplier:  rateMul,
			MaxConcurrency:  maxConc,
		},
	}); err != nil {
		return nil, fmt.Errorf("发送 CONNECT 帧失败: %w", err)
	}

	return &grpcClientWebSocketConn{
		stream: stream,
		info:   info,
	}, nil
}

// grpcClientWebSocketConn 核心侧的 WebSocketConn 实现
// 核心通过此连接与插件进行 WebSocket 消息交换
type grpcClientWebSocketConn struct {
	stream pb.SimpleGatewayService_HandleWebSocketClient
	info   *sdk.WebSocketConnectInfo
}

func (c *grpcClientWebSocketConn) ReadMessage() (int, []byte, error) {
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

func (c *grpcClientWebSocketConn) WriteMessage(msgType int, data []byte) error {
	frameType := pb.WebSocketFrame_TEXT
	if msgType == sdk.WSMessageBinary {
		frameType = pb.WebSocketFrame_BINARY
	}
	return c.stream.Send(&pb.WebSocketFrame{
		Type: frameType,
		Data: data,
	})
}

func (c *grpcClientWebSocketConn) ConnectInfo() *sdk.WebSocketConnectInfo {
	return c.info
}

func (c *grpcClientWebSocketConn) Close(code int, reason string) error {
	_ = c.stream.Send(&pb.WebSocketFrame{
		Type:        pb.WebSocketFrame_CLOSE,
		CloseCode:   int32(code),
		CloseReason: reason,
	})
	return c.stream.CloseSend()
}
