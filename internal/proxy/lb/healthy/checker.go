package healthy

import (
	"context"
	"github.com/lccxxo/bailuoli/internal/proxy/lb/circuit_breaker"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/lccxxo/bailuoli/internal/logger"
	"github.com/lccxxo/bailuoli/internal/model"
	"github.com/lccxxo/bailuoli/pkg/utils"
	"go.uber.org/zap"
)

// 主动定时检查

type Checker struct {
	client     http.Client
	config     model.HealthyConfig
	statusMap  sync.Map // key: upstream host, value: *Status
	upstreams  []*url.URL
	breakerMap map[string]*circuit_breaker.CircuitBreaker // 熔断配置 key: upstream host value: *CircuitBreaker
	Cancel     context.CancelFunc
}

type Status struct {
	mu          sync.Mutex
	failures    int
	successes   int
	lastChecked time.Time
	isHealthy   bool
}

func NewChecker(config model.HealthyConfig) *Checker {
	return &Checker{
		client: http.Client{
			Timeout: config.Timeout,
		},
		config: config,
	}
}

func (c *Checker) Run(ctx context.Context) {
	ticker := time.NewTicker(c.config.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.checkAllUpstream(ctx, c.upstreams)
		case <-ctx.Done():
			return
		}
	}
}

func (c *Checker) checkAllUpstream(ctx context.Context, upstream []*url.URL) {
	var wg sync.WaitGroup
	for _, u := range upstream {
		wg.Add(1)
		go func(u *url.URL) {
			defer wg.Done()
			c.checkSingleUpstream(ctx, u)
		}(u)
	}
}

func (c *Checker) checkSingleUpstream(ctx context.Context, upstream *url.URL) {
	checkURL := upstream.String()

	resp, err := c.client.Get(checkURL)
	status, _ := c.statusMap.LoadOrStore(checkURL, &Status{})
	s := status.(*Status)

	s.mu.Lock()
	defer s.mu.Unlock()

	success := err == nil && utils.Contains(c.config.SuccessCode, resp.StatusCode)

	if success {
		s.successes++
		s.failures = 0
	} else {
		s.failures++
		s.successes = 0
	}

	// 更改健康状态
	if s.failures > c.config.UnhealthyThreshold {
		s.isHealthy = false
	} else if s.successes > c.config.HealthyThreshold {
		s.isHealthy = true
	}

	s.lastChecked = time.Now()

	if !s.isHealthy {
		logger.Logger.Warn("check upstream", zap.String("upstream", upstream.String()), zap.Bool("isHealthy", s.isHealthy))
	}
}

func (c *Checker) IsHealthy(host string) bool {
	status, ok := c.statusMap.Load(host)
	if !ok {
		return false
	}
	return status.(*Status).isHealthy
}

func (c *Checker) UpdateUpstreams(upstreams []*url.URL) {
	c.upstreams = upstreams

	c.statusMap.Range(func(key, value interface{}) bool {
		if !containsURL(upstreams, key.(string)) {
			c.statusMap.Delete(key)
		}
		return true
	})
}

func containsURL(list []*url.URL, target string) bool {
	for _, u := range list {
		if u.String() == target {
			return true
		}
	}
	return false
}
