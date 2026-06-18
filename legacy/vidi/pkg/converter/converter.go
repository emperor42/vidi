package converter

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strings"
)

// Converter handles data transformation to various formats.
type Converter struct{}

func New() *Converter {
	return &Converter{}
}

// ToJSON marshals data to pretty-printed JSON.
func (c *Converter) ToJSON(data interface{}) ([]byte, error) {
	return json.MarshalIndent(data, "", "  ")
}

// ToXML converts a map to a simple XML structure.
func (c *Converter) ToXML(data interface{}) ([]byte, error) {
	// Simple XML wrapper for maps
	if m, ok := data.(map[string]interface{}); ok {
		var sb strings.Builder
		sb.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
		sb.WriteString("<root>\n")
		for k, v := range m {
			sb.WriteString(fmt.Sprintf("  <%s>%v</%s>\n", k, v, k))
		}
		sb.WriteString("</root>")
		return []byte(sb.String()), nil
	}
	return xml.Marshal(data)
}

// ToSQL generates a simple SQL INSERT statement.
func (c *Converter) ToSQL(data map[string]interface{}, tableName string) ([]byte, error) {
	if tableName == "" {
		tableName = "form_data"
	}

	var cols, vals []string
	i := 1
	for k, v := range data {
		cols = append(cols, k)
		vals = append(vals, fmt.Sprintf("$%d", i))
		i++
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s);",
		tableName,
		strings.Join(cols, ", "),
		strings.Join(vals, ", "),
	)
	return []byte(query), nil
}
