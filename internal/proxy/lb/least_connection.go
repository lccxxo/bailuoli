package lb

import (
	"context"
	"github.com/lccxxo/bailuoli/internal/constants"
	"math"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
)

// 最少连接负载均衡策略

type LeastConnectionLoadBalancer struct {
	BaseLoadBalancer
	connCounts *ConnCounter // 连接计数器  key: host value: 连接数
}

func NewLeastConnectionLoadBalancer(upstreams []*url.URL) *LeastConnectionLoadBalancer {
	connCounts := ConnCounter{
		counters: make(map[string]*atomic.Int64),
	}

	for _, u := range upstreams {
		connCounts.counters[u.Host] = &atomic.Int64{}
	}

	return &LeastConnectionLoadBalancer{
		BaseLoadBalancer: BaseLoadBalancer{
			upstreams: upstreams,
			mu:        sync.RWMutex{},
		},
		connCounts: &connCounts,
	}
}

func (b *LeastConnectionLoadBalancer) Next(r *http.Request) (*url.URL, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	healthy := b.healthyUpstreams()
	if len(healthy) == 0 {
		return nil, constants.ErrNoHealthyUpstreams
	}

	// 选择最少连接的上游节点
	var minURL *url.URL
	var minCount = int64(math.MaxInt64)
	for _, upstream := range healthy {
		count := b.connCounts.counters[upstream.Host].Load()

		if count < minCount {
			minCount = count
			minURL = upstream
		}
	}

	if minURL == nil {
		return nil, constants.ErrNoUpstreams
	}

	// 连接数加1
	b.connCounts.Acquire(minURL.Host)

	// 将连接数减少逻辑放到上下文
	ctx := context.WithValue(r.Context(), "least_conn_counter", func() {
		b.connCounts.Release(minURL.Host)
	})
	*r = *r.WithContext(ctx)

	return minURL, nil
}

// AddUpstream 重写添加方法（初始化连接计数）
func (b *LeastConnectionLoadBalancer) AddUpstream(upstream *url.URL) {
	b.BaseLoadBalancer.AddUpstream(upstream)
	b.connCounts.counters[upstream.Host] = &atomic.Int64{}
	b.connCounts.counters[upstream.Host].Store(0)
}

// RemoveUpstream 重写移除方法（清理连接计数）
func (b *LeastConnectionLoadBalancer) RemoveUpstream(upstream *url.URL) {
	b.BaseLoadBalancer.RemoveUpstream(upstream)
	delete(b.connCounts.counters, upstream.Host)
}
