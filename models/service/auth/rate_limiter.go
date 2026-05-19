package auth

import (
	"context"
	"time"

	"agent/models/dao"
)

// RateLimiter 频率限制服务
type RateLimiter struct {
	dao *dao.RateLimitDAO
}

// RateLimitConfig 频率限制配置
type RateLimitConfig struct {
	Window time.Duration // 时间窗口
	MaxReq int           // 最大请求数
}

// 各角色的频率限制配置
var roleRateLimits = map[string]RateLimitConfig{
	"user":  {Window: 5 * time.Minute, MaxReq: 1}, // 普通用户：5分钟1次
	"vip":   {Window: 1 * time.Minute, MaxReq: 2}, // VIP用户：1分钟2次
	"admin": {Window: 0, MaxReq: 0},               // 管理员：无限制
}

// ErrRateLimitExceeded 频率限制错误
var ErrRateLimitExceeded = "请求过于频繁"

// NewRateLimiter 创建频率限制服务
func NewRateLimiter(dao *dao.RateLimitDAO) *RateLimiter {
	return &RateLimiter{dao: dao}
}

// CheckLimit 检查是否超过频率限制
// 返回：是否允许请求、需要等待的秒数、错误
func (r *RateLimiter) CheckLimit(ctx context.Context, userID int64, role string) (allowed bool, waitSeconds int, err error) {
	config, ok := roleRateLimits[role]
	if !ok {
		// 未知角色使用普通用户限制
		config = roleRateLimits["user"]
	}

	// 管理员无限制
	if config.MaxReq == 0 {
		return true, 0, nil
	}

	// 计算时间窗口起始时间
	windowStart := time.Now().Add(-config.Window)

	// 查询窗口内请求数
	count, err := r.dao.CountRequests(ctx, userID, "chat", windowStart)
	if err != nil {
		return false, 0, err
	}

	// 未超过限制
	if count < config.MaxReq {
		return true, 0, nil
	}

	// 超过限制，计算需要等待的时间
	oldest, err := r.dao.GetOldestInWindow(ctx, userID, "chat", windowStart)
	if err != nil {
		return false, 0, err
	}

	// 计算等待时间：最早请求时间 + 窗口时长 - 当前时间
	waitDuration := oldest.Add(config.Window).Sub(time.Now())
	if waitDuration < 0 {
		waitDuration = 0
	}

	return false, int(waitDuration.Seconds()) + 1, nil // 加1秒确保安全
}

// RecordRequest 记录一次请求
func (r *RateLimiter) RecordRequest(ctx context.Context, userID int64) error {
	return r.dao.CreateRecord(ctx, userID, "chat", time.Now())
}

// GetRoleRateLimit 获取角色的频率限制配置（供前端展示）
func GetRoleRateLimit(role string) RateLimitConfig {
	if config, ok := roleRateLimits[role]; ok {
		return config
	}
	return roleRateLimits["user"]
}
