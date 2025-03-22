package model

// LoadBalanceConfig 负载均衡配置
type LoadBalanceConfig struct {
	Strategy     string         `yaml:"strategy"`      // 负载均衡策略名称 默认 round-robin
	Weighted     map[string]int `yaml:"weight"`        // 权重配置
	MaxConn      int            `yaml:"max_conn"`      // 最大连接数
	HealthyCheck HealthyConfig  `yaml:"healthy_check"` // 健康检查配置
}
