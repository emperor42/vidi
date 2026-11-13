package main

import (
    "encoding/json"
    "fmt"
    "net/http"
    "sync"
    "time"
)

type Record struct {
    ID        string    `json:"id"`
    Title     string    `json:"title"`
    Content   string    `json:"content"`
    CreatedAt time.Time `json:"created_at"`
}

type Cache struct {
    mu       sync.RWMutex
    data     map[string]interface{}
    ttl      time.Duration
    lastClean time.Time
}

var cache = &Cache{
    data: make(map[string]interface{}),
    ttl:  5 * time.Minute,
}

func recordsHandler(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodGet:
        handleGetRecords(w, r)
    case http.MethodPost:
        handlePostRecord(w, r)
    default:
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    }
}

func handleGetRecords(w http.ResponseWriter, r *http.Request) {
    cache.mu.RLock()
    defer cache.mu.RUnlock()
    
    // Query parameters for filtering
    titleFilter := strings.TrimSpace(r.URL.Query().Get("title"))
    idFilter := strings.TrimSpace(r.URL.Query().Get("id"))
    
    records := []Record{}
    for id, data := range cache.data {
        if record, ok := data.(Record); ok {
            if titleFilter == "" || strings.Contains(strings.ToLower(record.Title), strings.ToLower(titleFilter)) {
                if idFilter == "" || id == idFilter {
                    records = append(records, record)
                }
            }
        }
    }
    
    w.Header().Set("Content-Type", "application/json")
    response := map[string]interface{}{
        "records": records,
        "total":   len(records),
        "filters": map[string]string{"title": titleFilter, "id": idFilter},
    }
    
    if err := json.NewEncoder(w).Encode(response); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

func handlePostRecord(w http.ResponseWriter, r *http.Request) {
    var newRecord Record
    if err := json.NewDecoder(r.Body).Decode(&newRecord); err != nil {
        http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
        return
    }
    
    if newRecord.Title == "" {
        http.Error(w, "Title is required", http.StatusBadRequest)
        return
    }
    
    if newRecord.ID == "" {
        newRecord.ID = fmt.Sprintf("rec-%d", time.Now().UnixNano())
    }
    
    newRecord.CreatedAt = time.Now()
    
    cache.mu.Lock()
    defer cache.mu.Unlock()
    
    cache.data[newRecord.ID] = newRecord
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    if err := json.NewEncoder(w).Encode(newRecord); err != nil {
        fmt.Printf("Error encoding response: %v\n", err)
    }
}

func cleanCache() {
    cache.mu.Lock()
    defer cache.mu.Unlock()
    
    now := time.Now()
    if now.Sub(cache.lastClean) < time.Minute {
        return
    }
    
    for id, data := range cache.data {
        if record, ok := data.(Record); ok {
            if now.Sub(record.CreatedAt) > cache.ttl {
                delete(cache.data, id)
            }
        }
    }
    
    cache.lastClean = now
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
    cache.mu.RLock()
    defer cache.mu.RUnlock()
    
    health := map[string]interface{}{
        "status": "healthy",
        "service": "vidi",
        "version": "3.0.0",
        "records_count": len(cache.data),
        "cache_ttl_seconds": int(cache.ttl.Seconds()),
        "last_clean": cache.lastClean,
        "features": []string{"records", "search", "filter", "cache", "api"},
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(health)
}

func main() {
    go func() {
        for {
            cleanCache()
            time.Sleep(time.Minute)
        }
    }()
    
    http.HandleFunc("/api/records", recordsHandler)
    http.HandleFunc("/health", healthHandler)
    
    fmt.Printf("VIDI record service starting on port 8084\n")
    if err := http.ListenAndServe(":8084", nil); err != nil {
        fmt.Printf("Error starting server: %v\n", err)
    }
}