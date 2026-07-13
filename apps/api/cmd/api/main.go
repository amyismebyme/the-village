package main

import (
	"encoding/json"
	"log"
	"net/http"
	"github.com/amyismebyme/the-village/apps/api/internal/handlers"
)

type Response struct {
	Service string `json:"service"`
	Status  string `json:"status"`
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	response := Response{
		Service: "village-api",
		Status:  "running",
	}

	w.Header().Set("Content-Type", "application/json")

	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func main() {

	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/health", handlers.HealthHandler)

	port := ":8080"

	log.Println("Starting Village API...")
	log.Println("Listening on", port)

	err := http.ListenAndServe(port, nil)

	if err != nil {
		log.Fatal(err)
	}
}