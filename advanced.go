package sdk

import "context"

// SchedulerService 账户调度服务（核心提供给 Advanced 插件）
type SchedulerService interface {
	// SelectAccount 选择一个可用账户
	SelectAccount(ctx context.Context, req ScheduleRequest) (*Account, error)
	// SelectWithLoadAwareness 负载感知选择账户
	SelectWithLoadAwareness(ctx context.Context, req ScheduleRequest) (*AccountSelection, error)
	// ReportResult 上报调度结果（用于动态调整）
	ReportResult(accountID int64, success bool)
}

// ConcurrencyService 并发控制服务
type ConcurrencyService interface {
	// AcquireSlot 获取并发槽位，返回释放函数
	AcquireSlot(ctx context.Context, accountID int64) (release func(), err error)
}

// RateLimitService 限流服务
type RateLimitService interface {
	// Check 检查用户是否超出限流（RPM/TPM）
	Check(ctx context.Context, userID int64, platform string) error
}

// BillingService 计费服务
type BillingService interface {
	// RecordUsage 记录使用量
	RecordUsage(ctx context.Context, log *UsageLog) error
	// DeductBalance 扣减用户余额
	DeductBalance(ctx context.Context, userID int64, amount float64) error
}
