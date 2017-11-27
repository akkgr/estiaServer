package adapters

import (
	"net/http"
	"os"
	"strings"
)

// Static ...
func Static(folder string, page string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestPath := r.URL.Path
		fileSystemPath := folder + r.URL.Path
		endURIPath := strings.Split(requestPath, "/")[len(strings.Split(requestPath, "/"))-1]
		splitPath := strings.Split(endURIPath, ".")
		if len(splitPath) > 1 {
			if f, err := os.Stat(fileSystemPath); err == nil && !f.IsDir() {
				http.ServeFile(w, r, fileSystemPath)
				return
			}
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, folder+"/"+page)
	})
}
