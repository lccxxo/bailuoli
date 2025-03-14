package proxy

import (
	"github.com/lccxxo/bailuoli/internal/model"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
)

// 用于转发请求

type LoadBalanceReverseProxy struct {
	UpStreams []*url.URL
	mu        sync.RWMutex
}

func NewLoadBalanceReverseProxy(upstreams []string) *LoadBalanceReverseProxy {
	var urls []*url.URL
	for _, u := range upstreams {
		parse, _ := url.Parse(u)
		urls = append(urls, parse)
	}
	return &LoadBalanceReverseProxy{
		UpStreams: urls,
	}
}

// 负载均衡逻辑 随机轮询策略
func (p *LoadBalanceReverseProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if len(p.UpStreams) == 0 {
		http.Error(w, "no upstream available", http.StatusBadGateway)
		return
	}

	target := p.UpStreams[rand.Intn(len(p.UpStreams))]
	proxy := httputil.NewSingleHostReverseProxy(target)

	// 剔除路径
	if route := r.Context().Value("route").(*model.Route); route.StripPrefix {
		r.URL.Path = strings.TrimPrefix(r.URL.Path, route.Path)
	}

	proxy.ServeHTTP(w, r)
}
