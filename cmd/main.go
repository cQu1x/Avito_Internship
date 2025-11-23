package main

import (
	"AvitoInternship/internal/config"
	"AvitoInternship/internal/handlers"
	"AvitoInternship/internal/repository/db"
	"fmt"
	"log"
	"net/http"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Println("failed to load config:", err)
		return
	}
	host := "localhost"
	port := cfg.DB_HOST_PORT
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.POSTGRES_USER, cfg.POSTGRES_PASSWORD, host, port, cfg.POSTGRES_DB)
	database, err := db.InitDB(dsn)
	if err != nil {
		log.Println("failed to connect to database:", err)
		return
	}
	defer database.Close()
	router := handlers.SetupRouter(database)
	log.Printf("Server starting on port %s", cfg.APP_PORT)
	if err := http.ListenAndServe(":"+cfg.APP_PORT, router); err != nil {
		log.Fatal("failed to start server:", err)
	}
}
