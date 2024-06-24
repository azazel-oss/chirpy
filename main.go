package main

import (
	"chirpy/handlers"
	"chirpy/internal/database"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	// by default, godotenv will look for a file named .env in the current directory
	godotenv.Load()
	jwtSecret := os.Getenv("JWT_SECRET_KEY")
	db, err := database.NewDB("database.json")
	if err != nil {
		log.Fatal("Database crashed:", err)
	}
	apiCfg := &handlers.ApiConfig{
		FileserverHits: 0,
		JwtSecret:      jwtSecret,
		Database:       db,
	}
	mux := http.NewServeMux()
	server := http.Server{
		Handler: mux,
		Addr:    "localhost:8080",
	}
	handlers.RegisterRoutes(mux, apiCfg)

	log.Println("Starting server on :8080")
	server.ListenAndServe()
}
