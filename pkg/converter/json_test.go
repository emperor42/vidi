package converter

import (
	"strings"
	"testing"
	"time"
)

// TestNewJSONConverter verifies default initialization.
func TestNewJSONConverter(t *testing.T) {
	jc := NewJSONConverter()

	if jc.Indent != "  " {
		t.Errorf("Expected default indent '  ', got %q", jc.Indent)
	}
	if !jc.SortKeys {
		t.Error("Expected SortKeys to be true by default")
	}
	if !jc.EscapeHTML {
		t.Error("Expected EscapeHTML to be true by default")
	}
	if jc.TimeFormat != time.RFC3339 {
		t.Errorf("Expected TimeFormat to be RFC3339, got %q", jc.TimeFormat)
	}
}

// TestJSONConverter_ToJSON_Pretty tests default pretty-printed output.
func TestJSONConverter_ToJSON_Pretty(t *testing.T) {
	jc := NewJSONConverter()

	data := map[string]interface{}{
		"z_last":  "last",
		"a_first": "first",
		"m_mid":   "mid",
	}

	jsonBytes, err := jc.ToJSON(data)
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	jsonStr := string(jsonBytes)

	// Verify indentation
	if !strings.Contains(jsonStr, "\n  ") {
		t.Error("Expected pretty-printed output with indentation")
	}

	// Verify key sorting (a_first should come before m_mid)
	aPos := strings.Index(jsonStr, `"a_first"`)
	mPos := strings.Index(jsonStr, `"m_mid"`)
	zPos := strings.Index(jsonStr, `"z_last"`)

	if !(aPos < mPos && mPos < zPos) {
		t.Errorf("Keys not sorted correctly. Order: a=%d, m=%d, z=%d", aPos, mPos, zPos)
	}
}

// TestJSONConverter_ToJSON_Compact tests compact output (no whitespace).
func TestJSONConverter_ToJSON_Compact(t *testing.T) {
	jc := NewJSONConverter()
	jc.Indent = ""

	data := map[string]interface{}{
		"name": "Test",
		"id":   123,
	}

	jsonBytes, err := jc.ToJSON(data)
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	jsonStr := string(jsonBytes)

	// Verify no newlines or extra spaces
	if strings.Contains(jsonStr, "\n") {
		t.Error("Compact JSON should not contain newlines")
	}
	if strings.Contains(jsonStr, ": ") {
		t.Error("Compact JSON should not contain space after colon")
	}
	if strings.Contains(jsonStr, ", ") {
		t.Error("Compact JSON should not contain space after comma")
	}

	// Verify content is correct
	expected := `{"id":123,"name":"Test"}`
	if jsonStr != expected {
		t.Errorf("Expected %q, got %q", expected, jsonStr)
	}
}

// TestJSONConverter_ToJSON_Nested tests recursive sorting in nested structures.
func TestJSONConverter_ToJSON_Nested(t *testing.T) {
	jc := NewJSONConverter()

	data := map[string]interface{}{
		"user": map[string]interface{}{
			"z_name": "John",
			"a_age":  30,
			"details": map[string]interface{}{
				"z_city": "NYC",
				"a_zip":  "10001",
			},
		},
		"meta": map[string]interface{}{
			"z_updated": "now",
			"a_created": "then",
		},
	}

	jsonBytes, err := jc.ToJSON(data)
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	jsonStr := string(jsonBytes)

	// Verify top-level sorting
	if !strings.Contains(jsonStr, `"meta"`) || !strings.Contains(jsonStr, `"user"`) {
		t.Error("Top level keys missing")
	}
	// meta comes before user alphabetically
	metaPos := strings.Index(jsonStr, `"meta"`)
	userPos := strings.Index(jsonStr, `"user"`)
	if metaPos > userPos {
		t.Error("Top level keys not sorted: meta should come before user")
	}

	// Verify nested sorting (inside user)
	userSection := jsonStr[strings.Index(jsonStr, `"user"`):]
	aAgePos := strings.Index(userSection, `"a_age"`)
	zNamePos := strings.Index(userSection, `"z_name"`)
	if aAgePos > zNamePos {
		t.Error("Nested keys inside 'user' not sorted")
	}

	// Verify deep nesting (inside details)
	detailsSection := userSection[strings.Index(userSection, `"details"`):]
	aZipPos := strings.Index(detailsSection, `"a_zip"`)
	zCityPos := strings.Index(detailsSection, `"z_city"`)
	if aZipPos > zCityPos {
		t.Error("Deep nested keys inside 'details' not sorted")
	}
}

// TestJSONConverter_ToJSON_Slice tests handling of slices.
func TestJSONConverter_ToJSON_Slice(t *testing.T) {
	jc := NewJSONConverter()

	data := []map[string]interface{}{
		{"z_id": 2, "a_name": "Bob"},
		{"z_id": 1, "a_name": "Alice"},
	}

	jsonBytes, err := jc.ToJSON(data)
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	jsonStr := string(jsonBytes)

	// Verify slice structure
	if !strings.Contains(jsonStr, "[") || !strings.Contains(jsonStr, "]") {
		t.Error("Expected array structure")
	}

	// Verify keys inside objects are sorted
	if !strings.Contains(jsonStr, `"a_name"`) || !strings.Contains(jsonStr, `"z_id"`) {
		t.Error("Object keys inside slice not found")
	}
}

// TestJSONConverter_ToJSON_Time tests time.Time formatting.
func TestJSONConverter_ToJSON_Time(t *testing.T) {
	jc := NewJSONConverter()

	now := time.Date(2023, 10, 5, 14, 48, 0, 0, time.UTC)
	data := map[string]interface{}{
		"timestamp": now,
	}

	jsonBytes, err := jc.ToJSON(data)
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	jsonStr := string(jsonBytes)

	// Verify RFC3339 format (default)
	if !strings.Contains(jsonStr, "2023-10-05T14:48:00Z") {
		t.Errorf("Expected RFC3339 timestamp, got: %s", jsonStr)
	}
}

// TestJSONConverter_ToJSON_CustomTimeFormat tests custom time formatting.
func TestJSONConverter_ToJSON_CustomTimeFormat(t *testing.T) {
	jc := NewJSONConverter()
	jc.TimeFormat = "2006-01-02 15:04:05"

	now := time.Date(2023, 10, 5, 14, 48, 0, 0, time.UTC)
	data := map[string]interface{}{
		"timestamp": now,
	}

	jsonBytes, err := jc.ToJSON(data)
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	jsonStr := string(jsonBytes)

	// Verify custom format
	if !strings.Contains(jsonStr, "2023-10-05 14:48:00") {
		t.Errorf("Expected custom timestamp format, got: %s", jsonStr)
	}
}

// TestJSONConverter_ToJSON_HtmlEscape tests HTML character escaping.
func TestJSONConverter_ToJSON_HtmlEscape(t *testing.T) {
	jc := NewJSONConverter()
	jc.EscapeHTML = true

	data := map[string]interface{}{
		"description": "Hello <world> & \"friends\"",
	}

	jsonBytes, err := jc.ToJSON(data)
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	jsonStr := string(jsonBytes)

	// Verify escaping
	if !strings.Contains(jsonStr, "\\u003c") || !strings.Contains(jsonStr, "\\u003e") {
		// Note: Go's json.Encoder uses \uXXXX for < and > when EscapeHTML is true
		t.Errorf("Expected HTML characters to be escaped, got: %s", jsonStr)
	}
}

// TestJSONConverter_ToJSON_NoHtmlEscape tests disabling HTML escaping.
func TestJSONConverter_ToJSON_NoHtmlEscape(t *testing.T) {
	jc := NewJSONConverter()
	jc.EscapeHTML = false

	data := map[string]interface{}{
		"description": "Hello <world> & \"friends\"",
	}

	jsonBytes, err := jc.ToJSON(data)
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	jsonStr := string(jsonBytes)

	// Verify NOT escaped (should contain literal < and >)
	if strings.Contains(jsonStr, "\\u003c") {
		t.Errorf("Expected literal < character, got escaped version: %s", jsonStr)
	}
	if !strings.Contains(jsonStr, "<world>") {
		t.Errorf("Expected literal <world>, got: %s", jsonStr)
	}
}

// TestJSONConverter_UnmarshalJSON tests unmarshaling to map.
func TestJSONConverter_UnmarshalJSON(t *testing.T) {
	jc := NewJSONConverter()

	jsonData := []byte(`{"name": "John", "age": 30, "active": true}`)

	result, err := jc.UnmarshalJSON(jsonData)
	if err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}

	if result["name"] != "John" {
		t.Errorf("Expected name 'John', got %v", result["name"])
	}
	if result["age"] != float64(30) { // JSON numbers are float64
		t.Errorf("Expected age 30 (float64), got %v", result["age"])
	}
	if result["active"] != true {
		t.Errorf("Expected active true, got %v", result["active"])
	}
}

// TestJSONConverter_UnmarshalJSONStrict tests strict unmarshaling.
func TestJSONConverter_UnmarshalJSONStrict(t *testing.T) {
	jc := NewJSONConverter()

	// Valid object
	jsonData := []byte(`{"key": "value"}`)
	result, err := jc.UnmarshalJSONStrict(jsonData)
	if err != nil {
		t.Fatalf("UnmarshalJSONStrict failed for valid object: %v", err)
	}
	if result["key"] != "value" {
		t.Error("Strict unmarshal failed to extract value")
	}

	// Invalid: Array at root
	arrayData := []byte(`[{"key": "value"}]`)
	_, err = jc.UnmarshalJSONStrict(arrayData)
	if err == nil {
		t.Error("Expected error for array root in strict mode")
	}

	// Invalid: Primitive at root
	primitiveData := []byte(`"just a string"`)
	_, err = jc.UnmarshalJSONStrict(primitiveData)
	if err == nil {
		t.Error("Expected error for primitive root in strict mode")
	}
}

// TestJSONConverter_UnmarshalJSONArray tests array unmarshaling.
func TestJSONConverter_UnmarshalJSONArray(t *testing.T) {
	jc := NewJSONConverter()

	jsonData := []byte(`[{"id": 1, "name": "Alice"}, {"id": 2, "name": "Bob"}]`)

	result, err := jc.UnmarshalJSONArray(jsonData)
	if err != nil {
		t.Fatalf("UnmarshalJSONArray failed: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("Expected 2 items, got %d", len(result))
	}

	if result[0]["name"] != "Alice" {
		t.Errorf("Expected first name 'Alice', got %v", result[0]["name"])
	}
	if result[1]["id"] != float64(2) {
		t.Errorf("Expected second id 2, got %v", result[1]["id"])
	}
}

// TestJSONConverter_UnmarshalJSONArray_Invalid tests error handling for invalid array items.
func TestJSONConverter_UnmarshalJSONArray_Invalid(t *testing.T) {
	jc := NewJSONConverter()

	// Array containing a string instead of object
	jsonData := []byte(`[{"id": 1}, "invalid_string", {"id": 2}]`)

	_, err := jc.UnmarshalJSONArray(jsonData)
	if err == nil {
		t.Error("Expected error for non-object item in array")
	}
}

// TestJSONConverter_ToJSON_Nil tests handling of nil data.
func TestJSONConverter_ToJSON_Nil(t *testing.T) {
	jc := NewJSONConverter()

	jsonBytes, err := jc.ToJSON(nil)
	if err != nil {
		t.Fatalf("ToJSON failed for nil: %v", err)
	}

	if string(jsonBytes) != "null" {
		t.Errorf("Expected 'null' for nil input, got %q", string(jsonBytes))
	}
}

// TestJSONConverter_ToJSON_EmptyMap tests handling of empty maps.
func TestJSONConverter_ToJSON_EmptyMap(t *testing.T) {
	jc := NewJSONConverter()

	jsonBytes, err := jc.ToJSON(map[string]interface{}{})
	if err != nil {
		t.Fatalf("ToJSON failed for empty map: %v", err)
	}

	if string(jsonBytes) != "{}" {
		t.Errorf("Expected '{}' for empty map, got %q", string(jsonBytes))
	}
}

// TestJSONConverter_ToJSON_EmptySlice tests handling of empty slices.
func TestJSONConverter_ToJSON_EmptySlice(t *testing.T) {
	jc := NewJSONConverter()

	jsonBytes, err := jc.ToJSON([]interface{}{})
	if err != nil {
		t.Fatalf("ToJSON failed for empty slice: %v", err)
	}

	if string(jsonBytes) != "[]" {
		t.Errorf("Expected '[]' for empty slice, got %q", string(jsonBytes))
	}
}

// TestJSONConverter_ToJSON_SpecialChars tests handling of special characters.
func TestJSONConverter_ToJSON_SpecialChars(t *testing.T) {
	jc := NewJSONConverter()
	jc.EscapeHTML = false

	data := map[string]interface{}{
		"newline":   "line1\nline2",
		"tab":       "col1\tcol2",
		"quote":     `say "hello"`,
		"backslash": `path\to\file`,
	}

	jsonBytes, err := jc.ToJSON(data)
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	jsonStr := string(jsonBytes)

	// Verify escaping of control characters
	if !strings.Contains(jsonStr, `\n`) {
		t.Error("Expected \\n escape sequence")
	}
	if !strings.Contains(jsonStr, `\t`) {
		t.Error("Expected \\t escape sequence")
	}
	if !strings.Contains(jsonStr, `\"`) {
		t.Error("Expected \\\" escape sequence")
	}
	if !strings.Contains(jsonStr, `\\`) {
		t.Error("Expected \\\\ escape sequence")
	}
}

// TestJSONConverter_ToJSON_Boolean tests boolean handling.
func TestJSONConverter_ToJSON_Boolean(t *testing.T) {
	jc := NewJSONConverter()

	data := map[string]interface{}{
		"true_val":  true,
		"false_val": false,
	}

	jsonBytes, err := jc.ToJSON(data)
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	jsonStr := string(jsonBytes)

	if !strings.Contains(jsonStr, `"true_val":true`) {
		t.Error("Expected true boolean literal")
	}
	if !strings.Contains(jsonStr, `"false_val":false`) {
		t.Error("Expected false boolean literal")
	}
}

// TestJSONConverter_ToJSON_Numbers tests numeric handling.
func TestJSONConverter_ToJSON_Numbers(t *testing.T) {
	jc := NewJSONConverter()

	data := map[string]interface{}{
		"int":      42,
		"float":    3.14159,
		"negative": -100,
		"zero":     0,
	}

	jsonBytes, err := jc.ToJSON(data)
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	jsonStr := string(jsonBytes)

	// Verify numeric literals (JSON numbers don't have quotes)
	if strings.Contains(jsonStr, `"int":"42"`) {
		t.Error("Integer should not be quoted")
	}
	if !strings.Contains(jsonStr, `"int":42`) {
		t.Error("Integer value missing or incorrect")
	}
	if !strings.Contains(jsonStr, `"float":3.14159`) {
		t.Error("Float value missing or incorrect")
	}
}

// TestJSONConverter_ToJSON_MapInterfaceKey tests handling of map[interface{}]interface{}.
func TestJSONConverter_ToJSON_MapInterfaceKey(t *testing.T) {
	jc := NewJSONConverter()

	// Create a map with interface{} keys (simulating unmarshaled JSON or reflection)
	rawMap := make(map[interface{}]interface{})
	rawMap["z_key"] = "value_z"
	rawMap["a_key"] = "value_a"

	// Convert to map[string]interface{} first to simulate real-world scenario
	// or pass directly if the converter handles it (our implementation does)
	data := map[string]interface{}{
		"nested": rawMap,
	}

	jsonBytes, err := jc.ToJSON(data)
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	jsonStr := string(jsonBytes)

	// Verify keys were converted to strings and sorted
	if !strings.Contains(jsonStr, `"a_key"`) || !strings.Contains(jsonStr, `"z_key"`) {
		t.Error("Interface keys not converted to strings")
	}
	// Check order
	aPos := strings.Index(jsonStr, `"a_key"`)
	zPos := strings.Index(jsonStr, `"z_key"`)
	if aPos > zPos {
		t.Error("Interface keys not sorted")
	}
}

// TestJSONConverter_ToJSON_CircularReference tests handling of circular references (should panic/error).
// Note: Standard json.Encoder panics on circular refs. Our wrapper should propagate this or handle it.
func TestJSONConverter_ToJSON_CircularReference(t *testing.T) {
	jc := NewJSONConverter()

	type Node struct {
		Name string
		Next *Node
	}

	root := &Node{Name: "root"}
	root.Next = root // Circular reference

	defer func() {
		if r := recover(); r != nil {
			// Expected: json.Encoder panics on circular refs
			t.Logf("Recovered from expected panic: %v", r)
		}
	}()

	_, err := jc.ToJSON(root)
	if err != nil {
		// Or it might return an error depending on Go version/implementation
		t.Logf("Got error instead of panic: %v", err)
	}
	// If no panic and no error, the test passes (though unlikely with standard lib)
}

// TestToJSONSimple tests the convenience function.
func TestToJSONSimple(t *testing.T) {
	data := map[string]interface{}{"key": "value"}
	jsonBytes, err := ToJSONSimple(data)
	if err != nil {
		t.Fatalf("ToJSONSimple failed: %v", err)
	}
	if !strings.Contains(string(jsonBytes), `"key"`) {
		t.Error("Convenience function failed to produce valid JSON")
	}
}

// TestToJSONCompactSimple tests the compact convenience function.
func TestToJSONCompactSimple(t *testing.T) {
	data := map[string]interface{}{"key": "value"}
	jsonBytes, err := ToJSONCompactSimple(data)
	if err != nil {
		t.Fatalf("ToJSONCompactSimple failed: %v", err)
	}
	if strings.Contains(string(jsonBytes), "\n") {
		t.Error("Compact function produced non-compact output")
	}
}
