package controller

import (
	"fmt"
	"github.com/lccxxo/bailuoli/internal/model"
	"github.com/lccxxo/bailuoli/internal/proxy"
	"net/http"
	"regexp"
	"strings"
	"sync"
)

type Router struct {
	Routes  []*model.Route
	proxies map[string]http.Handler // 存储的是路由名称 -》反向代理实例
	mu      sync.RWMutex
}

func NewRouter(routes []*model.Route) *Router {
	r := &Router{
		Routes:  routes,
		proxies: make(map[string]http.Handler),
	}
	r.initProxies()
	return r
}

func (r *Router) initProxies() {
	for _, route := range r.Routes {
		reverseProxy := proxy.NewLoadBalanceReverseProxy(route.Upstreams)
		r.proxies[route.Name] = reverseProxy
	}
}

func (r *Router) UpdateRoutes(newRoutes []*model.Route) error {
	// 1. 验证新配置合法性
	if err := validateRoutes(newRoutes); err != nil {
		return err
	}

	// 2. 原子化替换路由表
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Routes = newRoutes

	return nil
}

// MatchRoute 路由匹配规则
func (r *Router) MatchRoute(req *http.Request) (*model.Route, http.Handler) {
	for _, route := range r.Routes {
		// 没匹配上请求的路由
		if route.Method != "" && route.Method != req.Method {
			continue
		}

		switch route.MatchType {
		case "exact":
			if req.URL.Path == route.Path {
				return route, r.proxies[route.Name]
			}
		case "prefix":
			if strings.HasPrefix(req.URL.Path, route.Path) {
				return route, r.proxies[route.Name]
			}
		case "regex":
			if matched, _ := regexp.MatchString(req.URL.Path, route.Path); matched {
				return route, r.proxies[route.Name]
			}
		}
	}

	return nil, nil
}

// 验证路由规则
func validateRoutes(routes []*model.Route) error {
	for _, route := range routes {
		if route.Path == "" {
			return fmt.Errorf("route %s: path cannot be empty", route.Name)
		}
	}
	return nil
}
