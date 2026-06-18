package main

import (
	"encoding/json"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var templatesDir = "templates_data"

type TemplateInfo struct {
	Name string `json:"name"`
	HTML string `json:"html"`
}

func main() {
	os.MkdirAll(templatesDir, 0755)

	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/api/templates", templatesHandler)
	http.HandleFunc("/api/templates/", templateByNameHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.ListenAndServe(":8084", nil)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		name := strings.TrimPrefix(r.URL.Path, "/")
		if name != "" && !strings.Contains(name, "/") {
			serveTemplate(w, r, name)
			return
		}
		http.NotFound(w, r)
		return
	}
	tmpl := template.Must(template.ParseFiles("templates/index.html"))
	entries, _ := os.ReadDir(templatesDir)
	var names []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") {
			names = append(names, strings.TrimSuffix(e.Name(), ".json"))
		}
	}
	tmpl.Execute(w, names)
}

func serveTemplate(w http.ResponseWriter, r *http.Request, name string) {
	data, err := os.ReadFile(filepath.Join(templatesDir, name+".json"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	var t TemplateInfo
	json.Unmarshal(data, &t)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(t.HTML))
}

func templatesHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		entries, _ := os.ReadDir(templatesDir)
		var list []TemplateInfo
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") {
				data, _ := os.ReadFile(filepath.Join(templatesDir, e.Name()))
				var t TemplateInfo
				json.Unmarshal(data, &t)
				list = append(list, t)
			}
		}
		js := "var templates = " + mustJSON(list) + ";"
		w.Header().Set("Content-Type", "application/javascript")
		w.Write([]byte(js))

	case http.MethodPost:
		body, _ := io.ReadAll(r.Body)
		var t TemplateInfo
		json.Unmarshal(body, &t)
		if t.Name == "" {
			http.Error(w, "name required", http.StatusBadRequest)
			return
		}
		data, _ := json.Marshal(t)
		os.WriteFile(filepath.Join(templatesDir, t.Name+".json"), data, 0644)
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(t)
	}
}

func templateByNameHandler(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/api/templates/")
	if name == "" || strings.Contains(name, "/") {
		http.NotFound(w, r)
		return
	}
	switch r.Method {
	case http.MethodGet:
		serveTemplate(w, r, name)
	case http.MethodDelete:
		os.Remove(filepath.Join(templatesDir, name+".json"))
		w.WriteHeader(http.StatusNoContent)
	}
}

func mustJSON(v interface{}) string {
	data, _ := json.Marshal(v)
	return string(data)
}
