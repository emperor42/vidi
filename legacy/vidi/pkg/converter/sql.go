package converter

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// SQLDialect represents the SQL dialect for statement generation.
type SQLDialect string

const (
	// PostgreSQL dialect uses $1, $2, ... placeholders.
	PostgreSQL SQLDialect = "postgresql"

	// MySQL dialect uses ? placeholders.
	MySQL SQLDialect = "mysql"

	// SQLite dialect uses ? placeholders.
	SQLite SQLDialect = "sqlite"

	// SQLServer dialect uses @p1, @p2, ... placeholders.
	SQLServer SQLDialect = "sqlserver"
)

// SQLConverter handles conversion of data structures to SQL statements.
// It generates parameterized INSERT, UPDATE, and SELECT statements.
type SQLConverter struct {
	// Dialect specifies the SQL dialect for placeholder formatting.
	// Default is PostgreSQL.
	Dialect SQLDialect

	// QuoteIdentifiers determines whether to quote column/table names.
	// Default is true for safety with reserved keywords.
	QuoteIdentifiers bool

	// IdentifierQuote is the character used to quote identifiers.
	// Default is double-quote (") for PostgreSQL/ANSI SQL.
	IdentifierQuote string

	// IncludeReturning determines whether to add RETURNING clause.
	// Only applicable for PostgreSQL. Default is true.
	IncludeReturning bool

	// BatchSize is the maximum number of rows per INSERT statement.
	// Default is 1 (individual inserts).
	BatchSize int
}

// NewSQLConverter creates a new SQLConverter with default settings.
func NewSQLConverter() *SQLConverter {
	return &SQLConverter{
		Dialect:          PostgreSQL,
		QuoteIdentifiers: true,
		IdentifierQuote:  "\"",
		IncludeReturning: true,
		BatchSize:        1,
	}
}

// ToSQL generates an INSERT statement for the given data.
func (sc *SQLConverter) ToSQL(data map[string]interface{}, tableName string) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("cannot generate SQL for empty data")
	}

	if tableName == "" {
		tableName = "form_data"
	}

	return sc.GenerateInsert(data, tableName)
}

// GenerateInsert creates a parameterized INSERT statement.
func (sc *SQLConverter) GenerateInsert(data map[string]interface{}, tableName string) ([]byte, error) {
	var stmt strings.Builder

	// Sort keys for consistent output
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build column list
	stmt.WriteString("INSERT INTO ")
	stmt.WriteString(sc.quoteIdentifier(tableName))
	stmt.WriteString(" (")

	for i, key := range keys {
		if i > 0 {
			stmt.WriteString(", ")
		}
		stmt.WriteString(sc.quoteIdentifier(key))
	}

	stmt.WriteString(") VALUES (")

	// Build placeholders
	for i := range keys {
		if i > 0 {
			stmt.WriteString(", ")
		}
		stmt.WriteString(sc.placeholder(i + 1))
	}

	stmt.WriteString(")")

	// Add RETURNING clause for PostgreSQL
	if sc.Dialect == PostgreSQL && sc.IncludeReturning {
		stmt.WriteString(" RETURNING *")
	}

	stmt.WriteString(";")

	return []byte(stmt.String()), nil
}

// GenerateInsertWithValues creates an INSERT statement with actual values embedded.
// WARNING: This is less secure than parameterized queries. Use only for debugging/logs.
func (sc *SQLConverter) GenerateInsertWithValues(data map[string]interface{}, tableName string) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("cannot generate SQL for empty data")
	}

	if tableName == "" {
		tableName = "form_data"
	}

	var stmt strings.Builder

	// Sort keys for consistent output
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build column list
	stmt.WriteString("INSERT INTO ")
	stmt.WriteString(sc.quoteIdentifier(tableName))
	stmt.WriteString(" (")

	for i, key := range keys {
		if i > 0 {
			stmt.WriteString(", ")
		}
		stmt.WriteString(sc.quoteIdentifier(key))
	}

	stmt.WriteString(") VALUES (")

	// Build values
	for i, key := range keys {
		if i > 0 {
			stmt.WriteString(", ")
		}
		stmt.WriteString(sc.formatValue(data[key]))
	}

	stmt.WriteString(")")

	// Add RETURNING clause for PostgreSQL
	if sc.Dialect == PostgreSQL && sc.IncludeReturning {
		stmt.WriteString(" RETURNING *")
	}

	stmt.WriteString(";")

	return []byte(stmt.String()), nil
}

// GenerateUpdate creates a parameterized UPDATE statement.
func (sc *SQLConverter) GenerateUpdate(data map[string]interface{}, tableName, idField string, idValue interface{}) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("cannot generate UPDATE for empty data")
	}

	if tableName == "" {
		tableName = "form_data"
	}

	if idField == "" {
		idField = "id"
	}

	var stmt strings.Builder
	paramIndex := 1

	// Sort keys for consistent output
	keys := make([]string, 0, len(data))
	for k := range data {
		if k == idField {
			continue // Skip ID field in SET clause
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	stmt.WriteString("UPDATE ")
	stmt.WriteString(sc.quoteIdentifier(tableName))
	stmt.WriteString(" SET ")

	for i, key := range keys {
		if i > 0 {
			stmt.WriteString(", ")
		}
		stmt.WriteString(sc.quoteIdentifier(key))
		stmt.WriteString(" = ")
		stmt.WriteString(sc.placeholder(paramIndex))
		paramIndex++
	}

	stmt.WriteString(" WHERE ")
	stmt.WriteString(sc.quoteIdentifier(idField))
	stmt.WriteString(" = ")
	stmt.WriteString(sc.placeholder(paramIndex))

	stmt.WriteString(";")

	return []byte(stmt.String()), nil
}

// GenerateSelect creates a SELECT statement with optional WHERE conditions.
func (sc *SQLConverter) GenerateSelect(tableName string, conditions map[string]interface{}) ([]byte, error) {
	if tableName == "" {
		tableName = "form_data"
	}

	var stmt strings.Builder

	stmt.WriteString("SELECT * FROM ")
	stmt.WriteString(sc.quoteIdentifier(tableName))

	if len(conditions) > 0 {
		stmt.WriteString(" WHERE ")

		// Sort keys for consistent output
		keys := make([]string, 0, len(conditions))
		for k := range conditions {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for i, key := range keys {
			if i > 0 {
				stmt.WriteString(" AND ")
			}
			stmt.WriteString(sc.quoteIdentifier(key))
			stmt.WriteString(" = ")
			stmt.WriteString(sc.placeholder(i + 1))
		}
	}

	stmt.WriteString(";")

	return []byte(stmt.String()), nil
}

// GenerateDelete creates a DELETE statement.
func (sc *SQLConverter) GenerateDelete(tableName, idField string) ([]byte, error) {
	if tableName == "" {
		tableName = "form_data"
	}

	if idField == "" {
		idField = "id"
	}

	var stmt strings.Builder

	stmt.WriteString("DELETE FROM ")
	stmt.WriteString(sc.quoteIdentifier(tableName))
	stmt.WriteString(" WHERE ")
	stmt.WriteString(sc.quoteIdentifier(idField))
	stmt.WriteString(" = ")
	stmt.WriteString(sc.placeholder(1))
	stmt.WriteString(";")

	return []byte(stmt.String()), nil
}

// GenerateCreateTable creates a CREATE TABLE statement based on data structure.
func (sc *SQLConverter) GenerateCreateTable(tableName string, sampleData map[string]interface{}) ([]byte, error) {
	if tableName == "" {
		tableName = "form_data"
	}

	var stmt strings.Builder

	stmt.WriteString("CREATE TABLE IF NOT EXISTS ")
	stmt.WriteString(sc.quoteIdentifier(tableName))
	stmt.WriteString(" (\n")

	// Sort keys for consistent output
	keys := make([]string, 0, len(sampleData))
	for k := range sampleData {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for i, key := range keys {
		stmt.WriteString("    ")
		stmt.WriteString(sc.quoteIdentifier(key))
		stmt.WriteString(" ")
		stmt.WriteString(sc.inferColumnType(sampleData[key]))

		if key == "id" {
			stmt.WriteString(" PRIMARY KEY")
		}

		if i < len(keys)-1 {
			stmt.WriteString(",")
		}
		stmt.WriteString("\n")
	}

	stmt.WriteString(");")

	return []byte(stmt.String()), nil
}

// GenerateBatchInsert creates a batch INSERT statement for multiple rows.
func (sc *SQLConverter) GenerateBatchInsert(data []map[string]interface{}, tableName string) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("cannot generate batch INSERT for empty data")
	}

	if tableName == "" {
		tableName = "form_data"
	}

	// Get columns from first row
	firstRow := data[0]
	keys := make([]string, 0, len(firstRow))
	for k := range firstRow {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var stmt strings.Builder

	stmt.WriteString("INSERT INTO ")
	stmt.WriteString(sc.quoteIdentifier(tableName))
	stmt.WriteString(" (")

	for i, key := range keys {
		if i > 0 {
			stmt.WriteString(", ")
		}
		stmt.WriteString(sc.quoteIdentifier(key))
	}

	stmt.WriteString(") VALUES ")

	paramIndex := 1
	for rowIdx, row := range data {
		if rowIdx > 0 {
			stmt.WriteString(", ")
		}

		stmt.WriteString("(")
		for i, key := range keys {
			if i > 0 {
				stmt.WriteString(", ")
			}
			stmt.WriteString(sc.placeholder(paramIndex))
			paramIndex++
		}
		stmt.WriteString(")")
	}

	if sc.Dialect == PostgreSQL && sc.IncludeReturning {
		stmt.WriteString(" RETURNING *")
	}

	stmt.WriteString(";")

	return []byte(stmt.String()), nil
}

// placeholder returns the appropriate placeholder for the current dialect.
func (sc *SQLConverter) placeholder(index int) string {
	switch sc.Dialect {
	case PostgreSQL:
		return fmt.Sprintf("$%d", index)
	case MySQL, SQLite:
		return "?"
	case SQLServer:
		return fmt.Sprintf("@p%d", index)
	default:
		return fmt.Sprintf("$%d", index)
	}
}

// quoteIdentifier quotes an identifier (table/column name) for safety.
func (sc *SQLConverter) quoteIdentifier(name string) string {
	if !sc.QuoteIdentifiers {
		return name
	}

	// Sanitize the identifier first
	name = sanitizeIdentifier(name)

	// Use backticks for MySQL
	if sc.Dialect == MySQL {
		return fmt.Sprintf("`%s`", name)
	}

	// Use brackets for SQL Server
	if sc.Dialect == SQLServer {
		return fmt.Sprintf("[%s]", name)
	}

	// Use double quotes for PostgreSQL, SQLite, and ANSI SQL
	return fmt.Sprintf("%s%s%s", sc.IdentifierQuote, name, sc.IdentifierQuote)
}

// sanitizeIdentifier removes dangerous characters from identifiers.
func sanitizeIdentifier(name string) string {
	// Remove any quote characters that might break the identifier
	name = strings.ReplaceAll(name, "\"", "")
	name = strings.ReplaceAll(name, "'", "")
	name = strings.ReplaceAll(name, "`", "")
	name = strings.ReplaceAll(name, ";", "")
	name = strings.ReplaceAll(name, "--", "")

	// Remove whitespace
	name = strings.TrimSpace(name)

	return name
}

// formatValue formats a value for embedding in SQL.
// WARNING: This is for debugging only. Use parameterized queries for production.
func (sc *SQLConverter) formatValue(value interface{}) string {
	if value == nil {
		return "NULL"
	}

	switch v := value.(type) {
	case string:
		// Escape single quotes
		escaped := strings.ReplaceAll(v, "'", "''")
		return fmt.Sprintf("'%s'", escaped)
	case bool:
		if v {
			return "TRUE"
		}
		return "FALSE"
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", v)
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v)
	case float32, float64:
		return fmt.Sprintf("%g", v)
	case time.Time:
		return fmt.Sprintf("'%s'", v.Format("2006-01-02 15:04:05"))
	default:
		// Fall back to string representation with quoting
		escaped := strings.ReplaceAll(fmt.Sprintf("%v", v), "'", "''")
		return fmt.Sprintf("'%s'", escaped)
	}
}

// inferColumnType infers the SQL column type from a Go value.
func (sc *SQLConverter) inferColumnType(value interface{}) string {
	if value == nil {
		return "TEXT"
	}

	switch value.(type) {
	case int, int8, int16, int32, int64:
		return "BIGINT"
	case uint, uint8, uint16, uint32, uint64:
		return "BIGINT"
	case float32:
		return "REAL"
	case float64:
		return "DOUBLE PRECISION"
	case bool:
		return "BOOLEAN"
	case time.Time:
		return "TIMESTAMP"
	case string:
		return "TEXT"
	default:
		return "TEXT"
	}
}

// ExtractValues extracts values from data in the correct order for parameterized queries.
// The order matches the column order in generated INSERT statements.
func (sc *SQLConverter) ExtractValues(data map[string]interface{}) []interface{} {
	// Sort keys for consistent ordering
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	values := make([]interface{}, len(keys))
	for i, key := range keys {
		values[i] = data[key]
	}

	return values
}

// ExtractValuesForUpdate extracts values for UPDATE statements.
// Returns values for SET clause followed by the ID value for WHERE clause.
func (sc *SQLConverter) ExtractValuesForUpdate(data map[string]interface{}, idField string) []interface{} {
	if idField == "" {
		idField = "id"
	}

	// Sort keys (excluding ID) for consistent ordering
	keys := make([]string, 0, len(data))
	for k := range data {
		if k != idField {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	values := make([]interface{}, 0, len(keys)+1)

	// Add SET values
	for _, key := range keys {
		values = append(values, data[key])
	}

	// Add WHERE value (ID)
	if idValue, ok := data[idField]; ok {
		values = append(values, idValue)
	}

	return values
}

// ToSQLWithDialect generates SQL with a specific dialect.
// This is a convenience function for one-off conversions.
func ToSQLWithDialect(data map[string]interface{}, tableName string, dialect SQLDialect) ([]byte, error) {
	converter := &SQLConverter{
		Dialect:          dialect,
		QuoteIdentifiers: true,
		IdentifierQuote:  "\"",
		IncludeReturning: dialect == PostgreSQL,
		BatchSize:        1,
	}
	return converter.ToSQL(data, tableName)
}

// ToSQLSimple is a convenience function for quick SQL generation with PostgreSQL defaults.
func ToSQLSimple(data map[string]interface{}, tableName string) ([]byte, error) {
	converter := NewSQLConverter()
	return converter.ToSQL(data, tableName)
}

// Example usage:
//
//	data := map[string]interface{}{
//		"id":    "123",
//		"name":  "John O'Brien",
//		"email": "john@example.com",
//		"age":   30,
//	}
//
//	converter := NewSQLConverter()
//	sqlBytes, err := converter.ToSQL(data, "users")
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Println(string(sqlBytes))
//
// Output (PostgreSQL):
// INSERT INTO "users" ("age", "email", "id", "name") VALUES ($1, $2, $3, $4) RETURNING *;
//
// To execute with parameterized query:
// values := converter.ExtractValues(data)
// db.Exec(string(sqlBytes), values...)
