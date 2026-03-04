package sdk

import (
	"context"
	"net/http"
)

// PaymentPlugin 支付插件接口
type PaymentPlugin interface {
	Plugin
	// CreateOrder 创建支付订单，返回支付链接或二维码
	CreateOrder(ctx context.Context, req CreateOrderRequest) (*PaymentOrder, error)
	// QueryOrder 查询订单状态
	QueryOrder(ctx context.Context, orderID string) (*PaymentOrder, error)
	// HandleCallback 处理支付平台异步回调
	HandleCallback(w http.ResponseWriter, r *http.Request) error
}

// CreateOrderRequest 创建订单请求
type CreateOrderRequest struct {
	UserID  int64   `json:"user_id"`
	Amount  float64 `json:"amount"`  // 充值金额（USD）
	Channel string  `json:"channel"` // 支付渠道
}

// PaymentOrder 支付订单
type PaymentOrder struct {
	OrderID string  `json:"order_id"`
	PayURL  string  `json:"pay_url"` // 支付链接
	Status  string  `json:"status"`  // pending / paid / failed / expired
	Amount  float64 `json:"amount"`
}
