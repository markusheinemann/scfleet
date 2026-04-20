package extractor

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// coerce converts a raw string value to the target type.
// For "array", raw must be a JSON-encoded string array produced by a multiple-mode extractor.
func coerce(raw, targetType, itemType string) (any, error) {
	switch targetType {
	case "string", "":
		return raw, nil

	case "number":
		cleaned := strings.ReplaceAll(raw, ",", "")
		return strconv.ParseFloat(strings.TrimSpace(cleaned), 64)

	case "boolean":
		return parseBool(raw)

	case "array":
		var items []string
		if err := json.Unmarshal([]byte(raw), &items); err != nil {
			return nil, fmt.Errorf("coerce array: %w", err)
		}

		result := make([]any, 0, len(items))
		for _, item := range items {
			v, err := coerce(item, itemType, "")
			if err != nil {
				continue
			}
			result = append(result, v)
		}

		return result, nil

	default:
		return nil, fmt.Errorf("unknown type: %s", targetType)
	}
}

func parseBool(s string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "true", "1", "yes":
		return true, nil
	case "false", "0", "no":
		return false, nil
	}

	return false, fmt.Errorf("cannot coerce %q to boolean", s)
}
