package main

import (
	"log"
	"os"

	"template/internal/app"
)

func main() {
	log.Println("Application starting...")
	application := app.NewApp()
	if err := application.Run(); err != nil {
		log.Printf("Application run failed: %v", err)
		os.Exit(1)
	}
	log.Println("Application finished.")
}
