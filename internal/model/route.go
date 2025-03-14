package model

import (
	"github.com/lccxxo/bailuoli/internal/match"
	"time"
)

type Route struct {
	Name        string            `yaml:"name"`         // 路由名称
	Path        string            `yaml:"path"`         // 匹配路径（精确匹配、前缀匹配、正则匹配）
	Method      string            `yaml:"method"`       // HTTP方法（GET、POST等）
	MatchType   string            `yaml:"match_type"`   // 匹配规则类型（exact、prefix、regex）
	Upstreams   []string          `yaml:"upstreams"`    // 后端服务列表
	StripPrefix bool              `yaml:"strip_prefix"` // 是否去除前缀
	LoadBalance LoadBalanceConfig `yaml:"load_balance"` // 负载均衡配置
	Matcher     match.Matcher     // 匹配器
}

type RouteConfig struct {
	Routes []*Route `yaml:"routes"`
}

// LoadBalanceConfig 负载均衡配置
type LoadBalanceConfig struct {
	Strategy    string         `yaml:"strategy"`     // 负载均衡策略名称 默认 round-robin
	Weighted    map[string]int `yaml:"weight"`       // 权重配置
	MaxConn     int            `yaml:"max_conn"`     // 最大连接数
	HealthCheck time.Duration  `yaml:"health_check"` // 健康检查间隔
	Timeout     time.Duration  `yaml:"timeout"`      // 超时时间间隔
}
