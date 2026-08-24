package ai

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// ExtractJSON pulls the first JSON object out of a provider completion,
// stripping markdown fences or surrounding prose that some models add.
func ExtractJSON(text string) (map[string]any, error) {
	start := strings.Index(text, "{")
	if start < 0 {
		return nil, errors.New("ai: no JSON object in output")
	}
	end := strings.LastIndex(text, "}")
	if end <= start {
		return nil, errors.New("ai: malformed JSON output")
	}
	var out map[string]any
	if err := json.Unmarshal([]byte(text[start:end+1]), &out); err != nil {
		return nil, fmt.Errorf("ai: invalid JSON: %w", err)
	}
	return out, nil
}

// RequireString validates that a structured output has the required string
// fields (none empty) and rejects unknown enum values.
func RequireString(v map[string]any, key string, allowed ...string) (string, error) {
	raw, ok := v[key]
	if !ok {
		return "", fmt.Errorf("ai: missing field %q", key)
	}
	s, ok := raw.(string)
	if !ok || strings.TrimSpace(s) == "" {
		return "", fmt.Errorf("ai: field %q is not a non-empty string", key)
	}
	if len(allowed) > 0 {
		found := false
		for _, a := range allowed {
			if s == a {
				found = true
				break
			}
		}
		if !found {
			return "", fmt.Errorf("ai: field %q has unknown value %q", key, s)
		}
	}
	return s, nil
}

// RequireFloat validates an optional float field.
func RequireFloat(v map[string]any, key string) (float64, error) {
	raw, ok := v[key]
	if !ok {
		return 0, nil
	}
	f, ok := raw.(float64)
	if !ok {
		return 0, fmt.Errorf("ai: field %q is not a number", key)
	}
	return f, nil
}
