package main

import (
	"log"
	"net/http"

	"github.com/amyismebyme/the-village/apps/api/internal/server"
	"github.com/amyismebyme/the-village/apps/api/internal/config"

)

func main() {

		// Load application configuration
    	cfg := config.Load()

    	// Build router
    	router := server.NewRouter()

    	// Build listening address
    	address := ":" + cfg.Port

    	log.Println("====================================")
    	log.Println("Starting Village API")
    	log.Println("Environment :", cfg.Environment)
    	log.Println("Log Level   :", cfg.LogLevel)
    	log.Println("Listening   :", address)
    	log.Println("====================================")

    	err := http.ListenAndServe(address, router)
    	if err != nil {
    		log.Fatal(err)
    	}

}
