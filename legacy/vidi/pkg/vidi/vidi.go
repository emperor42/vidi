package vidi

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Emperor42/vidi/pkg/converter"
	"github.com/Emperor42/vidi/pkg/storage"
)

// VIDI is the middleware handler.
type VIDI struct {
	config    Config
	storage   storage.Backend
	converter *converter.Converter
}

// New creates a new VIDI instance.
func New(cfg Config) *VIDI {
	if cfg.Storage == nil {
		cfg.Storage = storage.NewMemoryBackend()
	}
	if cfg.Format == "" {
		cfg.Format = "json"
	}
	if cfg.TableName == "" {
		cfg.TableName = "form_data"
	}
	if cfg.Endpoint == "" {
		cfg.Endpoint = "/forms"
	}

	return &VIDI{
		config:    cfg,
		storage:   cfg.Storage,
		converter: converter.New(),
	}
}

// Handle wraps the next handler with VIDI logic.
func (v *VIDI) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if this request should be intercepted
		if !v.isIntercepted(r) {
			next.ServeHTTP(w, r)
			return
		}

		switch r.Method {
		case http.MethodPost:
			v.handlePost(w, r)
		case http.MethodGet:
			v.handleGet(w, r)
		default:
			next.ServeHTTP(w, r)
		}
	})
}

func (v *VIDI) isIntercepted(r *http.Request) bool {
	return r.URL.Path == v.config.Endpoint || strings.HasPrefix(r.URL.Path, v.config.Endpoint+"/")
}

func (v *VIDI) handlePost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	data := make(map[string]interface{})
	for k, v := range r.Form {
		if len(v) == 1 {
			data[k] = v[0]
		} else {
			data[k] = v
		}
	}

	// Auto-generate ID and timestamp
	if v.config.AutoGenerateID {
		if _, ok := data["id"]; !ok {
			data["id"] = fmt.Sprintf("%d", time.Now().UnixNano())
		}
		if _, ok := data["created_at"]; !ok {
			data["created_at"] = time.Now().UTC().Format(time.RFC3339)
		}
	}

	if err := v.storage.Store(data); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	v.sendResponse(w, data, http.StatusCreated)
}

func (v *VIDI) handleGet(w http.ResponseWriter, r *http.Request) {
	params := make(map[string]interface{})
	for k, v := range r.URL.Query() {
		if len(v) == 1 {
			params[k] = v[0]
		} else {
			params[k] = v
		}
	}

	results, err := v.storage.Query(params)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	v.sendResponse(w, results, http.StatusOK)
}

func (v *VIDI) sendResponse(w http.ResponseWriter, data interface{}, status int) {
	var resp []byte
	var ctype string
	var err error

	switch v.config.Format {
	case "xml":
		resp, err = v.converter.ToXML(data)
		ctype = "application/xml"
	case "sql":
		if m, ok := data.(map[string]interface{}); ok {
			resp, err = v.converter.ToSQL(m, v.config.TableName)
		} else if arr, ok := data.([]map[string]interface{}); ok && len(arr) > 0 {
			// For arrays in SQL mode, just return the first one or JSON
			resp, _ = v.converter.ToJSON(data)
			ctype = "application/json"
		} else {
			resp, _ = v.converter.ToJSON(data)
			ctype = "application/json"
		}
		if ctype == "" {
			ctype = "text/plain"
		}
	default: // json
		resp, err = v.converter.ToJSON(data)
		ctype = "application/json"
	}

	if err != nil {
		http.Error(w, "Format Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", ctype)
	w.WriteHeader(status)
	w.Write(resp)
}
