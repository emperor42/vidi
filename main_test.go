package main

import (
	"encoding/json"
	"html/template"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func newTestHandler(t *testing.T) (*templateServer, http.Handler) {
	t.Helper()
	dataDir := filepath.Join(t.TempDir(), "templates")
	server, err := newTemplateServer(dataDir, defaultIndexTemplate)
	if err != nil {
		t.Fatalf("newTemplateServer() error = %v", err)
	}
	return server, newHandler(server)
}

func TestValidTemplateName(t *testing.T) {
	tests := []struct {
		name  string
		valid bool
	}{
		{name: "home", valid: true},
		{name: "Home-page_1.0", valid: true},
		{name: "", valid: false},
		{name: ".", valid: false},
		{name: "..", valid: false},
		{name: "../outside", valid: false},
		{name: `..\outside`, valid: false},
		{name: "nested/name", valid: false},
		{name: "name with spaces", valid: false},
		{name: strings.Repeat("a", maxTemplateNameLength+1), valid: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := validTemplateName(tt.name); got != tt.valid {
				t.Fatalf("validTemplateName(%q) = %v, want %v", tt.name, got, tt.valid)
			}
		})
	}
}

func TestAPIRequiresTokenWhenConfigured(t *testing.T) {
	t.Setenv("VIDI_API_TOKEN", "test-token")
	server, handler := newTestHandler(t)
	body, err := json.Marshal(TemplateInfo{Name: "protected", HTML: "<p>no</p>"})
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/templates", strings.NewReader(string(body)))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated POST status = %d, want 401", rr.Code)
	}
	if _, err := os.Stat(filepath.Join(server.templatesDir, "protected.json")); !os.IsNotExist(err) {
		t.Fatalf("unauthenticated request wrote a file: %v", err)
	}

	req = httptest.NewRequest(http.MethodPost, "/api/templates", strings.NewReader(string(body)))
	req.Header.Set("X-Vidi-Token", "test-token")
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("authenticated POST status = %d, want 201", rr.Code)
	}
}

func TestNonLoopbackListenRequiresToken(t *testing.T) {
	t.Setenv("VIDI_API_TOKEN", "")
	if err := validateListenSecurity("0.0.0.0:8084"); err == nil {
		t.Fatal("non-loopback listener without token was accepted")
	}
	t.Setenv("VIDI_API_TOKEN", "test-token")
	if err := validateListenSecurity("0.0.0.0:8084"); err != nil {
		t.Fatalf("non-loopback listener with token rejected: %v", err)
	}
}

func TestSaveAndServeTemplate(t *testing.T) {
	server, handler := newTestHandler(t)
	info := TemplateInfo{Name: "welcome", HTML: "<p>hello</p>"}
	body, err := json.Marshal(info)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/templates", strings.NewReader(string(body)))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("POST status = %d, want %d: %s", rr.Code, http.StatusCreated, rr.Body.String())
	}

	stored, err := os.ReadFile(filepath.Join(server.templatesDir, "welcome.json"))
	if err != nil {
		t.Fatalf("stored template: %v", err)
	}
	if string(stored) != string(body) {
		t.Fatalf("stored template = %q, want %q", stored, body)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/templates/welcome", nil)
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("GET status = %d, want %d", rr.Code, http.StatusOK)
	}
	if rr.Body.String() != info.HTML {
		t.Fatalf("GET body = %q, want %q", rr.Body.String(), info.HTML)
	}
}

func TestTemplateNameCannotEscapeDataDirectory(t *testing.T) {
	server, handler := newTestHandler(t)
	parent := filepath.Dir(server.templatesDir)
	outside := filepath.Join(parent, "outside.json")

	for _, name := range []string{"../outside", `..\outside`, "nested/name"} {
		t.Run(name, func(t *testing.T) {
			body, err := json.Marshal(TemplateInfo{Name: name, HTML: "<p>no</p>"})
			if err != nil {
				t.Fatal(err)
			}
			req := httptest.NewRequest(http.MethodPost, "/api/templates", strings.NewReader(string(body)))
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
			if rr.Code != http.StatusBadRequest {
				t.Fatalf("POST status = %d, want %d", rr.Code, http.StatusBadRequest)
			}
		})
	}

	if _, err := os.Stat(outside); !os.IsNotExist(err) {
		t.Fatalf("outside file exists or stat failed unexpectedly: %v", err)
	}
}

func TestMalformedStoredTemplateReturnsServerError(t *testing.T) {
	server, handler := newTestHandler(t)
	path := filepath.Join(server.templatesDir, "broken.json")
	if err := os.WriteFile(path, []byte("{"), 0600); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/templates/broken", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("GET status = %d, want %d", rr.Code, http.StatusInternalServerError)
	}
}

func TestTemplateWriteFailureIsReported(t *testing.T) {
	blocked := filepath.Join(t.TempDir(), "not-a-directory")
	if err := os.WriteFile(blocked, []byte("occupied"), 0600); err != nil {
		t.Fatal(err)
	}
	index, err := template.ParseFiles(defaultIndexTemplate)
	if err != nil {
		t.Fatal(err)
	}
	server := &templateServer{templatesDir: blocked, index: index}
	handler := newHandler(server)

	body, err := json.Marshal(TemplateInfo{Name: "unwritable", HTML: "<p>no</p>"})
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/templates", strings.NewReader(string(body)))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("POST status = %d, want %d", rr.Code, http.StatusInternalServerError)
	}
}

func TestUnsupportedTemplateMethods(t *testing.T) {
	_, handler := newTestHandler(t)

	for _, tc := range []struct {
		method string
		path   string
		allow  string
	}{
		{method: http.MethodPut, path: "/api/templates", allow: "GET, POST"},
		{method: http.MethodPatch, path: "/api/templates/example", allow: "GET, DELETE"},
	} {
		t.Run(tc.method+tc.path, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
			if rr.Code != http.StatusMethodNotAllowed {
				t.Fatalf("status = %d, want %d", rr.Code, http.StatusMethodNotAllowed)
			}
			if got := rr.Header().Get("Allow"); got != tc.allow {
				t.Fatalf("Allow = %q, want %q", got, tc.allow)
			}
		})
	}
}
