package extractor

import (
	"regexp"
	"strings"
)

// applyTransform runs the optional transform pipeline on a raw string value.
// Trimming is applied by default unless explicitly disabled.
func applyTransform(raw string, t *Transform) string {
	if t == nil {
		return strings.TrimSpace(raw)
	}

	result := raw

	if t.Trim == nil || *t.Trim {
		result = strings.TrimSpace(result)
	}

	if t.Regex != "" {
		re, err := regexp.Compile(t.Regex)
		if err != nil {
			return ""
		}

		match := re.FindStringSubmatch(result)
		if match == nil {
			return ""
		}

		if len(match) > 1 {
			result = match[1]
		} else {
			result = match[0]
		}
	}

	result = t.Prepend + result + t.Append

	return result
}
