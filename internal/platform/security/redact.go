package security

import "strings"

func RedactToken(s string) string {
	if s == "" {
		return ""
	}
	if len(s) <= 8 {
		return "***"
	}
	return s[:4] + "..." + s[len(s)-4:]
}

func RedactURL(raw string) string {
	if i := strings.Index(raw, "@"); i > 0 && strings.Contains(raw[:i], ":") {
		return "***@" + raw[i+1:]
	}
	return raw
}
