package match

import "regexp"

// 正则匹配策略

type RegexMatcher struct {
	Re *regexp.Regexp
}

func (r *RegexMatcher) Match(path string) bool {
	return r.Re.MatchString(path)
}
