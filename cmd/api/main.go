package main

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"mangea-backend/internal/config"
	"mangea-backend/internal/db"
	"mangea-backend/internal/server"
)

func main() {
	godotenv.Load()

	cfg := config.Load()

	database, err := db.Connect(cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPass, cfg.DBName)
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer database.Close()

	router := server.NewRouter(server.RouterDeps{DB: database, JWTSecret: cfg.JWTSecret})

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Starting server on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
