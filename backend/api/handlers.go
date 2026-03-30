package api

import (
	"encoding/json"
	"net/http"
	"time"
)

func GetServices(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ServicesResponse{
		Services:  []Service{},
		Timestamp: time.Now().Unix(),
	})
}

func GetHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(HealthResponse{
		Status:    "ok",
		Timestamp: time.Now().Unix(),
	})
}
