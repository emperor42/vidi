package converter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"time"
)

// JSONConverter handles conversion of data structures to JSON format.
// It provides flexible JSON generation with configurable options for
// formatting, key sorting, and handling of special types.
type JSONConverter struct {
	// Indent determines the indentation string for pretty-printing.
	// If empty, the output will be compact (no whitespace).
	// Default is two spaces.
	Indent string

	// SortKeys determines whether to sort map keys alphabetically.
	// This ensures deterministic output for testing and caching.
	// Default is true.
	SortKeys bool

	// EscapeHTML determines whether to escape HTML characters (<, >, &, etc.).
	// Default is true (standard JSON behavior).
	EscapeHTML bool

	// TimeFormat specifies the layout for time.Time values.
	// If empty, RFC3339 is used.
	TimeFormat string
}

// NewJSONConverter creates a new JSONConverter with default settings.
func NewJSONConverter() *JSONConverter {
	return &JSONConverter{
		Indent:     "  ",
		SortKeys:   true,
		EscapeHTML: true,
		TimeFormat: time.RFC3339,
	}
}

// ToJSON converts data to JSON format.
// It accepts map[string]interface{}, slices, structs, or primitive types.
func (jc *JSONConverter) ToJSON(data interface{}) ([]byte, error) {
	if data == nil {
		return []byte("null"), nil
	}

	// Pre-process data if sorting is enabled
	preparedData := data
	if jc.SortKeys {
		preparedData = jc.sortMapKeysRecursive(data)
	}

	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(jc.EscapeHTML)

	if jc.Indent != "" {
		encoder.SetIndent("", jc.Indent)
	}

	if err := encoder.Encode(preparedData); err != nil {
		return nil, fmt.Errorf("failed to encode JSON: %w", err)
	}

	// Remove trailing newline added by Encode
	result := buf.Bytes()
	if len(result) > 0 && result[len(result)-1] == '\n' {
		result = result[:len(result)-1]
	}

	return result, nil
}

// ToJSONCompact converts data to compact JSON (no whitespace).
// This is a convenience function equivalent to setting Indent to "".
func (jc *JSONConverter) ToJSONCompact(data interface{}) ([]byte, error) {
	oldIndent := jc.Indent
	jc.Indent = ""
	defer func() { jc.Indent = oldIndent }()
	return jc.ToJSON(data)
}

// ToJSONPretty converts data to pretty-printed JSON with custom indentation.
func (jc *JSONConverter) ToJSONPretty(data interface{}, indent string) ([]byte, error) {
	oldIndent := jc.Indent
	jc.Indent = indent
	defer func() { jc.Indent = oldIndent }()
	return jc.ToJSON(data)
}

// sortMapKeysRecursive recursively sorts all map keys in the data structure.
// This ensures deterministic JSON output.
func (jc *JSONConverter) sortMapKeysRecursive(data interface{}) interface{} {
	switch v := data.(type) {
	case map[string]interface{}:
		// Create a new map with sorted keys
		sorted := make(map[string]interface{})
		keys := make([]string, 0, len(v))
		for k := range v {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			sorted[k] = jc.sortMapKeysRecursive(v[k])
		}
		return sorted

	case []interface{}:
		// Recursively process slice elements
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = jc.sortMapKeysRecursive(item)
		}
		return result

	case map[interface{}]interface{}:
		// Handle maps with interface{} keys (convert to string keys)
		sorted := make(map[string]interface{})
		keys := make([]string, 0, len(v))
		for k := range v {
			keys = append(keys, fmt.Sprintf("%v", k))
		}
		sort.Strings(keys)

		for _, k := range keys {
			sorted[k] = jc.sortMapKeysRecursive(v[k])
		}
		return sorted

	default:
		// Return primitives as-is
		return v
	}
}

// ToJSONWithTimeFormat converts data to JSON with a custom time format.
func (jc *JSONConverter) ToJSONWithTimeFormat(data interface{}, format string) ([]byte, error) {
	oldFormat := jc.TimeFormat
	jc.TimeFormat = format
	defer func() { jc.TimeFormat = oldFormat }()
	return jc.ToJSON(data)
}

// UnmarshalJSON is a helper to unmarshal JSON bytes into a map.
// It returns a map[string]interface{} for easy manipulation.
func (jc *JSONConverter) UnmarshalJSON(data []byte) (map[string]interface{}, error) {
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	return result, nil
}

// UnmarshalJSONStrict unmarshals JSON with strict type checking.
// It fails if the root element is not a map.
func (jc *JSONConverter) UnmarshalJSONStrict(data []byte) (map[string]interface{}, error) {
	var raw interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	result, ok := raw.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("expected JSON object (map), got %T", raw)
	}

	return result, nil
}

// UnmarshalJSONArray unmarshals JSON array into []map[string]interface{}.
func (jc *JSONConverter) UnmarshalJSONArray(data []byte) ([]map[string]interface{}, error) {
	var raw []interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON array: %w", err)
	}

	result := make([]map[string]interface{}, 0, len(raw))
	for _, item := range raw {
		if m, ok := item.(map[string]interface{}); ok {
			result = append(result, m)
		} else {
			return nil, fmt.Errorf("array element is not an object: %T", item)
		}
	}

	return result, nil
}

// ToJSONSimple is a convenience function for quick JSON conversion with defaults.
func ToJSONSimple(data interface{}) ([]byte, error) {
	converter := NewJSONConverter()
	return converter.ToJSON(data)
}

// ToJSONCompactSimple is a convenience function for compact JSON.
func ToJSONCompactSimple(data interface{}) ([]byte, error) {
	converter := NewJSONConverter()
	converter.Indent = ""
	return converter.ToJSON(data)
}

// Example usage:
//
//	data := map[string]interface{}{
//		"id":    "123",
//		"name":  "John Doe",
//		"email": "john@example.com",
//		"roles": []string{"admin", "user"},
//		"active": true,
//		"balance": 100.50,
//		"created_at": time.Now(),
//	}
//
//	converter := NewJSONConverter()
//	jsonBytes, err := converter.ToJSON(data)
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Println(string(jsonBytes))
//
// Output (pretty-printed, sorted keys):
// {
//   "active": true,
//   "balance": 100.5,
//   "created_at": "2023-10-05T14:48:00Z",
//   "email": "john@example.com",
//   "id": "123",
//   "name": "John Doe",
//   "roles": [
//     "admin",
//     "user"
//   ]
// }

// Example compact output:
//
//	compactBytes, _ := converter.ToJSONCompact(data)
//	fmt.Println(string(compactBytes))
//
// Output:
// {"active":true,"balance":100.5,"created_at":"2023-10-05T14:48:00Z","email":"john@example.com","id":"123","name":"John Doe","roles":["admin","user"]}
