// pkg/converter/xml_test.go (optional)
package converter

import (
	"strings"
	"testing"
)

func TestXMLConverter_ToXML(t *testing.T) {
	converter := NewXMLConverter()

	data := map[string]interface{}{
		"name":  "John Doe",
		"email": "john@example.com",
		"age":   30,
	}

	xmlBytes, err := converter.ToXML(data)
	if err != nil {
		t.Fatalf("ToXML failed: %v", err)
	}

	xmlStr := string(xmlBytes)

	// Verify expected content
	if !strings.Contains(xmlStr, "<name>John Doe</name>") {
		t.Error("Expected name element not found")
	}
	if !strings.Contains(xmlStr, "<email>john@example.com</email>") {
		t.Error("Expected email element not found")
	}
	if !strings.Contains(xmlStr, "<?xml") {
		t.Error("Expected XML declaration not found")
	}
}

func TestXMLConverter_EscapeSpecialChars(t *testing.T) {
	converter := NewXMLConverter()

	data := map[string]interface{}{
		"description": "Hello & <world>!",
	}

	xmlBytes, err := converter.ToXML(data)
	if err != nil {
		t.Fatalf("ToXML failed: %v", err)
	}

	xmlStr := string(xmlBytes)

	// Verify escaped characters
	if strings.Contains(xmlStr, "&") && !strings.Contains(xmlStr, "&amp;") {
		t.Error("Expected & to be escaped as &amp;")
	}
	if strings.Contains(xmlStr, "<") && !strings.Contains(xmlStr, "&lt;") {
		t.Error("Expected < to be escaped as &lt;")
	}
}
