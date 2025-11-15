package utils

import (
	"regexp"
	"strings"
)

func ExtractCode(content string) string {

	re := regexp.MustCompile("```python\n((?s).+?)```")
	matches := re.FindStringSubmatch(content)
	if len(matches) > 1 {
		code := strings.TrimSpace(matches[1])
		if strings.Contains(code, "class Scene") {
			return code
		}
	}

	re = regexp.MustCompile("```(?:python)?\n((?s).+?)```")
	matches = re.FindStringSubmatch(content)
	if len(matches) > 1 {
		code := strings.TrimSpace(matches[1])
		if strings.Contains(code, "class Scene") {
			return code
		}
	}

	if strings.Contains(content, "from manim import") && strings.Contains(content, "class Scene") {
		return strings.TrimSpace(content)
	}

	return ""
}
