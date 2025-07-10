package path

type PathMatcher interface {
	// isPattern 检查给定的路径是否为模式
	IsPattern(path string) bool
	// match 检查路径是否与模式匹配
	Match(pattern string, path string) bool
	// matchStart 检查路径是否以给定模式开头
	MatchStart(pattern string, path string) bool
	// extractPathWithinPattern 从模式和路径中提取路径部分
	ExtractPathWithinPattern(pattern string, path string) string
	// extractUriTemplateVariables 从模式和路径中提取 URI 模板变量
	ExtractUriTemplateVariables(pattern string, path string) map[string]string
	// getPatternComparator 获取模式比较器
	GetPatternComparator(path string) func(s1, s2 string) int
	// combine 组合两个模式
	Combine(pattern1 string, pattern2 string) string
}
