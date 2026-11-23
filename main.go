package main

import (
	"bytes"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	defaultTemplatesDir  = "templates_data"
	defaultIndexTemplate = "templates/index.html"
	defaultHost          = "127.0.0.1"
	defaultPort          = "8084"

	maxTemplateNameLength = 128
	maxTemplateBodySize   = int64(1 << 20) // 1 MiB
)

var errInvalidTemplateName = errors.New("invalid template name")

type TemplateInfo struct {
	Name string `json:"name"`
	HTML string `json:"html"`
}

type templateServer struct {
	templatesDir string
	index        *template.Template
}

func main() {
	server, err := newTemplateServer(defaultTemplatesDir, defaultIndexTemplate)
	if err != nil {
		log.Fatal(err)
	}
	addr := listenAddress()
	if err := validateListenSecurity(addr); err != nil {
		log.Fatal(err)
	}
	httpServer := &http.Server{
		Addr:              addr,
		Handler:           newHandler(server),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	log.Printf("vidi demo listening on %s", addr)
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func newTemplateServer(dataDir, indexPath string) (*templateServer, error) {
	if dataDir == "" {
		dataDir = defaultTemplatesDir
	}
	if indexPath == "" {
		indexPath = defaultIndexTemplate
	}

	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("create template directory: %w", err)
	}
	index, err := template.ParseFiles(indexPath)
	if err != nil {
		return nil, fmt.Errorf("parse index template: %w", err)
	}

	return &templateServer{
		templatesDir: dataDir,
		index:        index,
	}, nil
}

func newHandler(server *templateServer) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", server.rootHandler)
	mux.HandleFunc("/api/templates", server.templatesHandler)
	mux.HandleFunc("/api/templates/", server.templateByNameHandler)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	return demoSecurityHeaders(apiAuth(mux))
}

func apiToken() string { return strings.TrimSpace(os.Getenv("VIDI_API_TOKEN")) }

func apiAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		want := apiToken()
		if want == "" || strings.HasPrefix(r.URL.Path, "/static/") {
			// No token is the local-development mode. main() prevents that
			// mode from being combined with a non-loopback listener.
			next.ServeHTTP(w, r)
			return
		}
		got := strings.TrimSpace(r.Header.Get("X-Vidi-Token"))
		if auth := strings.TrimSpace(r.Header.Get("Authorization")); strings.HasPrefix(strings.ToLower(auth), "bearer ") {
			got = strings.TrimSpace(auth[len("Bearer "):])
		}
		if got == "" || subtle.ConstantTimeCompare([]byte(got), []byte(want)) != 1 {
			w.Header().Set("WWW-Authenticate", `Bearer realm="vidi"`)
			http.Error(w, "template API authentication required", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func demoSecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "no-referrer")
		next.ServeHTTP(w, r)
	})
}

func listenAddress() string {
	host := strings.TrimSpace(os.Getenv("VIDI_HOST"))
	if host == "" {
		host = defaultHost
	}
	port := strings.TrimSpace(os.Getenv("VIDI_PORT"))
	if port == "" {
		port = defaultPort
	}
	return net.JoinHostPort(host, port)
}

func validateListenSecurity(addr string) error {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return fmt.Errorf("invalid listen address %q: %w", addr, err)
	}
	host = strings.Trim(host, "[]")
	if host == "localhost" {
		return nil
	}
	ip := net.ParseIP(host)
	if ip != nil && ip.IsLoopback() {
		return nil
	}
	if apiToken() == "" {
		return fmt.Errorf("refusing non-loopback listen address %q without VIDI_API_TOKEN", addr)
	}
	return nil
}

func (s *templateServer) rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		name := strings.TrimPrefix(r.URL.Path, "/")
		if validTemplateName(name) {
			s.serveTemplate(w, r, name)
			return
		}
		http.NotFound(w, r)
		return
	}

	entries, err := os.ReadDir(s.templatesDir)
	if err != nil {
		http.Error(w, "could not read templates", http.StatusInternalServerError)
		return
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), ".json")
		if validTemplateName(name) {
			names = append(names, name)
		}
	}

	var body bytes.Buffer
	if s.index == nil {
		http.Error(w, "index template is unavailable", http.StatusInternalServerError)
		return
	}
	if err := s.index.Execute(&body, names); err != nil {
		http.Error(w, "could not render index template", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	writeResponse(w, body.Bytes())
}

func (s *templateServer) serveTemplate(w http.ResponseWriter, r *http.Request, name string) {
	path, err := templateFilePath(s.templatesDir, name)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	info, err := readTemplateFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "could not read template", http.StatusInternalServerError)
		return
	}

	// The demo intentionally stores executable HTML templates. The storage
	// directory and its contents are not an authorization boundary; callers
	// that expose this demo must protect it separately.
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	writeResponse(w, []byte(info.HTML))
}

func (s *templateServer) templatesHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listTemplates(w, r)
	case http.MethodPost:
		s.saveTemplate(w, r)
	default:
		w.Header().Set("Allow", http.MethodGet+", "+http.MethodPost)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *templateServer) listTemplates(w http.ResponseWriter, _ *http.Request) {
	entries, err := os.ReadDir(s.templatesDir)
	if err != nil {
		http.Error(w, "could not read templates", http.StatusInternalServerError)
		return
	}

	list := make([]TemplateInfo, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), ".json")
		path, err := templateFilePath(s.templatesDir, name)
		if err != nil {
			continue
		}
		info, err := readTemplateFile(path)
		if err != nil {
			http.Error(w, "could not read template", http.StatusInternalServerError)
			return
		}
		list = append(list, info)
	}

	data, err := json.Marshal(list)
	if err != nil {
		http.Error(w, "could not encode templates", http.StatusInternalServerError)
		return
	}
	js := append([]byte("var templates = "), data...)
	js = append(js, []byte(";\n")...)
	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	writeResponse(w, js)
}

func (s *templateServer) saveTemplate(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		http.Error(w, "JSON body required", http.StatusBadRequest)
		return
	}

	body := http.MaxBytesReader(w, r.Body, maxTemplateBodySize)
	decoder := json.NewDecoder(body)
	var info TemplateInfo
	if err := decoder.Decode(&info); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	// Reject a second JSON value or trailing non-whitespace data. This keeps
	// malformed input from being silently accepted and persisted.
	var extra json.RawMessage
	if err := decoder.Decode(&extra); err != io.EOF {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	if !validTemplateName(info.Name) {
		http.Error(w, "invalid template name", http.StatusBadRequest)
		return
	}

	path, err := templateFilePath(s.templatesDir, info.Name)
	if err != nil {
		http.Error(w, "invalid template name", http.StatusBadRequest)
		return
	}
	data, err := json.Marshal(info)
	if err != nil {
		http.Error(w, "could not encode template", http.StatusInternalServerError)
		return
	}
	if err := writeFileAtomic(path, data); err != nil {
		http.Error(w, "could not save template", http.StatusInternalServerError)
		return
	}

	response, err := json.Marshal(info)
	if err != nil {
		http.Error(w, "could not encode template", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	writeResponse(w, append(response, '\n'))
}

func (s *templateServer) templateByNameHandler(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/api/templates/")
	if !validTemplateName(name) {
		http.NotFound(w, r)
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.serveTemplate(w, r, name)
	case http.MethodDelete:
		path, err := templateFilePath(s.templatesDir, name)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		if err := os.Remove(path); err != nil {
			if os.IsNotExist(err) {
				http.NotFound(w, r)
				return
			}
			http.Error(w, "could not delete template", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		w.Header().Set("Allow", http.MethodGet+", "+http.MethodDelete)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func writeFileAtomic(path string, data []byte) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".vidi-template-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if err := tmp.Chmod(0600); err != nil {
		_ = tmp.Close()
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}

func readTemplateFile(path string) (TemplateInfo, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return TemplateInfo{}, err
	}
	var info TemplateInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return TemplateInfo{}, err
	}
	return info, nil
}

// validTemplateName deliberately accepts only a single, conservative filename
// component. In particular, both slash styles are rejected so a JSON name
// cannot escape templatesDir even when a file is later moved between systems.
func validTemplateName(name string) bool {
	if name == "" || len(name) > maxTemplateNameLength || name == "." || name == ".." {
		return false
	}
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || r == '-' || r == '_' || r == '.' {
			continue
		}
		return false
	}
	return true
}

func templateFilePath(dir, name string) (string, error) {
	if !validTemplateName(name) {
		return "", errInvalidTemplateName
	}

	base, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	base = filepath.Clean(base)
	path := filepath.Join(base, name+".json")
	// This is defense in depth in addition to the character allowlist.
	if filepath.Dir(path) != base {
		return "", errInvalidTemplateName
	}
	return path, nil
}

func writeResponse(w http.ResponseWriter, data []byte) {
	if _, err := w.Write(data); err != nil {
		log.Printf("vidi: response write failed: %v", err)
	}
}
