package services

import (
	"testing"
)

// TestDecodeArrayOrMap tests the generic JSON decoder that handles both array and map formats
func TestDecodeArrayOrMap(t *testing.T) {
	type TestItem struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	setName := func(item *TestItem, name string) {
		item.Name = name
	}

	tests := []struct {
		name      string
		data      string
		wantLen   int
		wantErr   bool
		checkFunc func(t *testing.T, items []TestItem)
	}{
		{
			name:    "array format",
			data:    `[{"name": "item1", "value": 1}, {"name": "item2", "value": 2}]`,
			wantLen: 2,
			wantErr: false,
			checkFunc: func(t *testing.T, items []TestItem) {
				if items[0].Name != "item1" || items[0].Value != 1 {
					t.Errorf("first item = %+v, want {Name: item1, Value: 1}", items[0])
				}
				if items[1].Name != "item2" || items[1].Value != 2 {
					t.Errorf("second item = %+v, want {Name: item2, Value: 2}", items[1])
				}
			},
		},
		{
			name:    "map format",
			data:    `{"key1": {"value": 10}, "key2": {"value": 20}}`,
			wantLen: 2,
			wantErr: false,
			checkFunc: func(t *testing.T, items []TestItem) {
				// Map order is not guaranteed, so check both items exist
				found := make(map[string]bool)
				for _, item := range items {
					found[item.Name] = true
					if item.Name == "key1" && item.Value != 10 {
						t.Errorf("key1 value = %d, want 10", item.Value)
					}
					if item.Name == "key2" && item.Value != 20 {
						t.Errorf("key2 value = %d, want 20", item.Value)
					}
				}
				if !found["key1"] || !found["key2"] {
					t.Errorf("missing expected keys: found = %v", found)
				}
			},
		},
		{
			name:    "empty array",
			data:    `[]`,
			wantLen: 0,
			wantErr: false,
		},
		{
			name:    "empty map",
			data:    `{}`,
			wantLen: 0,
			wantErr: false,
		},
		{
			name:    "invalid JSON",
			data:    `{invalid}`,
			wantErr: true,
		},
		{
			name:    "null value",
			data:    `null`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items, err := DecodeArrayOrMap([]byte(tt.data), setName)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecodeArrayOrMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if len(items) != tt.wantLen {
				t.Errorf("DecodeArrayOrMap() len = %d, want %d", len(items), tt.wantLen)
			}
			if tt.checkFunc != nil {
				tt.checkFunc(t, items)
			}
		})
	}
}

// TestDecodeJSONSafe tests safe JSON decoding with proper error handling
func TestDecodeJSONSafe(t *testing.T) {
	type Config struct {
		Host string `json:"host"`
		Port int    `json:"port"`
	}

	tests := []struct {
		name    string
		data    string
		wantErr bool
		check   func(t *testing.T, result Config)
	}{
		{
			name:    "valid object",
			data:    `{"host": "localhost", "port": 8080}`,
			wantErr: false,
			check: func(t *testing.T, result Config) {
				if result.Host != "localhost" {
					t.Errorf("Host = %q, want %q", result.Host, "localhost")
				}
				if result.Port != 8080 {
					t.Errorf("Port = %d, want %d", result.Port, 8080)
				}
			},
		},
		{
			name:    "partial object uses zero values",
			data:    `{"host": "example.com"}`,
			wantErr: false,
			check: func(t *testing.T, result Config) {
				if result.Host != "example.com" {
					t.Errorf("Host = %q, want %q", result.Host, "example.com")
				}
				if result.Port != 0 {
					t.Errorf("Port = %d, want 0", result.Port)
				}
			},
		},
		{
			name:    "empty object",
			data:    `{}`,
			wantErr: false,
		},
		{
			name:    "invalid JSON",
			data:    `{"host": }`,
			wantErr: true,
		},
		{
			name:    "type mismatch",
			data:    `{"host": 123, "port": "not-a-number"}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := DecodeJSONSafe[Config]([]byte(tt.data))
			if (err != nil) != tt.wantErr {
				t.Errorf("DecodeJSONSafe() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && tt.check != nil {
				tt.check(t, result)
			}
		})
	}
}

// TestMarshalJSONSafe tests safe JSON marshaling
func TestMarshalJSONSafe(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		want    string
		wantErr bool
	}{
		{
			name:  "simple struct",
			input: struct{ Name string }{Name: "test"},
			want:  `{"Name":"test"}`,
		},
		{
			name:  "map",
			input: map[string]int{"a": 1, "b": 2},
			want:  "", // Not checking exact output due to map ordering
		},
		{
			name:  "slice",
			input: []string{"a", "b", "c"},
			want:  `["a","b","c"]`,
		},
		{
			name:  "nil",
			input: nil,
			want:  "null",
		},
		{
			name:    "unmarshalable value (channel)",
			input:   make(chan int),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MarshalJSONSafe(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalJSONSafe() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && tt.want != "" && string(got) != tt.want {
				t.Errorf("MarshalJSONSafe() = %s, want %s", got, tt.want)
			}
		})
	}
}

// TestConvertViaJSON tests struct conversion via JSON marshaling
func TestConvertViaJSON(t *testing.T) {
	type SourceType struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	type TargetType struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	type DifferentTarget struct {
		Name    string `json:"name"`
		Counter int    `json:"value"` // Different field name, same json tag
	}

	tests := []struct {
		name    string
		source  interface{}
		wantNil bool
		wantErr bool
		check   func(t *testing.T, result *TargetType)
	}{
		{
			name:   "matching struct types",
			source: SourceType{Name: "test", Value: 42},
			check: func(t *testing.T, result *TargetType) {
				if result.Name != "test" || result.Value != 42 {
					t.Errorf("result = %+v, want {Name: test, Value: 42}", result)
				}
			},
		},
		{
			name:   "map to struct",
			source: map[string]interface{}{"name": "fromMap", "value": 100},
			check: func(t *testing.T, result *TargetType) {
				if result.Name != "fromMap" {
					t.Errorf("Name = %q, want %q", result.Name, "fromMap")
				}
			},
		},
		{
			name:    "nil source",
			source:  nil,
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ConvertViaJSON[TargetType](tt.source)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertViaJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantNil && result != nil {
				t.Errorf("ConvertViaJSON() = %+v, want nil", result)
				return
			}
			if !tt.wantNil && tt.check != nil {
				tt.check(t, result)
			}
		})
	}
}

// TestTryConvertViaJSON tests the nil-returning variant of ConvertViaJSON
func TestTryConvertViaJSON(t *testing.T) {
	type Target struct {
		Name string `json:"name"`
	}

	tests := []struct {
		name    string
		source  interface{}
		wantNil bool
	}{
		{
			name:   "valid conversion",
			source: map[string]string{"name": "test"},
		},
		{
			name:    "nil source",
			source:  nil,
			wantNil: true,
		},
		{
			name:    "unconvertible type",
			source:  make(chan int),
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TryConvertViaJSON[Target](tt.source)
			if tt.wantNil && result != nil {
				t.Errorf("TryConvertViaJSON() = %+v, want nil", result)
			}
			if !tt.wantNil && result == nil {
				t.Errorf("TryConvertViaJSON() = nil, want non-nil")
			}
		})
	}
}
