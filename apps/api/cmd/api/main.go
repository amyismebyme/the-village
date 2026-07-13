package main

import (
	"log"
	"net/http"

	"github.com/amyismebyme/the-village/apps/api/internal/server"
)

func main() {

	router := server.NewRouter()

	port := ":8080"

	log.Println("Starting Village API...")
	log.Println("Listening on", port)

	err := http.ListenAndServe(port, router)

	if err != nil {
		log.Fatal(err)
	}
}
