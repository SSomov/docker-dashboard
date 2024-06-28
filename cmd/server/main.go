package main

import (
	"fmt"
	"net/http"
	"path/filepath"

	"docker-dashboard/internal/api"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()
	api.RegisterRoutes(r)

	// Serve static files from the "web/public" directory
	staticDir := filepath.Join("web", "public")
	r.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir(staticDir))))

	fmt.Println("Server started at http://localhost:8080")
	http.ListenAndServe(":8080", r)
}

