package config

import "fmt"

func NormalizeMap(in map[string]any) map[string]any {
	out := make(map[string]any, len(in))
	for key, value := range in {
		out[key] = normalizeAny(value)
	}
	return out
}

func normalizeAny(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return NormalizeMap(typed)
	case map[any]any:
		out := make(map[string]any, len(typed))
		for key, item := range typed {
			out[fmt.Sprint(key)] = normalizeAny(item)
		}
		return out
	case []any:
		out := make([]any, len(typed))
		for i, item := range typed {
			out[i] = normalizeAny(item)
		}
		return out
	default:
		return value
	}
}

func DeepCopyMap(in map[string]any) map[string]any {
	out := make(map[string]any, len(in))
	for key, value := range in {
		out[key] = deepCopyAny(value)
	}
	return out
}

func deepCopyAny(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return DeepCopyMap(typed)
	case []any:
		out := make([]any, len(typed))
		for i, item := range typed {
			out[i] = deepCopyAny(item)
		}
		return out
	default:
		return typed
	}
}

func MergeMap(base, overlay map[string]any) map[string]any {
	for key, value := range overlay {
		if value == nil {
			delete(base, key)
			continue
		}

		baseMap, baseIsMap := base[key].(map[string]any)
		overlayMap, overlayIsMap := value.(map[string]any)

		if baseIsMap && overlayIsMap {
			base[key] = MergeMap(baseMap, overlayMap)
			continue
		}

		base[key] = deepCopyAny(value)
	}
	return base
}

func SetNestedValue(root map[string]any, path []string, value any) {
	if len(path) == 0 {
		return
	}
	cursor := root
	for i := 0; i < len(path)-1; i++ {
		next, ok := cursor[path[i]].(map[string]any)
		if !ok {
			next = map[string]any{}
			cursor[path[i]] = next
		}
		cursor = next
	}
	cursor[path[len(path)-1]] = value
}
