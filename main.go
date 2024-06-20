package main

import (
	"fmt"
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
	mux.HandleFunc("/healthz", handleReadinessEndpoint)
	mux.HandleFunc("/metrics", apiCfg.handleMetricsEndpoint)
	mux.HandleFunc("/reset", apiCfg.handleResetEndpoint)

	server.ListenAndServe()
}

func handleReadinessEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("Content-Type", "text/plain; charset=utf-8")
	response.WriteHeader(200)
	response.Write([]byte("OK"))
}

func (a *apiConfig) handleMetricsEndpoint(response http.ResponseWriter, _ *http.Request) {
	response.Header().Add("Content-Type", "text/plain; charset=utf-8")
	response.WriteHeader(200)
	response.Write([]byte(fmt.Sprintf("Hits: %d", a.fileserverHits)))
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
