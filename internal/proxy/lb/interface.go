package lb

import (
	"github.com/lccxxo/bailuoli/internal/proxy/lb/healthy"
	"net/http"
	"net/url"
	"sync"
)

// LoadBalancer 负载均衡转发器
type LoadBalancer interface {
	BaseLB
	Next(r *http.Request) (*url.URL, error)
}

type BaseLB interface {
	AddUpstream(upstream *url.URL)
	RemoveUpstream(upstream *url.URL)
	SetHealthChecker(checker *healthy.Checker)
}

type BaseLoadBalancer struct {
	mu        sync.RWMutex
	upstreams []*url.URL
	checker   *healthy.Checker
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

func (b *BaseLoadBalancer) SetHealthChecker(checker *healthy.Checker) {
	b.checker = checker
	return
}

// 只获取健康的上游节点
func (b *BaseLoadBalancer) healthyUpstreams() []*url.URL {
	b.mu.RLock()
	defer b.mu.RUnlock()

	var urls []*url.URL
	for _, u := range b.upstreams {
		if b.checker == nil || b.checker.IsHealthy(u.String()) {
			urls = append(urls, u)
		}
	}
	return urls
}
