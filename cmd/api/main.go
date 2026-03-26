package main

import (
	"log"
	"portal-system/internal/app"
)

// Run with: go run github.com/air-verse/air@latest
func main() {
	application, err := app.New()
	if err != nil {
		log.Fatal(err)
	}

	if err := application.Run(); err != nil {
		log.Fatal(err)
	}
}
