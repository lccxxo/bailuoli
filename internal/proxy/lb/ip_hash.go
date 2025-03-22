package lb

import (
	"github.com/lccxxo/bailuoli/internal/constants"
	"hash/fnv"
	"net"
	"net/http"
	"net/url"
	"sync"
)

// IP哈系负载均衡策略

type IPHashLoadBalancer struct {
	BaseLoadBalancer
}

func NewIPHashLoadBalancer(upstreams []*url.URL) *IPHashLoadBalancer {
	return &IPHashLoadBalancer{BaseLoadBalancer{
		upstreams: upstreams,
		mu:        sync.RWMutex{},
	}}
}

func (b *IPHashLoadBalancer) Next(r *http.Request) (*url.URL, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	ip := getClientIP(r)
	if ip == "" {
		return nil, constants.ErrNoClientIP
	}

	hasher := fnv.New32a()
	hasher.Write([]byte(ip))
	hash := int(hasher.Sum32())
	index := hash % len(b.upstreams)
	if index < 0 {
		index = -index
	}
	return b.upstreams[index], nil
}

// 获取客户端IP
func getClientIP(r *http.Request) string {
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return ip
	}
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	host, _, _ := net.SplitHostPort(r.RemoteAddr)
	return host
}
