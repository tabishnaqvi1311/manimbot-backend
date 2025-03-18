package utils

import (
	"regexp"
	"strings"
)

func ExtractCode(content string) string {
	re := regexp.MustCompile("```python\n((?s).+?)```")
	matches := re.FindStringSubmatch(content)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	re = regexp.MustCompile("```(?:python)?\n((?s).*?class.*?Scene.*?)```")
	matches = re.FindStringSubmatch(content)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	return ""
}