package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type apiConfig struct {
	fileserverHits int
}

func main() {
	apiCfg := &apiConfig{
		fileserverHits: 0,
	}
	mux := http.NewServeMux()
	server := http.Server{
		Handler: mux,
		Addr:    "localhost:8080",
	}
	mux.Handle("/app/*", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))
	mux.HandleFunc("GET /api/healthz", handleReadinessEndpoint)
	mux.HandleFunc("POST /api/validate_chirp", handleValidateChirp)
	mux.HandleFunc("GET /admin/metrics", apiCfg.handleMetricsEndpoint)
	mux.HandleFunc("GET /api/reset", apiCfg.handleResetEndpoint)

	server.ListenAndServe()
}

func handleReadinessEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("Content-Type", "text/plain; charset=utf-8")
	response.WriteHeader(200)
	response.Write([]byte("OK"))
}

func (a *apiConfig) handleMetricsEndpoint(response http.ResponseWriter, _ *http.Request) {
	response.Header().Add("Content-Type", "text/html; charset=utf-8")
	response.WriteHeader(200)
	response.Write([]byte(fmt.Sprintf("<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>", a.fileserverHits)))
}

func (a *apiConfig) handleResetEndpoint(response http.ResponseWriter, _ *http.Request) {
	a.fileserverHits = 0
	response.Header().Add("Content-Type", "text/plain; charset=utf-8")
	response.WriteHeader(200)
	response.Write([]byte(fmt.Sprintf("Hits: %d", a.fileserverHits)))
}

func (a *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.fileserverHits += 1
		next.ServeHTTP(w, r)
	})
}

func handleValidateChirp(w http.ResponseWriter, r *http.Request) {
	type RequestBody struct {
		Body string `json:"body"`
	}

	type ErrorResponse struct {
		Error string `json:"error"`
	}
	type ValidResponse struct {
		Valid bool `json:"valid"`
	}
	bodyJson := RequestBody{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&bodyJson)
	if err != nil {
		// an error will be thrown if the JSON is invalid or has the wrong types
		// any missing fields will simply have their values in the struct set to their zero value
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(500)
		return
	}
	if len(bodyJson.Body) > 140 {
		respBody := ErrorResponse{
			Error: "Chirp is too long",
		}
		dat, err := json.Marshal(respBody)
		if err != nil {
			log.Printf("Error marshalling JSON: %s", err)
			w.WriteHeader(500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(400)
		w.Write(dat)
		return
	}
	respValid := ValidResponse{
		Valid: true,
	}
	dat, err := json.Marshal(respValid)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(dat)
}
