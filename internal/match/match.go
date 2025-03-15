package match

// 上游路由匹配器

// Matcher 接口定义
type Matcher interface {
	Match(path string) bool
}
