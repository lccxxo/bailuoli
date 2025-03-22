package lb

import (
	"github.com/lccxxo/bailuoli/internal/constants"
	"math/rand"
	"net/http"
	"net/url"
	"sync"
)

// RandomLoadBalancer 随机负载均衡策略

type RandomLoadBalancer struct {
	BaseLoadBalancer
}

func NewRandomLoadBalancer(upstreams []*url.URL) *RandomLoadBalancer {
	return &RandomLoadBalancer{BaseLoadBalancer{
		upstreams: upstreams,
		mu:        sync.RWMutex{},
	}}
}

func (b *RandomLoadBalancer) Next(r *http.Request) (*url.URL, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	//	判断是否有上游节点
	if len(b.upstreams) == 0 {
		return nil, constants.ErrNoUpstreams
	}

	//	随机选择一个上游节点
	return b.upstreams[rand.Intn(len(b.upstreams))], nil
}
