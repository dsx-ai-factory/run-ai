package settings

import (
	"regexp"
	"strings"
)

const disallowedContentMsg = "Value contains disallowed content"

func ValidateBoolean(value interface{}) string {
	if _, ok := value.(bool); !ok {
		return "Value must to be true or false"
	}
	return ""
}

func ValidateNumber(value interface{}) string {
	switch value.(type) {
	case float64, int, float32, int64:
		return ""
	default:
		return "Value must to be a number"
	}
}

func ValidateNone(_ interface{}) string {
	return ""
}

var allowedTags = map[string]bool{
	"a": true, "b": true, "i": true, "u": true, "p": true,
	"br": true, "hr": true, "ul": true, "ol": true, "li": true,
	"strong": true, "em": true,
	"h1": true, "h2": true, "h3": true, "h4": true, "h5": true, "h6": true,
	"div": true, "span": true, "blockquote": true, "pre": true, "code": true, "font": true,
}

var (
	eventHandlerPattern = regexp.MustCompile(`(?i)\bon\w+\s*=`)
	dangerousProtocol   = regexp.MustCompile(`(?i)(javascript|vbscript|data):`)
	htmlTagPattern      = regexp.MustCompile(`(?i)</?([a-zA-Z][a-zA-Z0-9]*)\b[^>]*>`)
	hrefPattern         = regexp.MustCompile(`(?i)href\s*=\s*["']([^"']*)["']`)
	safeURLPattern      = regexp.MustCompile(`(?i)^(https?://|mailto:|#|/)`)
)

func ValidateAdminMessage(value interface{}) string {
	valMap, ok := value.(map[string]interface{})
	if !ok {
		return "Value must be an object with msg, id, createdAt, enableDismiss fields"
	}

	msg, ok := valMap["msg"]
	if !ok {
		return ""
	}
	msgStr, ok := msg.(string)
	if !ok {
		return "Value must be a string"
	}

	return validateHTMLXSS(msgStr)
}

func validateHTMLXSS(val string) string {
	if eventHandlerPattern.MatchString(val) {
		return disallowedContentMsg
	}

	if dangerousProtocol.MatchString(val) {
		return disallowedContentMsg
	}

	tagMatches := htmlTagPattern.FindAllStringSubmatch(val, -1)
	for _, match := range tagMatches {
		tagName := strings.ToLower(match[1])
		if !allowedTags[tagName] {
			return disallowedContentMsg
		}
	}

	hrefMatches := hrefPattern.FindAllStringSubmatch(val, -1)
	for _, match := range hrefMatches {
		url := match[1]
		if url != "" && !safeURLPattern.MatchString(url) {
			return disallowedContentMsg
		}
	}
	return ""
}
