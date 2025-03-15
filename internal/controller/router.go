package controller

import (
	"fmt"
	"github.com/lccxxo/bailuoli/internal/match"
	"github.com/lccxxo/bailuoli/internal/model"
	"github.com/lccxxo/bailuoli/internal/proxy"
	"github.com/lccxxo/bailuoli/internal/validator"
	"net/http"
	"regexp"
	"sync"
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
	Routes    []*model.Route          // 路由表
	proxies   map[string]http.Handler // 存储的是路由名称 -》反向代理实例
	validator validator.Validator     // 验证责任链
	mu        sync.RWMutex
}

func NewRouter(routes []*model.Route) *Router {
	r := &Router{
		proxies:   make(map[string]http.Handler),
		validator: validator.NewValidationChain(),
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

	proxies := make(map[string]http.Handler)
	for _, route := range newRoutes {
		matcher, err := CreateMatcher(route)
		if err != nil {
			return err
		}
		route.Matcher = matcher
		proxies[route.Name] = proxy.NewLoadBalanceReverseProxy(route.LoadBalance, route.Upstreams)
	}

	// 2. 原子化替换路由表
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Routes = newRoutes
	r.proxies = proxies

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
