package vidi

import "github.com/Emperor42/vidi/pkg/storage"

// Config holds the configuration for the VIDI middleware.
type Config struct {
	// Storage is the backend used for persistence.
	Storage storage.Backend

	// Format determines the response format: "json", "xml", or "sql".
	Format string

	// AutoGenerateID adds a unique ID and timestamp if missing.
	AutoGenerateID bool

	// TableName is used for SQL generation.
	TableName string

	// Endpoint is the URL path to intercept (e.g., "/forms").
	Endpoint string
}

// DefaultConfig returns a sensible default configuration.
func DefaultConfig() Config {
	return Config{
		Format:         "json",
		AutoGenerateID: true,
		TableName:      "form_data",
		Endpoint:       "/forms",
	}
}
