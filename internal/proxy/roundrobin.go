package proxy

import (
	"github.com/lccxxo/bailuoli/internal/constants"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
)

// RoundRobinLoadBalancer 轮询负载均衡策略
type RoundRobinLoadBalancer struct {
	BaseLoadBalancer
	counter uint64
}

func NewRoundRobinLoadBalancer(upstreams []*url.URL) *RoundRobinLoadBalancer {
	return &RoundRobinLoadBalancer{
		BaseLoadBalancer: BaseLoadBalancer{
			upstreams: upstreams,
			mu:        sync.RWMutex{},
		},
		counter: 0,
	}
}

func (b *RoundRobinLoadBalancer) Next(r *http.Request) (*url.URL, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.upstreams) == 0 {
		return nil, constants.ErrNoUpstreams
	}

	idx := int64(atomic.AddUint64(&b.counter, 1) % uint64(len(b.upstreams)))

	return b.upstreams[idx], nil
}
