package lb

import (
	"github.com/lccxxo/bailuoli/internal/constants"
	"net/http"
	"net/url"
	"sync"
)

// 加权轮询负载均衡策略

type WeightRoundRobinLoadBalancer struct {
	BaseLoadBalancer
	weight        map[string]int // 权重配置
	currentIndex  int            // 当前索引
	currentWeight int            // 当前权重
}

func NewWeightRoundRobinLoadBalancer(upstreams []*url.URL, weight map[string]int) *WeightRoundRobinLoadBalancer {
	safeWeight := make(map[string]int)
	for k, v := range weight {
		if v <= 0 {
			v = 1
		}
		safeWeight[k] = v
	}

	return &WeightRoundRobinLoadBalancer{
		BaseLoadBalancer: BaseLoadBalancer{
			upstreams: upstreams,
			mu:        sync.RWMutex{},
		},
		weight:        safeWeight,
		currentIndex:  -1,
		currentWeight: 0,
	}
}

func (b *WeightRoundRobinLoadBalancer) Next(r *http.Request) (*url.URL, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.upstreams) == 0 {
		return nil, constants.ErrNoUpstreams
	}

	return b.upstreams[b.currentIndex], nil
}

// AddUpstream 添加上游节点 (并设置权重)
func (b *WeightRoundRobinLoadBalancer) AddUpstream(upstream *url.URL) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.weight[upstream.String()] = 1
	b.BaseLoadBalancer.AddUpstream(upstream)
}

// RemoveUpstream 移除上游节点 (并删除权重)
func (b *WeightRoundRobinLoadBalancer) RemoveUpstream(upstream *url.URL) {
	b.mu.Lock()
	defer b.mu.Unlock()

	delete(b.weight, upstream.String())
	b.BaseLoadBalancer.RemoveUpstream(upstream)
}
