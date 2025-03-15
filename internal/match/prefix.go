package match

import "strings"

// 前缀匹配策略

type PrefixMatcher struct {
	Prefix string
}

func (p *PrefixMatcher) Match(path string) bool {
	return strings.HasPrefix(path, p.Prefix)
}
