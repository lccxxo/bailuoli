package constants

import "errors"

// 存放标准的错误信息

var (
	ErrNoUpstreams        = errors.New("no upstream available")
	ErrNoClientIP         = errors.New("cannot get client IP")
	ErrNoWeightedConfig   = errors.New("cannot get weighted config")
	ErrCountIllegal       = errors.New("count is illegal")
	ErrNoHealthyUpstreams = errors.New("no healthy upstreams")
	ErrCircuitBreakerOpen = errors.New("circuit breaker is open")
)
