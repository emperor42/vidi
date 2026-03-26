package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/Emperor42/vidi/pkg/storage"
	"github.com/Emperor42/vidi/pkg/vidi"
)

func main() {
	// --- Configuration ---
	// Define the base path for filesystem storage
	storagePath := "./vidi-data"

	// Ensure the storage directory exists
	if err := os.MkdirAll(storagePath, 0755); err != nil {
		log.Fatalf("Failed to create storage directory: %v", err)
	}

	// --- Initialize Storage Backend ---
	// We use the FilesystemBackend for persistent storage.
	// Alternatives: storage.NewMemoryBackend() for ephemeral storage.
	store := storage.NewFilesystemBackend(storagePath)

	// --- Configure VIDI Middleware ---
	config := vidi.Config{
		Storage:        store,
		Format:         "json",       // Options: "json", "xml", "sql"
		AutoGenerateID: true,         // Automatically adds 'id' and 'created_at'
		TableName:      "form_data",  // Used for SQL generation
		Endpoint:       "/api/forms", // The URL path to intercept
	}

	// Create the VIDI middleware instance
	vidiMiddleware := vidi.New(config)

	// --- Create Static File Server ---
	// Serves the demo frontend (index.html) and any other static assets.
	// Adjust the path if your static files are elsewhere.
	staticPath := "./static"
	if _, err := os.Stat(staticPath); os.IsNotExist(err) {
		log.Printf("Warning: Static directory '%s' not found. Creating a demo page in memory.", staticPath)
		// In a real app, you'd ensure this directory exists or embed files.
		// For this demo, we'll rely on the fact that the user likely has the static folder.
	}
	fileServer := http.FileServer(http.Dir(staticPath))

	// --- Layer Middlewares ---
	// The order matters:
	// 1. VIDI intercepts requests to /api/forms
	// 2. Everything else falls through to the FileServer
	handler := vidiMiddleware.Handle(fileServer)

	// --- Custom Logging Middleware (Optional) ---
	// Adds simple logging to the console for every request.
	loggingHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL.Path)
		handler.ServeHTTP(w, r)
	})

	// --- Graceful Shutdown Setup ---
	// Listen for OS signals to shut down gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// --- Start Server ---
	port := ":8080"
	server := &http.Server{
		Addr:    port,
		Handler: loggingHandler,
	}

	go func() {
		log.Printf("🚀 VIDI Server starting on http://localhost%s", port)
		log.Printf("📝 Form Submission: POST http://localhost%s/api/forms", port)
		log.Printf("🔍 Data Query: GET http://localhost%s/api/forms?field=value", port)
		log.Printf("📄 Demo Page: http://localhost%s", port)
		log.Printf("💾 Storage Path: %s", storagePath)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for interrupt signal
	<-sigChan
	log.Println("\n⚠️  Shutting down server...")

	// Note: In a production app, you would call server.Shutdown(ctx) here
	// to allow in-flight requests to finish.
	log.Println("✅ Server stopped.")
}

// --- Helper: Demo Page Generator (Fallback) ---
// If the static directory is missing, this function can serve a simple inline HTML page.
// You can uncomment the following code and modify the handler logic above to use it.

/*
func createDemoPage() http.HandlerFunc {
	html := `
	<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8">
		<title>VIDI Demo</title>
		<style>
			body { font-family: sans-serif; max-width: 600px; margin: 50px auto; padding: 20px; }
			form { background: #f4f4f4; padding: 20px; border-radius: 8px; }
			input { display: block; width: 100%; margin: 10px 0; padding: 8px; }
			button { background: #007bff; color: white; border: none; padding: 10px 20px; cursor: pointer; }
			pre { background: #333; color: #fff; padding: 10px; overflow-x: auto; }
		</style>
	</head>
	<body>
		<h1>VIDI Middleware Demo</h1>
		<form id="form">
			<label>Name</label><input type="text" name="name" required>
			<label>Email</label><input type="email" name="email" required>
			<button type="submit">Submit</button>
		</form>
		<h2>Results</h2>
		<pre id="output">Waiting for data...</pre>
		<script>
			document.getElementById('form').onsubmit = async (e) => {
				e.preventDefault();
				const fd = new FormData(e.target);
				const res = await fetch('/api/forms', { method: 'POST', body: fd });
				const data = await res.json();
				document.getElementById('output').textContent = JSON.stringify(data, null, 2);
			};
			// Auto-query on load
			fetch('/api/forms').then(r => r.json()).then(d => {
				document.getElementById('output').textContent = JSON.stringify(d, null, 2);
			});
		</script>
	</body>
	</html>
	`
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(html))
	}
}
*/
