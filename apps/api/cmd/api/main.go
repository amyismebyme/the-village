package main

import (
	"log"

	"github.com/amyismebyme/the-village/apps/api/internal/app"
)

func main() {

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}

}
