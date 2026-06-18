package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// FilesystemBackend stores records as individual JSON files.
type FilesystemBackend struct {
	basePath string
	mu       sync.RWMutex
}

// NewFilesystemBackend initializes the backend and creates the directory if needed.
func NewFilesystemBackend(basePath string) *FilesystemBackend {
	if err := os.MkdirAll(basePath, 0755); err != nil {
		panic(fmt.Sprintf("failed to create storage directory: %v", err))
	}
	return &FilesystemBackend{basePath: basePath}
}

func (fs *FilesystemBackend) Store(data map[string]interface{}) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	id, ok := data["id"].(string)
	if !ok || id == "" {
		id = fmt.Sprintf("%d", time.Now().UnixNano())
		data["id"] = id
	}

	filename := filepath.Join(fs.basePath, id+".json")
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	// Atomic write: write to temp, then rename
	tmpFile := filename + ".tmp"
	if err := os.WriteFile(tmpFile, jsonData, 0644); err != nil {
		return err
	}
	if err := os.Rename(tmpFile, filename); err != nil {
		os.Remove(tmpFile)
		return err
	}
	return nil
}

func (fs *FilesystemBackend) Query(params map[string]interface{}) ([]map[string]interface{}, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	var results []map[string]interface{}

	// Optimization: if querying by ID, skip scanning
	if id, ok := params["id"].(string); ok && id != "" {
		rec, err := fs.getUnlocked(id)
		if err == nil && matchesQuery(rec, params) {
			results = append(results, rec)
		}
		return results, nil
	}

	files, err := os.ReadDir(fs.basePath)
	if err != nil {
		return nil, err
	}

	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".json") {
			continue
		}
		id := strings.TrimSuffix(f.Name(), ".json")
		rec, err := fs.getUnlocked(id)
		if err != nil {
			continue
		}
		if matchesQuery(rec, params) {
			results = append(results, rec)
		}
	}
	return results, nil
}

func (fs *FilesystemBackend) Get(id string) (map[string]interface{}, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	return fs.getUnlocked(id)
}

func (fs *FilesystemBackend) getUnlocked(id string) (map[string]interface{}, error) {
	filename := filepath.Join(fs.basePath, id+".json")
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var record map[string]interface{}
	if err := json.Unmarshal(data, &record); err != nil {
		return nil, err
	}
	return record, nil
}

func (fs *FilesystemBackend) Delete(id string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	filename := filepath.Join(fs.basePath, id+".json")
	if err := os.Remove(filename); err != nil {
		return err
	}
	return nil
}
