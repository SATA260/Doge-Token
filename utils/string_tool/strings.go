package string_tool

import "strings"

func IsBlank(s string) bool {
	return strings.TrimSpace(s) == ""
}

func TokenizeToArray(str string, delimiters string, trimTokens bool, ignoreEmptyTokens bool) []string {
	if str == "" {
		return nil
	}

	// 按分隔符分割字符串
	parts := strings.FieldsFunc(str, func(r rune) bool {
		return strings.ContainsRune(delimiters, r)
	})

	var tokens []string
	for _, part := range parts {
		token := part
		if trimTokens {
			token = strings.TrimSpace(token)
		}
		if ignoreEmptyTokens && token == "" {
			continue
		}
		tokens = append(tokens, token)
	}

	return tokens
}
