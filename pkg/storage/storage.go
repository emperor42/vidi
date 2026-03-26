package storage

// Backend defines the interface for data persistence.
// All implementations must be safe for concurrent use.
type Backend interface {
	// Store saves a record. If the data map does not contain an "id" key,
	// the implementation may generate one.
	Store(data map[string]interface{}) error

	// Query retrieves records matching the provided parameters.
	// An empty params map should return all records.
	Query(params map[string]interface{}) ([]map[string]interface{}, error)

	// Get retrieves a single record by its ID.
	// Returns an error if the record is not found.
	Get(id string) (map[string]interface{}, error)

	// Delete removes a record by its ID.
	// Returns an error if the record is not found.
	Delete(id string) error
}
