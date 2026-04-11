package main

import (
	"log"
	"os"

	"mihomo-manager/internal/app"
)

func main() {
	cfg := app.LoadConfigFromEnv()

	server, err := app.NewServer(cfg)
	if err != nil {
		log.Fatalf("create server: %v", err)
	}

	if err := server.ListenAndServe(); err != nil {
		log.Printf("server stopped: %v", err)
		os.Exit(1)
	}
}
