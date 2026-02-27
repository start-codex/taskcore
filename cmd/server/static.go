package main

import (
	"io/fs"
	"net/http"

	"github.com/start-codex/taskcode/ui"
)

// registerUI mounts the embedded SPA under "/". API routes registered before
// this catch their paths first; everything else falls through to the SPA.
// index.html is served for any path that doesn't match a real file so that
// client-side routing works correctly.
func registerUI(mux *http.ServeMux) {
	sub, err := fs.Sub(ui.FS, "dist")
	if err != nil {
		panic(err)
	}
	fileServer := http.FileServer(http.FS(sub))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Try to open the requested path; if not found serve index.html.
		f, err := sub.Open(r.URL.Path[1:]) // strip leading /
		if err != nil {
			r.URL.Path = "/"
		} else {
			f.Close()
		}
		fileServer.ServeHTTP(w, r)
	})
}
