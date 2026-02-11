package handlers

import (
	"io"
	"log"
	"net/http"
	"time"
)

func ProxyHandler(remote string) http.HandlerFunc {
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
		_, err = io.Copy(w, resp.Body)
		if err != nil {
			log.Printf("Error copying response body: %v", err)
		}
		log.Printf("Successfully proxied request to: %s", remote)
	}
}
