package main

import (
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"
    "time"
)

func TestRecordsHandlerGet(t *testing.T) {
    req, err := http.NewRequest("GET", "/api/records", nil)
    if err != nil {
        t.Fatal(err)
    }
    
    rr := httptest.NewRecorder()
    handler := http.HandlerFunc(recordsHandler)
    
    handler.ServeHTTP(rr, req)
    
    if status := rr.Code; status != http.StatusOK {
        t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
    }
    
    var response map[string]interface{}
    if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
        t.Errorf("Error decoding response: %v", err)
    }
    
    // The response structure is different - check for the correct fields
    if records, ok := response["records"]; ok {
        if recordSlice, ok := records.([]interface{}); ok {
            if len(recordSlice) != 0 {
                t.Errorf("Expected 0 records (empty cache), got %d", len(recordSlice))
            }
        } else {
            t.Errorf("Expected 'records' field to be a slice")
        }
    } else {
        t.Errorf("Expected 'records' field in response")
    }
}

func TestRecordsHandlerPost(t *testing.T) {
    newRecord := Record{
        ID: "test-rec-123",
        Title: "Test Record",
        Content: "This is a test record",
        CreatedAt: time.Now(),
    }
    
    recordJSON, err := json.Marshal(newRecord)
    if err != nil {
        t.Fatal(err)
    }
    
    req, err := http.NewRequest("POST", "/api/records", strings.NewReader(string(recordJSON)))
    if err != nil {
        t.Fatal(err)
    }
    req.Header.Set("Content-Type", "application/json")
    
    rr := httptest.NewRecorder()
    handler := http.HandlerFunc(recordsHandler)
    
    handler.ServeHTTP(rr, req)
    
    if status := rr.Code; status != http.StatusCreated {
        t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusCreated)
    }
    
    var response Record
    if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
        t.Errorf("Error decoding response: %v", err)
    }
    
    if response.Title != "Test Record" {
        t.Errorf("Expected record title 'Test Record', got %v", response.Title)
    }
}

func TestHealthHandler(t *testing.T) {
    req, err := http.NewRequest("GET", "/health", nil)
    if err != nil {
        t.Fatal(err)
    }
    
    rr := httptest.NewRecorder()
    handler := http.HandlerFunc(healthHandler)
    
    handler.ServeHTTP(rr, req)
    
    if status := rr.Code; status != http.StatusOK {
        t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
    }
    
    var health map[string]interface{}
    if err := json.NewDecoder(rr.Body).Decode(&health); err != nil {
        t.Errorf("Error decoding response: %v", err)
    }
    
    if health["service"] != "vidi" {
        t.Errorf("Expected service 'vidi', got %v", health["service"])
    }
    
    if health["status"] != "healthy" {
        t.Errorf("Expected status 'healthy', got %v", health["status"])
    }
}

func TestCacheCleanup(t *testing.T) {
    // Add a record with old timestamp
    oldRecord := Record{
        ID: "old-rec",
        Title: "Old Record",
        Content: "This record is old",
        CreatedAt: time.Now().Add(-10 * time.Minute),
    }
    
    cache.mu.Lock()
    cache.data[oldRecord.ID] = oldRecord
    cache.mu.Unlock()
    
    // Trigger cleanup
    cleanCache()
    
    // Check if old record was removed
    cache.mu.RLock()
    _, exists := cache.data["old-rec"]
    cache.mu.RUnlock()
    
    if exists {
        t.Errorf("Old record should have been cleaned up by cache")
    }
}