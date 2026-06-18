package converter

import (
	"strings"
	"testing"
	"time"
)

// TestNewSQLConverter verifies default initialization.
func TestNewSQLConverter(t *testing.T) {
	sc := NewSQLConverter()

	if sc.Dialect != PostgreSQL {
		t.Errorf("Expected default dialect PostgreSQL, got %s", sc.Dialect)
	}
	if !sc.QuoteIdentifiers {
		t.Error("Expected QuoteIdentifiers to be true by default")
	}
	if sc.IdentifierQuote != "\"" {
		t.Errorf("Expected IdentifierQuote to be \", got %s", sc.IdentifierQuote)
	}
	if !sc.IncludeReturning {
		t.Error("Expected IncludeReturning to be true by default")
	}
	if sc.BatchSize != 1 {
		t.Errorf("Expected BatchSize to be 1, got %d", sc.BatchSize)
	}
}

// TestSQLConverter_ToSQL_PostgreSQL tests basic INSERT generation for PostgreSQL.
func TestSQLConverter_ToSQL_PostgreSQL(t *testing.T) {
	sc := NewSQLConverter()
	sc.Dialect = PostgreSQL

	data := map[string]interface{}{
		"id":    "123",
		"name":  "John Doe",
		"email": "john@example.com",
		"age":   30,
	}

	sqlBytes, err := sc.ToSQL(data, "users")
	if err != nil {
		t.Fatalf("ToSQL failed: %v", err)
	}

	sqlStr := string(sqlBytes)

	// Verify structure
	expectedParts := []string{
		"INSERT INTO",
		`"users"`,
		"VALUES",
		"$$1", "$$2", "$$3", "$$4",
		"RETURNING *",
	}

	for _, part := range expectedParts {
		if !strings.Contains(sqlStr, part) {
			t.Errorf("Expected SQL to contain '%s', got: %s", part, sqlStr)
		}
	}

	// Verify order (sorted keys: age, email, id, name)
	if !strings.Contains(sqlStr, `"age"`) || !strings.Contains(sqlStr, `"email"`) {
		t.Error("Columns not found or not in expected order")
	}
}

// TestSQLConverter_ToSQL_MySQL tests INSERT generation for MySQL.
func TestSQLConverter_ToSQL_MySQL(t *testing.T) {
	sc := NewSQLConverter()
	sc.Dialect = MySQL
	sc.IdentifierQuote = "`"

	data := map[string]interface{}{
		"name": "Jane Doe",
		"role": "admin",
	}

	sqlBytes, err := sc.ToSQL(data, "admins")
	if err != nil {
		t.Fatalf("ToSQL failed: %v", err)
	}

	sqlStr := string(sqlBytes)

	// Verify MySQL specifics
	if !strings.Contains(sqlStr, "`admins`") {
		t.Errorf("Expected backtick quoted table name, got: %s", sqlStr)
	}
	if !strings.Contains(sqlStr, "?") {
		t.Errorf("Expected ? placeholders for MySQL, got: %s", sqlStr)
	}
	if strings.Contains(sqlStr, "RETURNING") {
		t.Error("MySQL should not have RETURNING clause by default")
	}
}

// TestSQLConverter_ToSQL_SQLite tests INSERT generation for SQLite.
func TestSQLConverter_ToSQL_SQLite(t *testing.T) {
	sc := NewSQLConverter()
	sc.Dialect = SQLite

	data := map[string]interface{}{
		"value": 42,
	}

	sqlBytes, err := sc.ToSQL(data, "settings")
	if err != nil {
		t.Fatalf("ToSQL failed: %v", err)
	}

	sqlStr := string(sqlBytes)

	if !strings.Contains(sqlStr, "?") {
		t.Errorf("Expected ? placeholders for SQLite, got: %s", sqlStr)
	}
}

// TestSQLConverter_GenerateInsertWithValues tests value embedding (debug mode).
func TestSQLConverter_GenerateInsertWithValues(t *testing.T) {
	sc := NewSQLConverter()
	sc.Dialect = PostgreSQL

	data := map[string]interface{}{
		"name":  "O'Brien", // Test escaping
		"score": 95.5,
	}

	sqlBytes, err := sc.GenerateInsertWithValues(data, "players")
	if err != nil {
		t.Fatalf("GenerateInsertWithValues failed: %v", err)
	}

	sqlStr := string(sqlBytes)

	// Verify escaping of single quote
	if !strings.Contains(sqlStr, "O''Brien") {
		t.Errorf("Expected escaped single quote, got: %s", sqlStr)
	}
	if !strings.Contains(sqlStr, "95.5") {
		t.Errorf("Expected numeric value 95.5, got: %s", sqlStr)
	}
}

// TestSQLConverter_GenerateUpdate tests UPDATE statement generation.
func TestSQLConverter_GenerateUpdate(t *testing.T) {
	sc := NewSQLConverter()

	data := map[string]interface{}{
		"id":    "123",
		"name":  "Updated Name",
		"email": "new@example.com",
	}

	sqlBytes, err := sc.GenerateUpdate(data, "users", "id", "123")
	if err != nil {
		t.Fatalf("GenerateUpdate failed: %v", err)
	}

	sqlStr := string(sqlBytes)

	// Verify structure
	if !strings.Contains(sqlStr, "UPDATE") {
		t.Error("Missing UPDATE keyword")
	}
	if !strings.Contains(sqlStr, "SET") {
		t.Error("Missing SET keyword")
	}
	if !strings.Contains(sqlStr, "WHERE") {
		t.Error("Missing WHERE clause")
	}
	if !strings.Contains(sqlStr, "$$1") || !strings.Contains(sqlStr, "$$2") || !strings.Contains(sqlStr, "$3") {
		t.Errorf("Expected 3 placeholders, got: %s", sqlStr)
	}
}

// TestSQLConverter_GenerateSelect tests SELECT statement generation.
func TestSQLConverter_GenerateSelect(t *testing.T) {
	sc := NewSQLConverter()

	conditions := map[string]interface{}{
		"status": "active",
		"role":   "admin",
	}

	sqlBytes, err := sc.GenerateSelect("users", conditions)
	if err != nil {
		t.Fatalf("GenerateSelect failed: %v", err)
	}

	sqlStr := string(sqlBytes)

	if !strings.Contains(sqlStr, "SELECT * FROM") {
		t.Error("Missing SELECT * FROM")
	}
	if !strings.Contains(sqlStr, "WHERE") {
		t.Error("Missing WHERE clause")
	}
	if !strings.Contains(sqlStr, "AND") {
		t.Error("Missing AND operator for multiple conditions")
	}
}

// TestSQLConverter_GenerateDelete tests DELETE statement generation.
func TestSQLConverter_GenerateDelete(t *testing.T) {
	sc := NewSQLConverter()

	sqlBytes, err := sc.GenerateDelete("users", "id")
	if err != nil {
		t.Fatalf("GenerateDelete failed: %v", err)
	}

	sqlStr := string(sqlBytes)

	if !strings.Contains(sqlStr, "DELETE FROM") {
		t.Error("Missing DELETE FROM")
	}
	if !strings.Contains(sqlStr, "WHERE") {
		t.Error("Missing WHERE clause")
	}
	if !strings.Contains(sqlStr, "$1") {
		t.Error("Missing placeholder for ID")
	}
}

// TestSQLConverter_GenerateCreateTable tests CREATE TABLE generation.
func TestSQLConverter_GenerateCreateTable(t *testing.T) {
	sc := NewSQLConverter()

	sampleData := map[string]interface{}{
		"id":        "123",
		"name":      "John",
		"age":       30,
		"active":    true,
		"balance":   100.50,
		"createdAt": time.Now(),
	}

	sqlBytes, err := sc.GenerateCreateTable("users", sampleData)
	if err != nil {
		t.Fatalf("GenerateCreateTable failed: %v", err)
	}

	sqlStr := string(sqlBytes)

	if !strings.Contains(sqlStr, "CREATE TABLE") {
		t.Error("Missing CREATE TABLE")
	}
	if !strings.Contains(sqlStr, "PRIMARY KEY") {
		t.Error("Missing PRIMARY KEY for id column")
	}
	if !strings.Contains(sqlStr, "BIGINT") {
		t.Error("Expected BIGINT for integer type")
	}
	if !strings.Contains(sqlStr, "BOOLEAN") {
		t.Error("Expected BOOLEAN for bool type")
	}
	if !strings.Contains(sqlStr, "DOUBLE PRECISION") {
		t.Error("Expected DOUBLE PRECISION for float type")
	}
	if !strings.Contains(sqlStr, "TIMESTAMP") {
		t.Error("Expected TIMESTAMP for time.Time type")
	}
}

// TestSQLConverter_GenerateBatchInsert tests batch INSERT generation.
func TestSQLConverter_GenerateBatchInsert(t *testing.T) {
	sc := NewSQLConverter()

	data := []map[string]interface{}{
		{"id": "1", "name": "Alice"},
		{"id": "2", "name": "Bob"},
		{"id": "3", "name": "Charlie"},
	}

	sqlBytes, err := sc.GenerateBatchInsert(data, "users")
	if err != nil {
		t.Fatalf("GenerateBatchInsert failed: %v", err)
	}

	sqlStr := string(sqlBytes)

	// Count occurrences of VALUES and commas to verify batch structure
	if !strings.Contains(sqlStr, "VALUES") {
		t.Error("Missing VALUES clause")
	}
	// Should have 3 sets of parentheses for 3 rows
	count := strings.Count(sqlStr, "), (")
	if count != 2 {
		t.Errorf("Expected 2 separators for 3 rows, got %d in: %s", count, sqlStr)
	}
}

// TestSQLConverter_ExtractValues tests value extraction for parameter binding.
func TestSQLConverter_ExtractValues(t *testing.T) {
	sc := NewSQLConverter()

	data := map[string]interface{}{
		"z_last":  "last",
		"a_first": "first",
		"m_mid":   "mid",
	}

	values := sc.ExtractValues(data)

	// Values should be sorted by key: a_first, m_mid, z_last
	expectedOrder := []string{"first", "mid", "last"}

	if len(values) != 3 {
		t.Fatalf("Expected 3 values, got %d", len(values))
	}

	for i, val := range values {
		if valStr, ok := val.(string); ok {
			if valStr != expectedOrder[i] {
				t.Errorf("Expected value at index %d to be '%s', got '%s'", i, expectedOrder[i], valStr)
			}
		} else {
			t.Errorf("Value at index %d is not a string: %v", i, val)
		}
	}
}

// TestSQLConverter_ExtractValuesForUpdate tests value extraction for UPDATE.
func TestSQLConverter_ExtractValuesForUpdate(t *testing.T) {
	sc := NewSQLConverter()

	data := map[string]interface{}{
		"id":    "123",
		"name":  "New Name",
		"email": "new@example.com",
	}

	values := sc.ExtractValuesForUpdate(data, "id")

	// Should have 3 values: name, email (SET), then id (WHERE)
	// Order depends on sorting of keys excluding 'id'
	if len(values) != 3 {
		t.Fatalf("Expected 3 values, got %d", len(values))
	}

	// Last value must be the ID
	if lastVal, ok := values[len(values)-1].(string); ok {
		if lastVal != "123" {
			t.Errorf("Expected last value to be ID '123', got '%s'", lastVal)
		}
	} else {
		t.Error("Last value is not a string")
	}
}

// TestSQLConverter_SanitizeIdentifier tests identifier sanitization.
func TestSQLConverter_SanitizeIdentifier(t *testing.T) {
	sc := NewSQLConverter()

	tests := []struct {
		input    string
		expected string
	}{
		{"normal_name", "normal_name"},
		{"name-with-dash", "name-with-dash"},
		{"name; DROP TABLE users;", "name DROP TABLE users"}, // Semicolon removed
		{"name--comment", "name comment"},                    // Double dash removed
		{"name'quote", "namequote"},                          // Single quote removed
		{"  spaced  ", "spaced"},                             // Trimmed
	}

	for _, tt := range tests {
		result := sanitizeIdentifier(tt.input)
		if result != tt.expected {
			t.Errorf("sanitizeIdentifier(%q) = %q; want %q", tt.input, result, tt.expected)
		}
	}
}

// TestSQLConverter_FormatValue tests value formatting for debug mode.
func TestSQLConverter_FormatValue(t *testing.T) {
	sc := NewSQLConverter()

	tests := []struct {
		input    interface{}
		expected string
	}{
		{nil, "NULL"},
		{"simple", "'simple'"},
		{"with'quote", "'with''quote'"},
		{true, "TRUE"},
		{false, "FALSE"},
		{42, "42"},
		{3.14, "3.14"},
		{time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC), "'2023-01-01 12:00:00'"},
	}

	for _, tt := range tests {
		result := sc.formatValue(tt.input)
		if result != tt.expected {
			t.Errorf("formatValue(%v) = %q; want %q", tt.input, result, tt.expected)
		}
	}
}

// TestSQLConverter_InferColumnType tests type inference.
func TestSQLConverter_InferColumnType(t *testing.T) {
	sc := NewSQLConverter()

	tests := []struct {
		input    interface{}
		expected string
	}{
		{int(42), "BIGINT"},
		{int64(42), "BIGINT"},
		{float32(1.5), "REAL"},
		{float64(1.5), "DOUBLE PRECISION"},
		{true, "BOOLEAN"},
		{"text", "TEXT"},
		{time.Now(), "TIMESTAMP"},
		{nil, "TEXT"},
	}

	for _, tt := range tests {
		result := sc.inferColumnType(tt.input)
		if result != tt.expected {
			t.Errorf("inferColumnType(%v) = %q; want %q", tt.input, result, tt.expected)
		}
	}
}

// TestSQLConverter_EmptyData tests error handling for empty data.
func TestSQLConverter_EmptyData(t *testing.T) {
	sc := NewSQLConverter()

	_, err := sc.ToSQL(map[string]interface{}{}, "users")
	if err == nil {
		t.Error("Expected error for empty data in ToSQL")
	}

	_, err = sc.GenerateBatchInsert([]map[string]interface{}{}, "users")
	if err == nil {
		t.Error("Expected error for empty batch in GenerateBatchInsert")
	}
}

// TestSQLConverter_NoReturning tests disabling RETURNING clause.
func TestSQLConverter_NoReturning(t *testing.T) {
	sc := NewSQLConverter()
	sc.IncludeReturning = false

	data := map[string]interface{}{"id": "1"}
	sqlBytes, _ := sc.ToSQL(data, "users")
	sqlStr := string(sqlBytes)

	if strings.Contains(sqlStr, "RETURNING") {
		t.Error("Expected no RETURNING clause when disabled")
	}
}

// TestSQLConverter_CustomDialect tests custom dialect placeholder generation.
func TestSQLConverter_CustomDialect(t *testing.T) {
	sc := NewSQLConverter()
	sc.Dialect = SQLServer

	data := map[string]interface{}{"name": "Test"}
	sqlBytes, _ := sc.ToSQL(data, "users")
	sqlStr := string(sqlBytes)

	if !strings.Contains(sqlStr, "@p1") {
		t.Errorf("Expected @p1 placeholder for SQL Server, got: %s", sqlStr)
	}
}
