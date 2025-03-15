package match

// 精确匹配策略

type ExactMatcher struct {
	Path string
}

func (e *ExactMatcher) Match(path string) bool {
	return e.Path == path
}
