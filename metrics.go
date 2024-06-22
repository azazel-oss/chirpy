package main

import (
	"fmt"
	"net/http"
)

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
