package path_matcher_impl

import (
	"github.com/SATA260/Doge-Token/utils/string_tool"
	"regexp"
	"strings"
)

const (
	DEFAULT_PATH_SEPARATOR  = "/"
	CACHE_TURNOFF_THRESHOLD = 65536
)

var (
	VARIABLE_PATTERN         = regexp.MustCompile(`\{[^/]+\}`)
	WILDCARD_CHARS           = []rune{'*', '?', '{'}
	GLOB_PATTERN             = regexp.MustCompile(`\?|\*|\{((?:\{[^/]+\}|[^/{}]|\\[{}])+?)\}`)
	DEFAULT_VARIABLE_PATTERN = `((?s).*)`
)

type AntPathMatcher struct {
	pathSeparator             string
	pathSeparatorPatternCache *PathSeparatorPatternCache
	caseSensitive             bool
	trimTokens                bool
	cachePatterns             *bool
	tokenizedPatternCache     map[string][]string
	stringMatcherCache        map[string]*AntPathStringMatcher
}

type PathSeparatorPatternCache struct {
	endsOnWildCard       string
	endsOnDoubleWildCard string
}

type AntPathStringMatcher struct {
	rawPattern    string
	caseSensitive bool
	exactMatch    bool
	pattern       *regexp.Regexp
	variableNames []string
}

type AntPatternComparator struct {
	path string
}

type PatternInfo struct {
	pattern         string
	uriVars         int
	singleWildcards int
	doubleWildcards int
	catchAllPattern bool
	prefixPattern   bool
	length          int
}

func NewAntPathMatcher() *AntPathMatcher {
	return &AntPathMatcher{
		pathSeparator:             DEFAULT_PATH_SEPARATOR,
		pathSeparatorPatternCache: NewPathSeparatorPatternCache(DEFAULT_PATH_SEPARATOR),
		caseSensitive:             true,
		trimTokens:                false,
		tokenizedPatternCache:     make(map[string][]string),
		stringMatcherCache:        make(map[string]*AntPathStringMatcher),
	}
}

func NewAntPathMatcherWithSeparator(pathSeparator string) *AntPathMatcher {
	if pathSeparator == "" {
		pathSeparator = DEFAULT_PATH_SEPARATOR
	}
	return &AntPathMatcher{
		pathSeparator:             pathSeparator,
		pathSeparatorPatternCache: NewPathSeparatorPatternCache(pathSeparator),
		caseSensitive:             true,
		trimTokens:                false,
		tokenizedPatternCache:     make(map[string][]string),
		stringMatcherCache:        make(map[string]*AntPathStringMatcher),
	}
}

func NewPathSeparatorPatternCache(pathSeparator string) *PathSeparatorPatternCache {
	return &PathSeparatorPatternCache{
		endsOnWildCard:       pathSeparator + "*",
		endsOnDoubleWildCard: pathSeparator + "**",
	}
}

func (m *AntPathMatcher) SetPathSeparator(pathSeparator string) {
	if pathSeparator == "" {
		pathSeparator = DEFAULT_PATH_SEPARATOR
	}
	m.pathSeparator = pathSeparator
	m.pathSeparatorPatternCache = NewPathSeparatorPatternCache(pathSeparator)
}

func (m *AntPathMatcher) SetCaseSensitive(caseSensitive bool) {
	m.caseSensitive = caseSensitive
}

func (m *AntPathMatcher) SetTrimTokens(trimTokens bool) {
	m.trimTokens = trimTokens
}

func (m *AntPathMatcher) SetCachePatterns(cachePatterns bool) {
	m.cachePatterns = &cachePatterns
}

func (m *AntPathMatcher) deactivatePatternCache() {
	m.cachePatterns = new(bool)
	*m.cachePatterns = false
	m.tokenizedPatternCache = make(map[string][]string)
	m.stringMatcherCache = make(map[string]*AntPathStringMatcher)
}

func (m *AntPathMatcher) IsPattern(path string) bool {
	if path == "" {
		return false
	}
	uriVar := false
	for _, c := range path {
		switch c {
		case '*', '?':
			return true
		case '{':
			uriVar = true
		case '}':
			if uriVar {
				return true
			}
		}
	}
	return false
}

func (m *AntPathMatcher) Match(pattern string, path string) bool {
	return m.doMatch(pattern, path, true, nil)
}

func (m *AntPathMatcher) MatchStart(pattern string, path string) bool {
	return m.doMatch(pattern, path, false, nil)
}

func (m *AntPathMatcher) doMatch(pattern string, path string, fullMatch bool, uriTemplateVariables map[string]string) bool {
	if path == "" || strings.HasPrefix(path, m.pathSeparator) != strings.HasPrefix(pattern, m.pathSeparator) {
		return false
	}

	pattDirs := m.tokenizePattern(pattern)
	if fullMatch && m.caseSensitive && !m.isPotentialMatch(path, pattDirs) {
		return false
	}

	pathDirs := m.tokenizePath(path)
	pattIdxStart := 0
	pattIdxEnd := len(pattDirs) - 1
	pathIdxStart := 0
	pathIdxEnd := len(pathDirs) - 1

	// Match all elements up to the first **
	for pattIdxStart <= pattIdxEnd && pathIdxStart <= pathIdxEnd {
		pattDir := pattDirs[pattIdxStart]
		if pattDir == "**" {
			break
		}
		if !m.matchStrings(pattDir, pathDirs[pathIdxStart], uriTemplateVariables) {
			return false
		}
		pattIdxStart++
		pathIdxStart++
	}

	if pathIdxStart > pathIdxEnd {
		if pattIdxStart > pattIdxEnd {
			return strings.HasSuffix(pattern, m.pathSeparator) == strings.HasSuffix(path, m.pathSeparator)
		}
		if !fullMatch {
			return true
		}
		if pattIdxStart == pattIdxEnd && pattDirs[pattIdxStart] == "*" && strings.HasSuffix(path, m.pathSeparator) {
			return true
		}
		for i := pattIdxStart; i <= pattIdxEnd; i++ {
			if pattDirs[i] != "**" {
				return false
			}
		}
		return true
	} else if pattIdxStart > pattIdxEnd {
		return false
	} else if !fullMatch && pattDirs[pattIdxStart] == "**" {
		return true
	}

	// up to last '**'
	for pattIdxStart <= pattIdxEnd && pathIdxStart <= pathIdxEnd {
		pattDir := pattDirs[pattIdxEnd]
		if pattDir == "**" {
			break
		}
		if !m.matchStrings(pattDir, pathDirs[pathIdxEnd], uriTemplateVariables) {
			return false
		}
		pattIdxEnd--
		pathIdxEnd--
	}
	if pathIdxStart > pathIdxEnd {
		for i := pattIdxStart; i <= pattIdxEnd; i++ {
			if pattDirs[i] != "**" {
				return false
			}
		}
		return true
	}

	for pattIdxStart != pattIdxEnd && pathIdxStart <= pathIdxEnd {
		patIdxTmp := -1
		for i := pattIdxStart + 1; i <= pattIdxEnd; i++ {
			if pattDirs[i] == "**" {
				patIdxTmp = i
				break
			}
		}
		if patIdxTmp == pattIdxStart+1 {
			pattIdxStart++
			continue
		}
		patLength := patIdxTmp - pattIdxStart - 1
		strLength := pathIdxEnd - pathIdxStart + 1
		foundIdx := -1

	STR_LOOP:
		for i := 0; i <= strLength-patLength; i++ {
			for j := 0; j < patLength; j++ {
				subPat := pattDirs[pattIdxStart+j+1]
				subStr := pathDirs[pathIdxStart+i+j]
				if !m.matchStrings(subPat, subStr, uriTemplateVariables) {
					continue STR_LOOP
				}
			}
			foundIdx = pathIdxStart + i
			break
		}

		if foundIdx == -1 {
			return false
		}

		pattIdxStart = patIdxTmp
		pathIdxStart = foundIdx + patLength
	}

	for i := pattIdxStart; i <= pattIdxEnd; i++ {
		if pattDirs[i] != "**" {
			return false
		}
	}

	return true
}

func (m *AntPathMatcher) isPotentialMatch(path string, pattDirs []string) bool {
	if !m.trimTokens {
		pos := 0
		for _, pattDir := range pattDirs {
			skipped := m.skipSeparator(path, pos)
			pos += skipped
			skipped = m.skipSegment(path, pos, pattDir)
			if skipped < len(pattDir) {
				return skipped > 0 || (len(pattDir) > 0 && m.isWildcardChar(rune(pattDir[0])))
			}
			pos += skipped
		}
	}
	return true
}

func (m *AntPathMatcher) skipSegment(path string, pos int, prefix string) int {
	skipped := 0
	for i := 0; i < len(prefix); i++ {
		c := rune(prefix[i])
		if m.isWildcardChar(c) {
			return skipped
		}
		currPos := pos + skipped
		if currPos >= len(path) {
			return 0
		}
		if c == rune(path[currPos]) {
			skipped++
		}
	}
	return skipped
}

func (m *AntPathMatcher) skipSeparator(path string, pos int) int {
	skipped := 0
	for strings.HasPrefix(path[pos+skipped:], m.pathSeparator) {
		skipped += len(m.pathSeparator)
	}
	return skipped
}

func (m *AntPathMatcher) isWildcardChar(c rune) bool {
	for _, candidate := range WILDCARD_CHARS {
		if c == candidate {
			return true
		}
	}
	return false
}

func (m *AntPathMatcher) tokenizePattern(pattern string) []string {
	var tokenized []string
	if m.cachePatterns == nil || *m.cachePatterns {
		if val, ok := m.tokenizedPatternCache[pattern]; ok {
			tokenized = val
		}
	}
	if tokenized == nil {
		tokenized = m.tokenizePath(pattern)
		if m.cachePatterns == nil && len(m.tokenizedPatternCache) >= CACHE_TURNOFF_THRESHOLD {
			m.deactivatePatternCache()
			return tokenized
		}
		if m.cachePatterns == nil || *m.cachePatterns {
			m.tokenizedPatternCache[pattern] = tokenized
		}
	}
	return tokenized
}

func (m *AntPathMatcher) tokenizePath(path string) []string {
	return string_tool.TokenizeToArray(path, m.pathSeparator, m.trimTokens, true)
}

func (m *AntPathMatcher) matchStrings(pattern string, str string, uriTemplateVariables map[string]string) bool {
	return m.getStringMatcher(pattern).MatchStrings(str, uriTemplateVariables)
}

func (m *AntPathMatcher) getStringMatcher(pattern string) *AntPathStringMatcher {
	var matcher *AntPathStringMatcher
	if m.cachePatterns == nil || *m.cachePatterns {
		if val, ok := m.stringMatcherCache[pattern]; ok {
			matcher = val
		}
	}
	if matcher == nil {
		matcher = NewAntPathStringMatcher(pattern, m.caseSensitive)
		if m.cachePatterns == nil && len(m.stringMatcherCache) >= CACHE_TURNOFF_THRESHOLD {
			m.deactivatePatternCache()
			return matcher
		}
		if m.cachePatterns == nil || *m.cachePatterns {
			m.stringMatcherCache[pattern] = matcher
		}
	}
	return matcher
}

func NewAntPathStringMatcher(pattern string, caseSensitive bool) *AntPathStringMatcher {
	rawPattern := pattern
	exactMatch := false
	var compiledPattern *regexp.Regexp
	variableNames := []string{}

	patternBuilder := strings.Builder{}
	matches := GLOB_PATTERN.FindAllStringIndex(pattern, -1)
	end := 0
	if len(matches) == 0 {
		exactMatch = true
	} else {
		for _, match := range matches {
			patternBuilder.WriteString(regexp.QuoteMeta(pattern[end:match[0]]))
			matchStr := pattern[match[0]:match[1]]
			switch matchStr {
			case "?":
				patternBuilder.WriteString(".")
			case "*":
				patternBuilder.WriteString(".*")
			default:
				if strings.HasPrefix(matchStr, "{") && strings.HasSuffix(matchStr, "}") {
					colonIdx := strings.Index(matchStr, ":")
					if colonIdx == -1 {
						patternBuilder.WriteString(DEFAULT_VARIABLE_PATTERN)
						variableNames = append(variableNames, pattern[match[1]:match[1]])
					} else {
						variablePattern := matchStr[colonIdx+1 : len(matchStr)-1]
						patternBuilder.WriteString("(")
						patternBuilder.WriteString(variablePattern)
						patternBuilder.WriteString(")")
						variableName := matchStr[1:colonIdx]
						variableNames = append(variableNames, variableName)
					}
				}
			}
			end = match[1]
		}
		patternBuilder.WriteString(regexp.QuoteMeta(pattern[end:]))
		patternStr := patternBuilder.String()
		patternStr = "(?s)" + patternStr
		if !caseSensitive {
			patternStr = "(?i)" + patternStr
		}

		compiledPattern = regexp.MustCompile(patternStr)
	}

	return &AntPathStringMatcher{
		rawPattern:    rawPattern,
		caseSensitive: caseSensitive,
		exactMatch:    exactMatch,
		pattern:       compiledPattern,
		variableNames: variableNames,
	}
}

func (m *AntPathStringMatcher) MatchStrings(str string, uriTemplateVariables map[string]string) bool {
	if m.exactMatch {
		if m.caseSensitive {
			return m.rawPattern == str
		}
		return strings.EqualFold(m.rawPattern, str)
	}
	if m.pattern != nil {
		matches := m.pattern.FindStringSubmatch(str)
		if len(matches) > 0 {
			if uriTemplateVariables != nil {
				if len(m.variableNames) != len(matches)-1 {
					panic("The number of capturing groups in the pattern segment " + m.pattern.String() + " does not match the number of URI template variables it defines")
				}
				for i := 1; i <= len(matches)-1; i++ {
					name := m.variableNames[i-1]
					if strings.HasPrefix(name, "*") {
						panic("Capturing patterns (" + name + ") are not supported by the AntPathMatcher")
					}
					uriTemplateVariables[name] = matches[i]
				}
			}
			return true
		}
	}
	return false
}

func (m *AntPathMatcher) ExtractPathWithinPattern(pattern string, path string) string {
	patternParts := m.tokenizePath(pattern)
	pathParts := m.tokenizePath(path)
	var builder strings.Builder
	pathStarted := false

	for segment := 0; segment < len(patternParts); segment++ {
		patternPart := patternParts[segment]
		if strings.ContainsAny(patternPart, "*?") {
			for ; segment < len(pathParts); segment++ {
				if pathStarted || (segment == 0 && !strings.HasPrefix(pattern, m.pathSeparator)) {
					builder.WriteString(m.pathSeparator)
				}
				builder.WriteString(pathParts[segment])
				pathStarted = true
			}
		}
	}

	return builder.String()
}

func (m *AntPathMatcher) ExtractUriTemplateVariables(pattern string, path string) map[string]string {
	variables := make(map[string]string)
	result := m.doMatch(pattern, path, true, variables)
	if !result {
		panic("Pattern \"" + pattern + "\" is not a match for \"" + path + "\"")
	}
	return variables
}

func (m *AntPathMatcher) Combine(pattern1 string, pattern2 string) string {
	if string_tool.IsBlank(pattern1) && string_tool.IsBlank(pattern2) {
		return ""
	}
	if string_tool.IsBlank(pattern1) {
		return pattern2
	}
	if string_tool.IsBlank(pattern2) {
		return pattern1
	}

	pattern1ContainsUriVar := strings.Contains(pattern1, "{")
	if pattern1 != pattern2 && !pattern1ContainsUriVar && m.Match(pattern1, pattern2) {
		return pattern2
	}

	if strings.HasSuffix(pattern1, m.pathSeparatorPatternCache.endsOnWildCard) {
		return m.concat(pattern1[:len(pattern1)-2], pattern2)
	}

	if strings.HasSuffix(pattern1, m.pathSeparatorPatternCache.endsOnDoubleWildCard) {
		return m.concat(pattern1, pattern2)
	}

	starDotPos1 := strings.Index(pattern1, "*.")
	if pattern1ContainsUriVar || starDotPos1 == -1 || m.pathSeparator == "." {
		return m.concat(pattern1, pattern2)
	}

	ext1 := pattern1[starDotPos1+1:]
	dotPos2 := strings.Index(pattern2, ".")
	var file2, ext2 string
	if dotPos2 == -1 {
		file2 = pattern2
		ext2 = ""
	} else {
		file2 = pattern2[:dotPos2]
		ext2 = pattern2[dotPos2:]
	}
	ext1All := ext1 == ".*" || ext1 == ""
	ext2All := ext2 == ".*" || ext2 == ""
	if !ext1All && !ext2All {
		panic("Cannot combine patterns: " + pattern1 + " vs " + pattern2)
	}
	ext := ext1
	if ext1All {
		ext = ext2
	}
	return file2 + ext
}

func (m *AntPathMatcher) concat(path1 string, path2 string) string {
	path1EndsWithSeparator := strings.HasSuffix(path1, m.pathSeparator)
	path2StartsWithSeparator := strings.HasPrefix(path2, m.pathSeparator)

	if path1EndsWithSeparator && path2StartsWithSeparator {
		return path1 + path2[len(m.pathSeparator):]
	} else if path1EndsWithSeparator || path2StartsWithSeparator {
		return path1 + path2
	} else {
		return path1 + m.pathSeparator + path2
	}
}

func (m *AntPathMatcher) GetPatternComparator(path string) func(s1 string, s2 string) int {
	comparator := AntPatternComparator{path: path}
	return func(s1 string, s2 string) int {
		return comparator.Compare(s1, s2)
	}
}

func (c *AntPatternComparator) Compare(pattern1 string, pattern2 string) int {
	info1 := NewPatternInfo(pattern1)
	info2 := NewPatternInfo(pattern2)

	if info1.IsLeastSpecific() && info2.IsLeastSpecific() {
		return 0
	} else if info1.IsLeastSpecific() {
		return 1
	} else if info2.IsLeastSpecific() {
		return -1
	}

	pattern1EqualsPath := pattern1 == c.path
	pattern2EqualsPath := pattern2 == c.path
	if pattern1EqualsPath && pattern2EqualsPath {
		return 0
	} else if pattern1EqualsPath {
		return -1
	} else if pattern2EqualsPath {
		return 1
	}

	if info1.IsPrefixPattern() && info2.IsPrefixPattern() {
		return info2.Length() - info1.Length()
	} else if info1.IsPrefixPattern() && info2.doubleWildcards == 0 {
		return 1
	} else if info2.IsPrefixPattern() && info1.doubleWildcards == 0 {
		return -1
	}

	if info1.TotalCount() != info2.TotalCount() {
		return info1.TotalCount() - info2.TotalCount()
	}

	if info1.Length() != info2.Length() {
		return info2.Length() - info1.Length()
	}

	if info1.singleWildcards < info2.singleWildcards {
		return -1
	} else if info2.singleWildcards < info1.singleWildcards {
		return 1
	}

	if info1.uriVars < info2.uriVars {
		return -1
	} else if info2.uriVars < info1.uriVars {
		return 1
	}

	return 0
}

func NewPatternInfo(pattern string) *PatternInfo {
	info := &PatternInfo{
		pattern: pattern,
	}
	if pattern != "" {
		info.initCounters()
		info.catchAllPattern = pattern == "/**"
		info.prefixPattern = !info.catchAllPattern && strings.HasSuffix(pattern, "/**")
	}
	if info.uriVars == 0 {
		info.length = len(pattern)
	}
	return info
}

func (info *PatternInfo) initCounters() {
	pos := 0
	for pos < len(info.pattern) {
		switch info.pattern[pos] {
		case '{':
			info.uriVars++
			pos++
		case '*':
			if pos+1 < len(info.pattern) && info.pattern[pos+1] == '*' {
				info.doubleWildcards++
				pos += 2
			} else if pos > 0 && !strings.HasSuffix(info.pattern[:pos], ".*") {
				info.singleWildcards++
				pos++
			} else {
				pos++
			}
		default:
			pos++
		}
	}
}

func (info *PatternInfo) IsLeastSpecific() bool {
	return info.pattern == "" || info.catchAllPattern
}

func (info *PatternInfo) IsPrefixPattern() bool {
	return info.prefixPattern
}

func (info *PatternInfo) TotalCount() int {
	return info.uriVars + info.singleWildcards + 2*info.doubleWildcards
}

func (info *PatternInfo) Length() int {
	if info.length == 0 {
		info.length = len(VARIABLE_PATTERN.ReplaceAllString(info.pattern, "#"))
	}
	return info.length
}
