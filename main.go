package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"groupiepersso/internal/database"
	"groupiepersso/internal/handlers"
)

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		w.Header().Set("Content-Security-Policy", "default-src 'self'")
		w.Header().Set("Referrer-Policy", "no-referrer")

		next.ServeHTTP(w, r)
	})
}

func main() {
	handlers.DB, _ = database.InitDB()
	handlers.SetupStaticRoutes(filepath.Join("web", "static"))
	handlers.SetupTemplateRoutes()
	http.HandleFunc("/api/artists-proxy", handlers.ProxyHandler("https://groupietrackers.herokuapp.com/api/artists"))
	http.HandleFunc("/api/locations-proxy", handlers.ProxyHandler("https://groupietrackers.herokuapp.com/api/locations"))
	http.HandleFunc("/api/dates-proxy", handlers.ProxyHandler("https://groupietrackers.herokuapp.com/api/dates"))
	http.HandleFunc("/api/relation-proxy", handlers.ProxyHandler("https://groupietrackers.herokuapp.com/api/relation"))
	http.HandleFunc("/api/register", handlers.HandleRegister)
	http.HandleFunc("/api/login", handlers.HandleLogin)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Starting server on :%s — open http://localhost:%s/", port, port)
	handler := securityHeaders(http.DefaultServeMux)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}
