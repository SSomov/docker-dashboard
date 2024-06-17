package api

import (
	"encoding/json"
	"net/http"

	"docker-dashboard/internal/containers"

	"github.com/gorilla/mux"
)

func RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/api/containers", getContainersHandler).Methods("GET")
}

func getContainersHandler(w http.ResponseWriter, r *http.Request) {
	containers, err := containers.GetContainers()
	if err != nil {
		http.Error(w, "Failed to get containers", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(containers)
}

