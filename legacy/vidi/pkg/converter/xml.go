package converter

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"sort"
	"strings"
)

// XMLConverter handles conversion of data structures to XML format.
// It provides flexible XML generation with configurable options.
type XMLConverter struct {
	// RootElement is the name of the root XML element.
	// Default is "root".
	RootElement string

	// IncludeXMLDeclaration determines whether to include the XML declaration.
	// Default is true.
	IncludeXMLDeclaration bool

	// Indent determines the indentation string for pretty-printing.
	// Default is two spaces.
	Indent string

	// EscapeSpecialChars determines whether to escape special XML characters.
	// Default is true.
	EscapeSpecialChars bool
}

// NewXMLConverter creates a new XMLConverter with default settings.
func NewXMLConverter() *XMLConverter {
	return &XMLConverter{
		RootElement:           "root",
		IncludeXMLDeclaration: true,
		Indent:                "  ",
		EscapeSpecialChars:    true,
	}
}

// ToXML converts data to XML format.
// It accepts map[string]interface{}, []map[string]interface{}, or any type
// that encoding/xml can handle.
func (xc *XMLConverter) ToXML(data interface{}) ([]byte, error) {
	if data == nil {
		return []byte(fmt.Sprintf("<%s/>", xc.RootElement)), nil
	}

	var buf bytes.Buffer

	// Add XML declaration if configured
	if xc.IncludeXMLDeclaration {
		buf.WriteString(xml.Header)
	}

	// Handle different data types
	switch v := data.(type) {
	case map[string]interface{}:
		if err := xc.writeMap(&buf, v, xc.RootElement, 0); err != nil {
			return nil, err
		}
	case []map[string]interface{}:
		if err := xc.writeArray(&buf, v, xc.RootElement, 0); err != nil {
			return nil, err
		}
	case []interface{}:
		if err := xc.writeSlice(&buf, v, xc.RootElement, 0); err != nil {
			return nil, err
		}
	default:
		// Fall back to standard xml.Marshal for other types
		if err := xml.NewEncoder(&buf).Encode(data); err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}

// writeMap writes a map as XML elements.
func (xc *XMLConverter) writeMap(buf *bytes.Buffer, data map[string]interface{}, parentTag string, indentLevel int) error {
	indent := strings.Repeat(xc.Indent, indentLevel)
	nextIndent := strings.Repeat(xc.Indent, indentLevel+1)

	// Open parent tag
	buf.WriteString(fmt.Sprintf("%s<%s>", indent, parentTag))

	// Check if map is empty
	if len(data) == 0 {
		buf.WriteString("/>")
		return nil
	}

	buf.WriteString("\n")

	// Sort keys for consistent output
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Write each key-value pair
	for _, key := range keys {
		value := data[key]
		safeKey := sanitizeXMLName(key)

		if err := xc.writeValue(buf, safeKey, value, nextIndent); err != nil {
			return err
		}
	}

	// Close parent tag
	buf.WriteString(fmt.Sprintf("%s</%s>\n", indent, parentTag))
	return nil
}

// writeArray writes an array of maps as XML elements.
func (xc *XMLConverter) writeArray(buf *bytes.Buffer, data []map[string]interface{}, parentTag string, indentLevel int) error {
	indent := strings.Repeat(xc.Indent, indentLevel)
	nextIndent := strings.Repeat(xc.Indent, indentLevel+1)

	// Determine item tag name (singular form of parent)
	itemTag := strings.TrimSuffix(parentTag, "s")
	if itemTag == parentTag {
		itemTag = "item"
	}

	// Open parent tag
	buf.WriteString(fmt.Sprintf("%s<%s>\n", indent, parentTag))

	// Write each item
	for _, item := range data {
		if err := xc.writeMap(buf, item, itemTag, indentLevel+1); err != nil {
			return err
		}
	}

	// Close parent tag
	buf.WriteString(fmt.Sprintf("%s</%s>\n", indent, parentTag))
	return nil
}

// writeSlice writes a slice of interface{} as XML elements.
func (xc *XMLConverter) writeSlice(buf *bytes.Buffer, data []interface{}, parentTag string, indentLevel int) error {
	indent := strings.Repeat(xc.Indent, indentLevel)
	nextIndent := strings.Repeat(xc.Indent, indentLevel+1)

	itemTag := strings.TrimSuffix(parentTag, "s")
	if itemTag == parentTag {
		itemTag = "item"
	}

	buf.WriteString(fmt.Sprintf("%s<%s>\n", indent, parentTag))

	for _, item := range data {
		if err := xc.writeValue(buf, itemTag, item, nextIndent); err != nil {
			return err
		}
	}

	buf.WriteString(fmt.Sprintf("%s</%s>\n", indent, parentTag))
	return nil
}

// writeValue writes a single key-value pair as XML.
func (xc *XMLConverter) writeValue(buf *bytes.Buffer, key string, value interface{}, indent string) error {
	// Handle nil values
	if value == nil {
		buf.WriteString(fmt.Sprintf("%s<%s/>\n", indent, key))
		return nil
	}

	// Convert value to string representation
	valueStr := xc.valueToString(value)

	// Escape special characters if configured
	if xc.EscapeSpecialChars {
		valueStr = xmlEscape(valueStr)
	}

	// Check if value is empty
	if valueStr == "" {
		buf.WriteString(fmt.Sprintf("%s<%s/>\n", indent, key))
		return nil
	}

	// Write element with content
	buf.WriteString(fmt.Sprintf("%s<%s>%s</%s>\n", indent, key, valueStr, key))
	return nil
}

// valueToString converts a value to its string representation.
func (xc *XMLConverter) valueToString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case bool:
		return fmt.Sprintf("%t", v)
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", v)
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v)
	case float32, float64:
		return fmt.Sprintf("%g", v)
	case []interface{}:
		// Convert slice to comma-separated string
		parts := make([]string, len(v))
		for i, item := range v {
			parts[i] = xc.valueToString(item)
		}
		return strings.Join(parts, ", ")
	case map[string]interface{}:
		// Nested maps are converted to JSON string for XML compatibility
		jsonBytes, _ := New().ToJSON(v)
		return string(jsonBytes)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// sanitizeXMLName ensures the key is a valid XML element name.
func sanitizeXMLName(name string) string {
	if name == "" {
		return "empty_key"
	}

	// Replace invalid characters with underscores
	var sb strings.Builder
	for i, r := range name {
		if i == 0 {
			// First character must be letter or underscore
			if !isXMLNameStartChar(r) {
				sb.WriteRune('_')
			} else {
				sb.WriteRune(r)
			}
		} else {
			// Subsequent characters can be letters, digits, hyphens, underscores, periods
			if isXMLNameChar(r) {
				sb.WriteRune(r)
			} else {
				sb.WriteRune('_')
			}
		}
	}

	return sb.String()
}

// isXMLNameStartChar checks if a character can start an XML name.
func isXMLNameStartChar(r rune) bool {
	return (r >= 'a' && r <= 'z') ||
		(r >= 'A' && r <= 'Z') ||
		r == '_' ||
		r == ':'
}

// isXMLNameChar checks if a character can be in an XML name.
func isXMLNameChar(r rune) bool {
	return isXMLNameStartChar(r) ||
		(r >= '0' && r <= '9') ||
		r == '-' ||
		r == '.'
}

// xmlEscape escapes special XML characters in a string.
func xmlEscape(s string) string {
	var buf bytes.Buffer
	for _, r := range s {
		switch r {
		case '&':
			buf.WriteString("&amp;")
		case '<':
			buf.WriteString("&lt;")
		case '>':
			buf.WriteString("&gt;")
		case '"':
			buf.WriteString("&quot;")
		case '\'':
			buf.WriteString("&apos;")
		default:
			buf.WriteRune(r)
		}
	}
	return buf.String()
}

// ToXMLWithConfig converts data to XML with custom configuration.
// This is a convenience function for one-off conversions.
func ToXMLWithConfig(data interface{}, rootElement string, includeDeclaration bool) ([]byte, error) {
	converter := &XMLConverter{
		RootElement:           rootElement,
		IncludeXMLDeclaration: includeDeclaration,
		Indent:                "  ",
		EscapeSpecialChars:    true,
	}
	return converter.ToXML(data)
}

// ToXMLSimple is a convenience function for quick XML conversion with defaults.
func ToXMLSimple(data interface{}) ([]byte, error) {
	converter := NewXMLConverter()
	return converter.ToXML(data)
}

// Example usage:
//
//	data := map[string]interface{}{
//		"id":    "123",
//		"name":  "John Doe",
//		"email": "john@example.com",
//		"roles": []string{"admin", "user"},
//	}
//
//	converter := NewXMLConverter()
//	xmlBytes, err := converter.ToXML(data)
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Println(string(xmlBytes))
//
// Output:
// <?xml version="1.0" encoding="UTF-8"?>
// <root>
//   <email>john@example.com</email>
//   <id>123</id>
//   <name>John Doe</name>
//   <roles>admin, user</roles>
// </root>

// ToXMLPretty converts data to pretty-printed XML with custom indentation.
func (xc *XMLConverter) ToXMLPretty(data interface{}, indent string) ([]byte, error) {
	oldIndent := xc.Indent
	xc.Indent = indent
	defer func() { xc.Indent = oldIndent }()
	return xc.ToXML(data)
}

// ToXMLCompact converts data to compact XML (no indentation).
func (xc *XMLConverter) ToXMLCompact(data interface{}) ([]byte, error) {
	oldIndent := xc.Indent
	xc.Indent = ""
	defer func() { xc.Indent = oldIndent }()
	return xc.ToXML(data)
}

// FromXML parses XML data back into a map[string]interface{}.
// Note: This is a basic implementation and may not handle all XML features.
func (xc *XMLConverter) FromXML(xmlData []byte) (map[string]interface{}, error) {
	var result map[string]interface{}
	decoder := xml.NewDecoder(bytes.NewReader(xmlData))

	// Basic parsing implementation
	// For production use, consider using a more robust XML parser
	err := decoder.Decode(&result)
	if err != nil {
		return nil, fmt.Errorf("failed to parse XML: %w", err)
	}

	return result, nil
}
