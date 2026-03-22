package httpkit

import (
	"io/fs"
	"net/http"
	"strings"
)

// SPAHandler serves static files from the given filesystem and falls back to
// index.html for any path that doesn't match a file. This enables client-side
// routing in single-page applications.
func SPAHandler(staticFS fs.FS) http.HandlerFunc {
	fileServer := http.FileServer(http.FS(staticFS))

	return func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")

		// Try to open the file. If it exists, serve it directly.
		if f, err := staticFS.Open(path); err == nil {
			f.Close()
			fileServer.ServeHTTP(w, r)
			return
		}

		// File doesn't exist — serve index.html for client-side routing.
		r.URL.Path = "/"
		fileServer.ServeHTTP(w, r)
	}
}
