package api

import (
	"encoding/json"
	"net/http"

	"docker-dashboard/internal/containers"
	"docker-dashboard/internal/hostinfo"

	"github.com/gorilla/mux"
)

func RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/api/containers", getContainersHandler).Methods("GET")
	// Новый эндпоинт для системных метрик
	r.HandleFunc("/api/hostinfo", getHostInfoHandler).Methods("GET")
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

func getHostInfoHandler(w http.ResponseWriter, r *http.Request) {
	metrics, err := hostinfo.GetSystemMetrics()
	if err != nil {
		http.Error(w, "Failed to get system metrics", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}
