package handlers

import "net/http"

func handleReadinessEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("Content-Type", "text/plain; charset=utf-8")
	response.WriteHeader(http.StatusOK)
	response.Write([]byte("OK"))
}
