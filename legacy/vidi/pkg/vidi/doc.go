/*
Package vidi provides middleware for processing HTML form submissions in Go applications.

VIDI (Validate, Integrate, Data, Input) is a lightweight, standard-library-only middleware
that converts form data into structured formats (JSON, XML, or SQL) and persists them to
various backends (filesystem, in-memory, or PostgreSQL-ready interfaces).

# Overview

VIDI intercepts HTTP requests and provides two main functionalities:

 1. POST requests: Process form submissions and store data
 2. GET requests: Query stored data using URL parameters as filters

The middleware is designed to wrap around http.FileServer or any other http.Handler,
making it easy to integrate into existing Go web applications.

# Installation

	import "github.com/Emperor42/vidi/pkg/vidi"

# Quick Start

	package main

	import (
		"net/http"
		"github.com/Emperor42/vidi/pkg/vidi"
		"github.com/Emperor42/vidi/pkg/storage"
	)

	func main() {
		// Create storage backend
		store := storage.NewFilesystemBackend("./data")

		// Configure VIDI
		config := vidi.Config{
			Storage:        store,
			Format:         "json",
			AutoGenerateID: true,
			TableName:      "users",
		}

		// Create middleware
		middleware := vidi.New(config)

		// Wrap your handler
		fileServer := http.FileServer(http.Dir("./static"))
		handler := middleware.Handle(fileServer)

		// Start server
		http.ListenAndServe(":8080", handler)
	}

# Configuration

The Config struct controls VIDI's behavior:

	type Config struct {
		// Storage backend (required)
		Storage storage.Backend

		// Output format: "json", "xml", or "sql"
		Format string

		// Auto-generate unique IDs and timestamps
		AutoGenerateID bool

		// SQL table name for SQL format
		TableName string

		// URL endpoint to intercept (default: "/forms")
		Endpoint string
	}

# Storage Backends

VIDI supports pluggable storage backends via the storage.Backend interface:

 1. FilesystemBackend: Persists data as JSON files (pkg/storage/filesystem.go)
 2. MemoryBackend: In-memory storage for testing (pkg/storage/memory.go)
 3. Custom: Implement storage.Backend for PostgreSQL, MongoDB, etc.

# Output Formats

 1. JSON (default): Pretty-printed JSON objects or arrays
 2. XML: Simple XML wrapper with fields as elements
 3. SQL: Generated INSERT statements for PostgreSQL compatibility

# API Endpoints

POST /forms

	Submits form data. Returns the stored record in the configured format.
	Content-Type: application/x-www-form-urlencoded or multipart/form-data

GET /forms?field=value

	Queries stored data. Returns matching records as JSON array.
	Multiple filters can be combined: ?name=john&city=nyc

# Query Syntax

URL parameters are treated as query filters:

	GET /forms?name=john          # Find records where name == "john"
	GET /forms?city=nyc&age=30    # Find records matching both conditions
	GET /forms                    # Return all records

# Integration with VENI and VICI

VIDI is designed to be composable with other middlewares:

	// Layering order: VICI (Auth) → VIDI (Data) → VENI (Components) → FileServer
	handler := viciMiddleware.Handle(
		vidiMiddleware.Handle(
			veniMiddleware.Handle(fileServer),
		),
	)

# Thread Safety

All VIDI operations are thread-safe. The storage backends implement proper locking
mechanisms to handle concurrent requests.

# Error Handling

VIDI returns appropriate HTTP status codes:

	200 OK: Successful GET query
	201 Created: Successful POST submission
	400 Bad Request: Invalid form data
	500 Internal Server Error: Storage or processing errors

Error responses are returned in JSON format:

	{"error": "descriptive error message"}

# Examples

See the examples/ directory for complete working examples:

 1. Basic: Simple form submission and query
 2. Advanced: Custom storage backend implementation
 3. Integration: Combining with VENI and VICI

# Limitations

 1. Standard Library Only: No external dependencies (as per design)
 2. SQL Generation: Currently generates SQL strings only (no actual DB execution)
 3. File Uploads: Multipart form file fields are parsed but not stored
 4. Authentication: Not included (use VICI for auth requirements)

# Version Compatibility

Minimum Go version: 1.21
Latest tested: Go 1.22

# License

MIT License - See LICENSE file for details.

# Author

Emperor42
https://github.com/Emperor42/vidi

# See Also

 1. VENI: Custom element middleware for vanilla JS web components
 2. VICI: Context tracking and authentication middleware
 3. Proton: Privacy-focused secure communication suite
*/
package vidi
