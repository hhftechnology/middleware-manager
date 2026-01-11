package services

import (
	"encoding/json"
	"fmt"
)

// DecodeArrayOrMap decodes JSON data that can be either an array or a map
// This is a common pattern in Traefik API responses where the same endpoint
// can return data in different formats depending on the provider
//
// Type parameters:
// - T: The type of items in the array/map values
//
// Returns a slice of items with their names populated from map keys if applicable
func DecodeArrayOrMap[T any](data []byte, setName func(*T, string)) ([]T, error) {
	// Try array first (more common in newer Traefik versions)
	var items []T
	if err := json.Unmarshal(data, &items); err == nil {
		return items, nil
	}

	// Try map format (common in older versions or certain providers)
	var itemsMap map[string]T
	if err := json.Unmarshal(data, &itemsMap); err != nil {
		return nil, fmt.Errorf("failed to parse as array or map: %w", err)
	}

	// Convert map to slice, setting names from keys
	items = make([]T, 0, len(itemsMap))
	for name, item := range itemsMap {
		setName(&item, name)
		items = append(items, item)
	}

	return items, nil
}

// DecodeJSONSafe decodes JSON with better error handling
// Returns the decoded value and any error that occurred
func DecodeJSONSafe[T any](data []byte) (T, error) {
	var result T
	if err := json.Unmarshal(data, &result); err != nil {
		return result, fmt.Errorf("JSON decode error: %w", err)
	}
	return result, nil
}

// MarshalJSONSafe marshals to JSON with error handling
func MarshalJSONSafe(v interface{}) ([]byte, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("JSON encode error: %w", err)
	}
	return data, nil
}

// ConvertViaJSON converts one struct type to another via JSON marshaling
// This is useful for converting between similar structs with different types
func ConvertViaJSON[T any](source interface{}) (*T, error) {
	if source == nil {
		return nil, nil
	}

	jsonData, err := json.Marshal(source)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal source: %w", err)
	}

	var result T
	if err := json.Unmarshal(jsonData, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal to target: %w", err)
	}

	return &result, nil
}

// TryConvertViaJSON is like ConvertViaJSON but returns nil on error instead of an error
// This is useful when conversion failure is acceptable
func TryConvertViaJSON[T any](source interface{}) *T {
	result, err := ConvertViaJSON[T](source)
	if err != nil {
		return nil
	}
	return result
}
