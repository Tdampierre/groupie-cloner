package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"groupiepersso/internal/database"
	"groupiepersso/internal/handlers"
)

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
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
