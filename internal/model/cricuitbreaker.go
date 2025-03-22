package model

import "time"

// CircuitBreakerConfig 熔断器配置
type CircuitBreakerConfig struct {
	FailureThreshold        float64       `yaml:"failure_threshold"`         // 触发熔断的失败率阈值（例如0.5表示50%）
	ConsecutiveErrorTrigger int64         `yaml:"consecutive_error_trigger"` // 连续错误触发阈值
	HalfOpenMaxRequests     int64         `yaml:"half_open_max_requests"`    // 半开状态允许的最大试探请求数
	OpenStateTimeout        time.Duration `yaml:"open_state_timeout"`        // 熔断后进入半开的等待时间
	WindowDuration          time.Duration `yaml:"window_duration"`           // 统计窗口时长
}
