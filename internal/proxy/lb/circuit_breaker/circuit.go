package circuit_breaker

import (
	"github.com/lccxxo/bailuoli/internal/constants"
	"github.com/lccxxo/bailuoli/internal/model"
	"sync"
	"time"
)

// CircuitBreaker 熔断器

type State int

// State 熔断器状态 StateClosed:关闭 StateOpen:打开 StateHalfOpen:半开
const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

// CircuitBreaker 熔断器
type CircuitBreaker struct {
	// 熔断器配置
	config model.CircuitBreakerConfig
	// 熔断器状态
	state State
	// 熔断器统计窗口
	metrics Metrics
	// 状态变更通知
	stopChan chan State
	mu       sync.RWMutex
}

type BreakerManager struct {
	breakers map[string]*CircuitBreaker
	mu       sync.RWMutex
}

func NewBreakerManager() *BreakerManager {
	return &BreakerManager{
		breakers: make(map[string]*CircuitBreaker),
	}
}

func (m *BreakerManager) GetBreaker(key string, config model.CircuitBreakerConfig) *CircuitBreaker {
	m.mu.Lock()
	defer m.mu.Unlock()

	if b, ok := m.breakers[key]; ok {
		return b
	}

	b := NewCircuitBreaker(config)
	m.breakers[key] = b
	return b
}

func (m *BreakerManager) SetBreaker(key string, breaker *model.CircuitBreakerConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.breakers[key] = NewCircuitBreaker(*breaker)

	return
}

// Metrics 熔断器统计窗口
type Metrics struct {
	Requests          int64         // 总请求数
	TotalSuccesses    int64         // 成功数
	TotalFailures     int64         // 失败数
	ConsecutiveErrors int64         // 连续错误计数
	WindowStart       time.Time     // 当前统计窗口开始时间
	WindowDuration    time.Duration // 统计窗口时长（例如10秒）
}

func NewCircuitBreaker(config model.CircuitBreakerConfig) *CircuitBreaker {
	return &CircuitBreaker{
		config: config,
		state:  StateClosed,
		metrics: Metrics{
			WindowDuration: config.WindowDuration,
			WindowStart:    time.Now(),
		},
		stopChan: make(chan State, 10),
	}
}

// shouldTrip 判断是否需要熔断
func (c *CircuitBreaker) shouldTrip() bool {
	//	有两种情况需要熔断：
	//	1. 连续错误次数超过阈值
	//	2. 窗口统计错误率超过阈值
	if c.metrics.ConsecutiveErrors > c.config.ConsecutiveErrorTrigger {
		return true
	}

	total := c.metrics.Requests
	if total == 0 {
		return false
	}

	failure := float64(c.metrics.TotalFailures) / float64(total)
	return failure > c.config.FailureThreshold
}

// 重置熔断器统计窗口
func (c *CircuitBreaker) resetMetrics() {
	c.metrics.ConsecutiveErrors = 0
	c.metrics.TotalFailures = 0
	c.metrics.TotalSuccesses = 0
	c.metrics.Requests = 0
	c.metrics.WindowStart = time.Now()
}

// recordSuccess 记录成功
func (c *CircuitBreaker) recordSuccess() {
	c.metrics.TotalSuccesses++
	c.metrics.Requests++
	c.metrics.ConsecutiveErrors = 0

	// 请求成功后，如果当前状态是半开状态，判断是否需要切换到关闭状态
	if c.state == StateHalfOpen && c.metrics.TotalSuccesses >= c.config.HalfOpenMaxRequests {
		//	关闭熔断
		c.setState(StateClosed)
	}
}

// recordFailure 记录失败
func (c *CircuitBreaker) recordFailure() {
	c.metrics.TotalFailures++
	c.metrics.Requests++
	c.metrics.ConsecutiveErrors++

	//	请求失败后，如果当前状态是半开状态，则重新打开熔断
	if c.state == StateHalfOpen {
		c.setState(StateOpen)
	}
}

// setState 设置熔断器状态
func (c *CircuitBreaker) setState(newState State) {
	c.mu.Lock()
	defer c.mu.Unlock()

	oldState := c.state
	if newState == oldState {
		return
	}

	switch newState {
	case StateOpen:
		// 开启熔断时启动超时计时器
		time.AfterFunc(c.config.OpenStateTimeout, func() {
			c.setState(StateHalfOpen)
		})
		c.resetMetrics()
	case StateHalfOpen:
		c.resetMetrics() // 半开状态重置统计
	case StateClosed:
		c.resetMetrics()
	}

	c.state = newState
	select {
	case c.stopChan <- newState:
	default:
	}
}

// Execute 执行逻辑
func (c *CircuitBreaker) Execute(req func() error) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	switch c.state {
	case StateOpen:
		return constants.ErrCircuitBreakerOpen
	case StateHalfOpen:
		if c.metrics.Requests > c.config.HalfOpenMaxRequests {
			return constants.ErrCircuitBreakerOpen
		}
	}

	if err := req(); err != nil {
		c.recordFailure()
		if c.shouldTrip() {
			c.setState(StateOpen)
		}
		return err
	}

	c.recordSuccess()

	return nil
}
