package httpkit

import (
	"io/fs"
	"net/http"
	"strings"
)

// apiPrefixes are path prefixes that should never fall through to the SPA.
// Unknown paths under these prefixes return a JSON 404 instead of index.html.
var apiPrefixes = []string{"/api/", "/swagger/", "/healthz", "/ws"}

// SPAHandler serves static files from the given filesystem and falls back to
// index.html for any path that doesn't match a file. This enables client-side
// routing in single-page applications. API paths are excluded from the
// fallback and receive a JSON 404 instead.
func SPAHandler(staticFS fs.FS) http.HandlerFunc {
	fileServer := http.FileServer(http.FS(staticFS))

	return func(w http.ResponseWriter, r *http.Request) {
		// API paths should never serve the SPA — return a proper JSON 404.
		for _, prefix := range apiPrefixes {
			if strings.HasPrefix(r.URL.Path, prefix) {
				JSON(w, http.StatusNotFound, map[string]string{"message": "not found"})
				return
			}
		}

		path := strings.TrimPrefix(r.URL.Path, "/")

		// Try to open the file. If it exists and is not a directory, serve it.
		if f, err := staticFS.Open(path); err == nil {
			stat, statErr := f.Stat()
			f.Close()

			if statErr == nil && !stat.IsDir() {
				fileServer.ServeHTTP(w, r)
				return
			}
		}

		// File doesn't exist — serve index.html for client-side routing.
		r.URL.Path = "/"
		fileServer.ServeHTTP(w, r)
	}
}
