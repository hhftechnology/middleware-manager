package config

import "strings"

func MapFromAny(input any) map[string]any {
	if input == nil {
		return map[string]any{}
	}
	switch value := input.(type) {
	case map[string]any:
		return value
	case map[any]any:
		out := make(map[string]any, len(value))
		for key, item := range value {
			if text, ok := key.(string); ok {
				out[text] = NormalizeYAML(item)
			}
		}
		return out
	default:
		return map[string]any{}
	}
}

func NormalizeYAML(input any) any {
	switch value := input.(type) {
	case map[string]any:
		out := make(map[string]any, len(value))
		for key, item := range value {
			out[key] = NormalizeYAML(item)
		}
		return out
	case map[any]any:
		return MapFromAny(value)
	case []any:
		out := make([]any, 0, len(value))
		for _, item := range value {
			out = append(out, NormalizeYAML(item))
		}
		return out
	default:
		return input
	}
}

func StringFromAny(input any) string {
	value, ok := input.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(value)
}

func StringSliceFromAny(input any) []string {
	switch value := input.(type) {
	case []string:
		out := make([]string, 0, len(value))
		for _, item := range value {
			if trimmed := strings.TrimSpace(item); trimmed != "" {
				out = append(out, trimmed)
			}
		}
		return out
	case []any:
		out := make([]string, 0, len(value))
		for _, item := range value {
			if text := StringFromAny(item); text != "" {
				out = append(out, text)
			}
		}
		return out
	default:
		return nil
	}
}

func BoolFromAny(input any) bool {
	value, ok := input.(bool)
	return ok && value
}

func RouteNameFromID(routeID string) string {
	if strings.Contains(routeID, "::") {
		parts := strings.SplitN(routeID, "::", 2)
		return parts[1]
	}
	return routeID
}
