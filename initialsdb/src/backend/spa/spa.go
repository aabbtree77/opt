package spa

import (
	"net/http"
	"path/filepath"
)

type SPAHandler struct {
	Dir string
}

func (h SPAHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Try to open the requested file
	if _, err := http.Dir(h.Dir).Open(r.URL.Path); err == nil {
		http.FileServer(http.Dir(h.Dir)).ServeHTTP(w, r)
		return
	}

	// Fallback to index.html for client-side routing
	http.ServeFile(w, r, filepath.Join(h.Dir, "index.html"))
}
