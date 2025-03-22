package model

import "time"

type HealthyConfig struct {
	Interval           time.Duration `yaml:"interval"`            // 健康检查间隔
	Timeout            time.Duration `yaml:"timeout"`             // 超时时间间隔
	Path               string        `yaml:"path"`                // 健康检查路径
	SuccessCode        []int         `yaml:"success_code"`        // 健康检查成功状态码
	HealthyThreshold   int           `yaml:"healthy_threshold"`   // 健康检查成功阈值
	UnhealthyThreshold int           `yaml:"unhealthy_threshold"` // 健康检查失败阈值
}
