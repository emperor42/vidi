package main

import (
	"log"
	"net/http"

	"github.com/Emperor42/vidi/pkg/storage"
	"github.com/Emperor42/vidi/pkg/vidi"
)

func main() {
	// 1. Initialize Storage
	// Using MemoryBackend for this example (data is lost on restart).
	// For persistent storage, use: storage.NewFilesystemBackend("./data")
	store := storage.NewMemoryBackend()

	// 2. Configure VIDI
	config := vidi.Config{
		Storage:        store,
		Format:         "json",   // Output format: "json", "xml", "sql"
		AutoGenerateID: true,     // Automatically adds 'id' and 'created_at'
		TableName:      "users",  // Used if Format is "sql"
		Endpoint:       "/forms", // The URL path to intercept
	}

	// 3. Create Middleware
	vidiMiddleware := vidi.New(config)

	// 4. Create a simple handler for the root path
	// This serves a tiny HTML form for testing purposes.
	rootHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		html := `
		<!DOCTYPE html>
		<html>
		<head>
			<title>VIDI Basic Example</title>
			<style>
				body { font-family: sans-serif; max-width: 500px; margin: 50px auto; padding: 20px; }
				form { background: #f9f9f9; padding: 20px; border: 1px solid #ddd; border-radius: 5px; }
				label { display: block; margin-top: 10px; font-weight: bold; }
				input { width: 100%; padding: 8px; margin-top: 5px; box-sizing: border-box; }
				button { margin-top: 15px; padding: 10px 20px; background: #007bff; color: white; border: none; cursor: pointer; }
				button:hover { background: #0056b3; }
				pre { background: #333; color: #fff; padding: 15px; border-radius: 5px; overflow-x: auto; }
			</style>
		</head>
		<body>
			<h1>VIDI Basic Example</h1>
			<p>Submit a form to store data, or query it via URL parameters.</p>
			
			<h2>Submit Data</h2>
			<form action="/forms" method="POST">
				<label>Name</label>
				<input type="text" name="name" required placeholder="John Doe">
				
				<label>Email</label>
				<input type="email" name="email" required placeholder="john@example.com">
				
				<label>Role</label>
				<input type="text" name="role" placeholder="Developer">
				
				<button type="submit">Submit</button>
			</form>

			<h2>Query Data</h2>
			<p>Try visiting: <code>/forms?role=Developer</code></p>
			
			<h3>Current Records:</h3>
			<pre id="results">Loading...</pre>

			<script>
				// Fetch all records on load
				fetch('/forms')
					.then(res => res.json())
					.then(data => {
						document.getElementById('results').textContent = JSON.stringify(data, null, 2);
					})
					.catch(err => {
						document.getElementById('results').textContent = 'Error: ' + err.message;
					});
			</script>
		</body>
		</html>
		`
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(html))
	})

	// 5. Wrap the root handler with VIDI middleware
	// Requests to /forms will be handled by VIDI.
	// Requests to / will be handled by rootHandler.
	mux := http.NewServeMux()
	mux.Handle("/", rootHandler)

	// Wrap the mux with VIDI
	// Note: VIDI checks the path. If it matches the endpoint, it handles it.
	// Otherwise, it passes to the next handler (the mux).
	handler := vidiMiddleware.Handle(mux)

	// 6. Start Server
	port := ":8080"
	log.Printf("🚀 Starting VIDI Basic Example on http://localhost%s", port)
	log.Printf("📝 Submit: POST http://localhost%s/forms", port)
	log.Printf("🔍 Query: GET http://localhost%s/forms?role=Developer", port)

	if err := http.ListenAndServe(port, handler); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

// --- Optional: Demonstrate Programmatic Usage ---
// This function shows how to use VIDI without HTTP, purely as a library.
// Uncomment the lines below to test this logic separately.
/*
func demonstrateLibraryUsage() {
	store := storage.NewMemoryBackend()
	config := vidi.Config{
		Storage:        store,
		Format:         "json",
		AutoGenerateID: true,
	}
	vidiMiddleware := vidi.New(config)

	// Simulate form data
	formData := map[string]interface{}{
		"name":  "Alice",
		"email": "alice@example.com",
	}

	// Process directly (no HTTP request needed)
	record, err := vidiMiddleware.ProcessFormFromMap(formData)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Stored Record: %+v\n", record)

	// Query directly
	results, err := vidiMiddleware.ExecuteQuery(map[string]interface{}{"name": "Alice"})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Query Results: %+v\n", results)
}
*/
