package main

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// FrontendHandler wraps the API handler to serve frontend files for non-API requests
func FrontendHandler(apiHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Handle API requests with /api/ prefix (compatibility with frontend proxy)
		if strings.HasPrefix(r.URL.Path, "/api/") {
			r.URL.Path = strings.TrimPrefix(r.URL.Path, "/api")
			apiHandler.ServeHTTP(w, r)
			return
		}

		// 2. If not GET, it's definitely an API request (Frontend is static files + index.html for GET only)
		if r.Method != "GET" {
			apiHandler.ServeHTTP(w, r)
			return
		}

		distPath := "webui/dist"
		path := r.URL.Path

		// 3. Check if a physical file exists (e.g. /assets/main.css, /vite.svg)
		// Clean path to prevent directory traversal
		cleanPath := filepath.Clean(path)
		fullPath := filepath.Join(distPath, cleanPath)

		// Prevent escaping the dist directory
		if !strings.HasPrefix(fullPath, distPath) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		info, err := os.Stat(fullPath)
		if err == nil && !info.IsDir() {
			// File exists, serve it
			http.ServeFile(w, r, fullPath)
			return
		}

		// 4. If it's a browser navigation request (Accept: text/html), serve index.html
		accept := r.Header.Get("Accept")
		if strings.Contains(accept, "text/html") {
			http.ServeFile(w, r, filepath.Join(distPath, "index.html"))
			return
		}

		// 5. Otherwise, assume it's an API request (e.g. direct API call without /api prefix)
		apiHandler.ServeHTTP(w, r)
	})
}
