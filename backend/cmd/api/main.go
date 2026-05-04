package main

import (
	"log"

	"affiliatetrack/backend/internal/config"
	"affiliatetrack/backend/internal/db"
	"affiliatetrack/backend/internal/server"
)

func main() {
	cfg := config.Load()

	database, err := db.Open(cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	router := server.NewRouter(cfg, database)

	log.Printf("backend listening on %s", cfg.HTTPAddr)
	if err := router.Run(cfg.HTTPAddr); err != nil {
		log.Fatal(err)
	}
}
