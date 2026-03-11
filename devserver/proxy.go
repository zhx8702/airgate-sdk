package devserver

import (
	"context"
	"io"
	"log"
	"log/slog"
	"net/http"
	"strings"

	sdk "github.com/DouDOU-start/airgate-sdk"
	"github.com/gorilla/websocket"
)

// ProxyHandler 将请求代理给插件
type ProxyHandler struct {
	plugin sdk.GatewayPlugin
	store  *AccountStore
}

func (p *ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if isWebSocketUpgrade(r) {
		p.handleWebSocket(w, r)
		return
	}
	p.handleHTTP(w, r)
}

func isWebSocketUpgrade(r *http.Request) bool {
	return strings.EqualFold(r.Header.Get("Upgrade"), "websocket")
}

func (p *ProxyHandler) handleHTTP(w http.ResponseWriter, r *http.Request) {
	account := p.selectAccount()
	if account == nil {
		http.Error(w, `{"error":"no accounts configured"}`, http.StatusServiceUnavailable)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, `{"error":"read body failed"}`, http.StatusBadRequest)
		return
	}

	stream := strings.Contains(string(body), `"stream":true`) ||
		strings.Contains(string(body), `"stream": true`)

	headers := r.Header.Clone()
	headers.Set("X-Forwarded-Path", r.URL.Path)

	slog.Debug("[请求] 收到转发请求",
		"method", r.Method,
		"path", r.URL.Path,
		"body", string(body))

	fwdReq := &sdk.ForwardRequest{
		Account: &sdk.Account{
			ID:          account.ID,
			Credentials: account.Credentials,
			ProxyURL:    account.ProxyURL,
		},
		Body:    body,
		Headers: headers,
		Stream:  stream,
		Writer:  w,
	}

	result, err := p.plugin.Forward(r.Context(), fwdReq)
	if err != nil {
		log.Printf("Forward 失败: %v", err)
		if result == nil || result.StatusCode == 0 {
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadGateway)
		}
		return
	}

	log.Printf("Forward 完成: status=%d model=%s input=%d output=%d duration=%s",
		result.StatusCode, result.Model, result.InputTokens, result.OutputTokens, result.Duration)
}

func (p *ProxyHandler) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	account := p.selectAccount()
	if account == nil {
		http.Error(w, `{"error":"no accounts configured"}`, http.StatusServiceUnavailable)
		return
	}

	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	wsConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket 升级失败: %v", err)
		return
	}
	defer wsConn.Close()

	conn := &devWebSocketConn{
		conn: wsConn,
		info: &sdk.WebSocketConnectInfo{
			Path:       r.URL.Path,
			Query:      r.URL.RawQuery,
			Headers:    r.Header,
			RemoteAddr: r.RemoteAddr,
			Account: &sdk.Account{
				ID:          account.ID,
				Credentials: account.Credentials,
				ProxyURL:    account.ProxyURL,
			},
		},
	}

	log.Printf("WebSocket 连接建立: %s, account=%d", r.URL.Path, account.ID)

	if _, err := p.plugin.HandleWebSocket(context.Background(), conn); err != nil {
		log.Printf("WebSocket 处理结束: %v", err)
	}
}

func (p *ProxyHandler) selectAccount() *DevAccount {
	return p.store.First()
}

// devWebSocketConn 包装 gorilla/websocket.Conn 为 sdk.WebSocketConn
type devWebSocketConn struct {
	conn *websocket.Conn
	info *sdk.WebSocketConnectInfo
}

func (c *devWebSocketConn) ReadMessage() (int, []byte, error) {
	msgType, data, err := c.conn.ReadMessage()
	if err != nil {
		return 0, nil, err
	}
	sdkType := sdk.WSMessageText
	if msgType == websocket.BinaryMessage {
		sdkType = sdk.WSMessageBinary
	}
	return sdkType, data, nil
}

func (c *devWebSocketConn) WriteMessage(messageType int, data []byte) error {
	wsType := websocket.TextMessage
	if messageType == sdk.WSMessageBinary {
		wsType = websocket.BinaryMessage
	}
	return c.conn.WriteMessage(wsType, data)
}

func (c *devWebSocketConn) ConnectInfo() *sdk.WebSocketConnectInfo {
	return c.info
}

func (c *devWebSocketConn) Close(code int, reason string) error {
	msg := websocket.FormatCloseMessage(code, reason)
	_ = c.conn.WriteMessage(websocket.CloseMessage, msg)
	return c.conn.Close()
}
