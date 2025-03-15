package constants

import "time"

// 存放一些枚举类型值

const (
	DefaultLoadBalance = "round-robin"    // 默认负载均衡策略 round-robin 轮询策略
	DefaultHealthCheck = 1 * time.Minute  // 默认健康检查间隔
	DefaultConnTimeout = 10 * time.Second // 默认连接超时时间
)
