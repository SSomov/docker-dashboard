package main

import (
	"log"
	"net/http"
	"path/filepath"
	"time"

	"docker-dashboard/internal/api"

	"github.com/gorilla/mux"
)

func main() {
	for {
		r := mux.NewRouter()
		api.RegisterRoutes(r)

		// Serve static files from the "web/public" directory
		staticDir := filepath.Join("web", "public")
		r.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir(staticDir))))

		addr := ":8080"
		log.Printf("Server started at http://localhost%s", addr)
		err := http.ListenAndServe(addr, r)
		if err != nil {
			log.Printf("Server error: %v", err)
		}
		log.Println("Restarting server in 10 seconds...")
		time.Sleep(10 * time.Second)
	}
}
