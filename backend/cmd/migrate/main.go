package main

import (
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"affiliatetrack/backend/internal/config"
	"affiliatetrack/backend/internal/db"
)

func main() {
	cfg := config.Load()

	database, err := db.Open(cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	files, err := filepath.Glob("internal/db/migrations/*.sql")
	if err != nil {
		log.Fatal(err)
	}
	sort.Strings(files)

	for _, file := range files {
		sqlBytes, err := os.ReadFile(file)
		if err != nil {
			log.Fatal(err)
		}

		if strings.TrimSpace(string(sqlBytes)) == "" {
			continue
		}

		log.Printf("applying %s", file)
		if _, err := database.Exec(string(sqlBytes)); err != nil {
			log.Fatal(err)
		}
	}

	log.Println("migrations applied")
}
