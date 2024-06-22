package main

import (
	"chirpy/internal/database"
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

type apiConfig struct {
	fileserverHits int
	currentId      int
}

type ApiState struct {
	database *database.DB
}

func main() {
	apiCfg := &apiConfig{
		fileserverHits: 0,
	}
	db, err := database.NewDB("database.json")
	if err != nil {
		log.Fatal("Database crashed:", err)
	}
	apiState := &ApiState{
		database: db,
	}
	mux := http.NewServeMux()
	server := http.Server{
		Handler: mux,
		Addr:    "localhost:8080",
	}

	mux.Handle("/app/*", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))
	mux.HandleFunc("/api/healthz", handleReadinessEndpoint)
	mux.HandleFunc("/api/chirps", apiState.handleChirps)
	mux.HandleFunc("/admin/metrics", apiCfg.handleMetricsEndpoint)
	mux.HandleFunc("/api/reset", apiCfg.handleResetEndpoint)

	log.Println("Starting server on :8080")
	server.ListenAndServe()
}

func (a *ApiState) handleChirps(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		a.createChirps(w, r)
	case "GET":
		a.fetchChirps(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (a *ApiState) createChirps(w http.ResponseWriter, r *http.Request) {
	type RequestBody struct {
		Body string `json:"body"`
	}

	bodyJson := RequestBody{}
	err := json.NewDecoder(r.Body).Decode(&bodyJson)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	if len(bodyJson.Body) > 140 {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}
	chunks := strings.Split(bodyJson.Body, " ")
	profanes := []string{"kerfuffle", "sharbert", "fornax"}
	for i, chunk := range chunks {
		for _, profane := range profanes {
			if strings.EqualFold(strings.ToLower(chunk), strings.ToLower(profane)) {
				chunks[i] = "****"
			}
		}
	}
	chirp, err := a.database.CreateChirp(strings.Join(chunks, " "))
	if err != nil {
		respondWithError(w, 500, "your mother is a whore")
		return
	}
	respondWithJson(w, 201, chirp)
}

func (a *ApiState) fetchChirps(w http.ResponseWriter, _ *http.Request) {
	chirps, err := a.database.GetChirps()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	respondWithJson(w, http.StatusOK, chirps)
}
