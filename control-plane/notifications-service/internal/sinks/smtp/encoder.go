package smtp

import (
	"bytes"
	"mime/quotedprintable"
	"regexp"
	"strings"
)

type BodyEncoder struct{}

func (be *BodyEncoder) Encode(body string) string {
	switch be.getEncodingType(body) {
	case "text/html":
		body = be.encodeText(body)
	case "quoted-printable":
		body = be.encodeQuatedprintable(body)
	}
	return be.addDotTransparency(body)
}

func (be *BodyEncoder) encodeText(body string) string {
	var result strings.Builder
	maxLen := 78 // Recommended maximum line length per RFC 5322
	for len(body) > 0 {
		var lineEnd int
		if nextLineIndex := strings.Index(body, "\n"); nextLineIndex != -1 && nextLineIndex <= maxLen {
			lineEnd = nextLineIndex
		} else if len(body) <= maxLen {
			lineEnd = len(body)
		} else {
			lineEnd = be.wordSplitPoint(body, maxLen)
		}
		line := body[:lineEnd]
		result.WriteString(line)
		result.WriteString("\r\n") // RFC mandates CRLF line ending
		body = body[lineEnd:]
		if len(body) > 0 && body[0] == '\n' {
			body = body[1:]
		} else if len(body) > 0 && body[0] == ' ' {
			body = body[1:]
		}
	}
	return result.String()
}

func (be *BodyEncoder) wordSplitPoint(s string, maxLen int) int {
	if i := strings.LastIndex(s[:maxLen], " "); i > 0 {
		return i
	}
	return maxLen
}

func (be *BodyEncoder) encodeQuatedprintable(body string) string {
	var buf bytes.Buffer
	writer := quotedprintable.NewWriter(&buf)
	writer.Write([]byte(body))
	writer.Close()
	return buf.String()
}

func (be *BodyEncoder) addDotTransparency(body string) string {
	var result strings.Builder
	lines := strings.Split(body, "\r\n") // Split at RFC-compliant line break
	for _, line := range lines {
		if strings.HasPrefix(line, ".") {
			result.WriteString(".")
		}
		result.WriteString(line)
		result.WriteString("\r\n")
	}
	return result.String()
}

func (be *BodyEncoder) getEncodingType(body string) string {
	reEncoding := regexp.MustCompile(`(?mi)Content-Transfer-Encoding:\s*([^\r\n]+)`)
	reType := regexp.MustCompile(`(?mi)Content-Type:\s*([^;\r\n]+)`)

	encMatch := reEncoding.FindStringSubmatch(body)
	typeMatch := reType.FindStringSubmatch(body)

	if len(encMatch) > 1 && strings.TrimSpace(encMatch[1]) != "" {
		return strings.TrimSpace(encMatch[1])
	}
	if len(typeMatch) > 1 && strings.TrimSpace(typeMatch[1]) != "" {
		return strings.TrimSpace(typeMatch[1])
	}
	return ""
}
