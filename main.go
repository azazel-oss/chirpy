package main

import "net/http"

func main() {
	serveMux := http.NewServeMux()
	serveMux.Handle("/", http.FileServer(http.Dir(".")))
	server := http.Server{
		Handler: serveMux,
		Addr:    "localhost:8080",
	}
	server.ListenAndServe()
}
