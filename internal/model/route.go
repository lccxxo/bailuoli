package model

type Route struct {
	Name        string   `yaml:"name"`         // 路由名称
	Path        string   `yaml:"path"`         // 匹配路径（精确匹配、前缀匹配、正则匹配）
	Method      string   `yaml:"method"`       // HTTP方法（GET、POST等）
	MatchType   string   `yaml:"match_type"`   // 匹配规则类型（exact、prefix、regex）
	Upstreams   []string `yaml:"upstreams"`    // 后端服务列表
	StripPrefix bool     `yaml:"strip_prefix"` // 是否去除前缀
}

type RouteConfig struct {
	Routes []*Route `yaml:"routes"`
}
