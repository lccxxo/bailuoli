package model

import (
	"github.com/lccxxo/bailuoli/internal/match"
)

type Route struct {
	Name        string             `yaml:"name"`         // 路由名称
	Path        string             `yaml:"path"`         // 匹配路径（精确匹配、前缀匹配、正则匹配）
	Method      string             `yaml:"method"`       // HTTP方法（GET、POST等）
	MatchType   string             `yaml:"match_type"`   // 匹配规则类型（exact、prefix、regex）
	Upstreams   []*UpstreamsConfig `yaml:"upstreams"`    // 后端服务列表
	StripPrefix bool               `yaml:"strip_prefix"` // 是否去除前缀
	LoadBalance LoadBalanceConfig  `yaml:"load_balance"` // 负载均衡配置
	Matcher     match.Matcher      // 匹配器
}

type RouteConfig struct {
	Routes []*Route `yaml:"routes"`
}

type UpstreamsConfig struct {
	Host                 string               `yaml:"host"`            // 必填，目标主机（如：10.0.0.1 或 backend.service）
	Path                 string               `yaml:"path"`            // 转发后的基础路径（默认为空）
	Method               string               `yaml:"method"`          // 请求方法（默认 GET）
	ContentType          string               `yaml:"content_type"`    // 请求头中的 Content-Type（默认 application/json）
	CircuitBreakerConfig CircuitBreakerConfig `yaml:"circuit_breaker"` // 熔断器配置
}
