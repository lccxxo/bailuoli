package lb

import (
	"net/http"
	"net/url"
	"sync"
)

// LoadBalancer 负载均衡转发器
type LoadBalancer interface {
	Next(r *http.Request) (*url.URL, error) // 策略实现
	AddUpstream(upstream *url.URL)          // 添加上游节点
	RemoveUpstream(upstream *url.URL)       // 移除上游节点
}

type BaseLoadBalancer struct {
	mu        sync.RWMutex
	upstreams []*url.URL
}

func (b *BaseLoadBalancer) AddUpstream(upstream *url.URL) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, u := range b.upstreams {
		if u.String() == upstream.String() {
			return
		}
	}
	b.upstreams = append(b.upstreams, upstream)
}

func (b *BaseLoadBalancer) RemoveUpstream(upstream *url.URL) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for i, u := range b.upstreams {
		if u.String() == upstream.String() {
			b.upstreams = append(b.upstreams[:i], b.upstreams[i+1:]...)
		}
	}
}
