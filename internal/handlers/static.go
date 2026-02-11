package handlers

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func SetupStaticRoutes(staticDir string) {
	log.Printf("static root: %s", staticDir)

	http.HandleFunc("/static/", func(w http.ResponseWriter, r *http.Request) {
		reqPath := r.URL.Path[len("/static/"):]
		full := filepath.Join(staticDir, filepath.FromSlash(reqPath))

		if fi, err := os.Stat(full); err == nil && !fi.IsDir() {
			ext := filepath.Ext(full)
			switch ext {
			case ".css":
				w.Header().Set("Content-Type", "text/css")
			case ".js":
				w.Header().Set("Content-Type", "application/javascript")
			case ".png":
				w.Header().Set("Content-Type", "image/png")
			case ".jpg", ".jpeg":
				w.Header().Set("Content-Type", "image/jpeg")
			case ".svg":
				w.Header().Set("Content-Type", "image/svg+xml")
			case ".gif":
				w.Header().Set("Content-Type", "image/gif")
			case ".webp":
				w.Header().Set("Content-Type", "image/webp")
			}
			w.Header().Set("Cache-Control", "public, max-age=31536000")
			http.ServeFile(w, r, full)
			return
		}
		http.NotFound(w, r)
	})
}

func SetupTemplateRoutes() {
	serveIndex := func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join("index.html"))
	}

	http.HandleFunc("/", serveIndex)

	routes := []string{"/search", "/filters"}
	for _, rt := range routes {
		rlocal := rt
		http.HandleFunc(rlocal, serveIndex)
	}

	http.HandleFunc("/geoloc", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/geoloc.html", http.StatusMovedPermanently)
	})

	http.HandleFunc("/search.html", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join("web", "templates", "search.html"))
	})

	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join("web", "templates", "login.html"))
	})

	http.HandleFunc("/login.html", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join("web", "templates", "login.html"))
	})

	http.HandleFunc("/geoloc.html", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join("web", "templates", "geoloc.html"))
	})
}
