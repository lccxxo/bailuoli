package proxy

import (
	"github.com/lccxxo/bailuoli/internal/logger"
	"github.com/lccxxo/bailuoli/internal/model"
	"go.uber.org/zap"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"time"
)

// 用于转发请求

// 全局连接池配置（复用TCP连接）
var transport = &http.Transport{
	Proxy: http.ProxyFromEnvironment,
	DialContext: (&net.Dialer{
		Timeout:   30 * time.Second, // 连接超时
		KeepAlive: 60 * time.Second, // 保持连接时间
	}).DialContext,
	MaxIdleConns:          1000,             // 最大空闲连接数
	MaxIdleConnsPerHost:   500,              // 每个主机最大空闲连接
	IdleConnTimeout:       90 * time.Second, // 空闲连接超时
	TLSHandshakeTimeout:   10 * time.Second, // TLS握手超时
	ExpectContinueTimeout: 1 * time.Second,  // Expect头超时
	ForceAttemptHTTP2:     true,             // 启用HTTP/2
}

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

type LoadBalanceReverseProxy struct {
	loadBalance LoadBalancer
	proxy       *httputil.ReverseProxy
	reqPool     sync.Pool
}

func NewLoadBalanceReverseProxy(loadBalanceConfig model.LoadBalanceConfig, upstreams []*model.UpstreamsConfig) *LoadBalanceReverseProxy {
	urls := make([]*url.URL, 0, len(upstreams))
	for _, u := range upstreams {
		parse, _ := url.Parse(u.Host + u.Path)
		urls = append(urls, parse)
	}

	var lb LoadBalancer
	switch loadBalanceConfig.Strategy {
	case "round":
		lb = NewRandomLoadBalancer(urls)
	case "round_robin":
		lb = NewRoundRobinLoadBalancer(urls)
	case "ip_hash":
		lb = NewIPHashLoadBalancer(urls)
	case "weighted":
		lb = NewWeightRoundRobinLoadBalancer(urls, loadBalanceConfig.Weighted)
	case "least-connections":
		lb = NewLeastConnectionLoadBalancer(urls)
	default:
		return nil
	}

	p := &LoadBalanceReverseProxy{
		loadBalance: lb,
	}

	p.reqPool.New = func() interface{} {
		return &requestContext{proxy: p}
	}

	p.proxy = &httputil.ReverseProxy{
		Transport:      transport,
		Director:       p.director,
		ModifyResponse: p.modifyResponse,
		ErrorHandler:   p.errHandler,
	}

	return p
}

func (p *LoadBalanceReverseProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := p.reqPool.Get().(*requestContext)
	defer p.reqPool.Put(ctx)

	done := make(chan struct{})

	go func() {
		defer close(done)
		ctx.process(w, r)
	}()

	<-done
}

/*
	1. director：在每次请求被转发前调用Director函数，自动去除前缀
	2. modifyResponse：如果是LeastConnectionLoadBalancer策略，更新连接计数
	3. errHandler：处理转发时出现的错误
*/

// 请求预处理
func (p *LoadBalanceReverseProxy) director(r *http.Request) {
	logger.Logger.Info("request", zap.String("url", r.URL.String()))
	if route := r.Context().Value("route").(*model.Route); route != nil {
		if route.StripPrefix {
			r.URL.Path = strings.TrimPrefix(r.URL.Path, route.Path)
		}
	}
}

// 错误处理
func (p *LoadBalanceReverseProxy) errHandler(w http.ResponseWriter, r *http.Request, err error) {
	// 释放最小连接策略的计数器
	if release, ok := r.Context().Value("least_conn_counter").(func()); ok {
		release()
	}
	// todo 可以记录故障的上游节点
	http.Error(w, "Gateway error", http.StatusBadGateway)
}

// 转发响应后的钩子函数
func (p *LoadBalanceReverseProxy) modifyResponse(resp *http.Response) error {
	if host := resp.Request.URL.Host; host != "" {
		if release, ok := resp.Request.Context().Value("least_conn_counter").(func()); ok {
			defer release()
		}
	}
	return nil
}

type requestContext struct {
	proxy *LoadBalanceReverseProxy
}

func (p *requestContext) process(w http.ResponseWriter, r *http.Request) {
	target, err := p.proxy.loadBalance.Next(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	r.URL.Host = target.Host
	r.URL.Scheme = target.Scheme
	r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
	r.Host = target.Host
	r.URL.Path = target.Path

	p.proxy.proxy.ServeHTTP(w, r)
}
