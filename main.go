package main

import "net/http"

func main() {
	serveMux := http.ServeMux{}
	server := http.Server{
		Handler: &serveMux,
		Addr:    "localhost:8080",
	}
	server.ListenAndServe()
}
