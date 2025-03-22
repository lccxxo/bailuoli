package controller

import (
	"context"
	"fmt"
	"github.com/lccxxo/bailuoli/internal/proxy/lb/circuit_breaker"
	"net/http"
	"net/url"
	"regexp"
	"sync"

	"github.com/lccxxo/bailuoli/internal/match"
	"github.com/lccxxo/bailuoli/internal/model"
	"github.com/lccxxo/bailuoli/internal/proxy"
	"github.com/lccxxo/bailuoli/internal/proxy/lb/healthy"
	"github.com/lccxxo/bailuoli/internal/validator"
)

// 路由匹配规则

// CreateMatcher 工厂模式 创建匹配器
func CreateMatcher(route *model.Route) (match.Matcher, error) {
	switch route.MatchType {
	case "exact":
		return &match.ExactMatcher{Path: route.Path}, nil
	case "prefix":
		return &match.PrefixMatcher{Prefix: route.Path}, nil
	case "regex":
		re, err := regexp.Compile(route.Path)
		if err != nil {
			return nil, fmt.Errorf("invalid regex pattern: %w", err)
		}
		return &match.RegexMatcher{Re: re}, nil
	default:
		return nil, fmt.Errorf("unknown match type: %s", route.MatchType)
	}
}

type Router struct {
	Routes         []*model.Route                  // 路由表
	proxies        map[string]http.Handler         // 存储的是路由名称 -》反向代理实例
	healthCheckers map[string]*healthy.Checker     // 路由名称 -> 健康检查器
	validator      validator.Validator             // 验证责任链
	breakerManager *circuit_breaker.BreakerManager // 熔断器管理器
	mu             sync.RWMutex
}

func NewRouter(routes []*model.Route) *Router {
	r := &Router{
		proxies:        make(map[string]http.Handler),
		validator:      validator.NewValidationChain(),
		breakerManager: circuit_breaker.NewBreakerManager(),
	}
	if err := r.UpdateRoutes(routes); err != nil {
		return nil
	}
	return r
}

func (r *Router) UpdateRoutes(newRoutes []*model.Route) error {
	// 1. 验证新配置合法性
	for _, route := range newRoutes {
		if err := r.validator.Validate(route); err != nil {
			return fmt.Errorf("invalid route %s: %w", route.Name, err)
		}
	}

	// 创建新的转发路由映射
	proxies := make(map[string]http.Handler)
	// 创建新健康检查器
	newCheckers := make(map[string]*healthy.Checker)
	// 创建新的熔断器

	for _, route := range newRoutes {
		for _, upstream := range route.Upstreams {
			key := upstream.Host + upstream.Path
			r.breakerManager.SetBreaker(key, &upstream.CircuitBreakerConfig)
		}
	}

	for _, route := range newRoutes {
		matcher, err := CreateMatcher(route)
		if err != nil {
			return err
		}
		route.Matcher = matcher
		proxies[route.Name] = proxy.NewLoadBalanceReverseProxy(route.LoadBalance, route.Upstreams, r.breakerManager)
	}

	for _, route := range newRoutes {
		// 转换上游地址
		upstreams := convertToURLs(route.Upstreams)

		// 创建健康检查器
		checker := healthy.NewChecker(model.HealthyConfig{
			Interval:           route.LoadBalance.HealthyCheck.Interval,
			Timeout:            route.LoadBalance.HealthyCheck.Timeout,
			Path:               route.LoadBalance.HealthyCheck.Path,
			SuccessCode:        route.LoadBalance.HealthyCheck.SuccessCode,
			HealthyThreshold:   route.LoadBalance.HealthyCheck.HealthyThreshold,
			UnhealthyThreshold: route.LoadBalance.HealthyCheck.UnhealthyThreshold,
		})
		ctx, cancel := context.WithCancel(context.Background())
		checker.UpdateUpstreams(upstreams)
		go checker.Run(ctx)

		checker.Cancel = cancel
		newCheckers[route.Name] = checker

		// 创建熔断器

	}

	// 2. 原子化替换路由表
	r.mu.Lock()
	r.Routes = newRoutes
	r.proxies = proxies

	oldHealthCheckers := r.healthCheckers

	r.healthCheckers = newCheckers
	r.mu.Unlock()

	//  清理旧的健康检查
	for _, cancel := range oldHealthCheckers {
		cancel.Cancel()
	}

	return nil
}

// MatchRoute 路由匹配规则
func (r *Router) MatchRoute(req *http.Request) (*model.Route, http.Handler) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// 匹配路由方法
	for _, route := range r.Routes {
		if route.Method != "" && route.Method != req.Method {
			continue
		}

		// 如果匹配上则返回
		if route.Matcher.Match(req.URL.Path) {
			return route, r.proxies[route.Name]
		}
	}

	return nil, nil
}

// 辅助函数：转换配置到URL列表
func convertToURLs(upstreams []*model.UpstreamsConfig) []*url.URL {
	var urls []*url.URL
	for _, u := range upstreams {
		parsed, _ := url.Parse(u.Host + u.Path)
		urls = append(urls, parsed)
	}
	return urls
}
