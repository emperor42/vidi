/*
Package storage provides pluggable storage backends for the VIDI middleware.

The Backend interface defines the contract that all storage implementations must satisfy:

	type Backend interface {
		Store(data map[string]interface{}) error
		Query(params map[string]interface{}) ([]map[string]interface{}, error)
		Get(id string) (map[string]interface{}, error)
		Delete(id string) error
	}

# Available Implementations

 1. FilesystemBackend: Persists records as individual JSON files
 2. MemoryBackend: In-memory storage for testing and development

# Thread Safety

All implementations are safe for concurrent use by multiple goroutines.

# Extending Storage

To create a custom backend:

	type MyBackend struct {
		// Your fields
	}

	func (m *MyBackend) Store(data map[string]interface{}) error {
		// Implementation
	}

	func (m *MyBackend) Query(params map[string]interface{}) ([]map[string]interface{}, error) {
		// Implementation
	}

	func (m *MyBackend) Get(id string) (map[string]interface{}, error) {
		// Implementation
	}

	func (m *MyBackend) Delete(id string) error {
		// Implementation
	}

Then use it with VIDI:

	store := &MyBackend{}
	config := vidi.Config{Storage: store}
	middleware := vidi.New(config)

# See Also

 1. VIDI package: github.com/Emperor42/vidi/pkg/vidi
 2. Converter package: github.com/Emperor42/vidi/pkg/converter

# License

MIT License - See LICENSE file for details.

# Author

Emperor42
*/
package storage
