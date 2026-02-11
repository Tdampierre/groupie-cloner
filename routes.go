package main

import (
	"database/sql"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func SetupRoutes(db *sql.DB) {
	// Proxies API
	http.HandleFunc("/api/artists-proxy", CreateProxy("https://groupietrackers.herokuapp.com/api/artists"))
	http.HandleFunc("/api/locations-proxy", CreateProxy("https://groupietrackers.herokuapp.com/api/locations"))
	http.HandleFunc("/api/dates-proxy", CreateProxy("https://groupietrackers.herokuapp.com/api/dates"))
	http.HandleFunc("/api/relation-proxy", CreateProxy("https://groupietrackers.herokuapp.com/api/relation"))
	// Alias pour compatibilité (certains JS utilisent /api/relations-proxy)
	http.HandleFunc("/api/relations-proxy", CreateProxy("https://groupietrackers.herokuapp.com/api/relation"))
	log.Println("API proxy routes registered")

	// Static files
	staticDir := filepath.Join("web", "static")
	log.Printf("static root: %s", staticDir)
	http.HandleFunc("/static/", ServeStatic(staticDir))

	// Templates
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

	// Auth routes
	http.HandleFunc("/api/register", HandleRegister(db))
	http.HandleFunc("/api/login", HandleLogin(db))
}

func CreateProxy(remote string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Proxying request to: %s", remote)

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Get(remote)
		if err != nil {
			log.Printf("Error fetching %s: %v", remote, err)
			http.Error(w, "failed to fetch remote API", http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		if ct := resp.Header.Get("Content-Type"); ct != "" {
			w.Header().Set("Content-Type", ct)
		} else {
			w.Header().Set("Content-Type", "application/json")
		}
		w.Header().Set("Access-Control-Allow-Origin", "*")

		w.WriteHeader(resp.StatusCode)
		if _, err := io.Copy(w, resp.Body); err != nil {
			log.Printf("Error copying response body: %v", err)
		}
		log.Printf("Successfully proxied request to: %s", remote)
	}
}

func ServeStatic(staticDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
	}
}
