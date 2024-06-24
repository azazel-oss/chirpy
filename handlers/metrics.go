package handlers

import (
	"fmt"
	"net/http"
)

func (a *ApiConfig) handleMetricsEndpoint(response http.ResponseWriter, _ *http.Request) {
	response.Header().Add("Content-Type", "text/html; charset=utf-8")
	response.WriteHeader(200)
	response.Write([]byte(fmt.Sprintf("<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>", a.FileserverHits)))
}

func (a *ApiConfig) handleResetEndpoint(response http.ResponseWriter, _ *http.Request) {
	a.FileserverHits = 0
	response.Header().Add("Content-Type", "text/plain; charset=utf-8")
	response.WriteHeader(200)
	response.Write([]byte(fmt.Sprintf("Hits: %d", a.FileserverHits)))
}
