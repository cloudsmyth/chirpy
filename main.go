package main

import (
	"log"
	"net/http"
)

func main() {
	port := "8080"
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(".")))
	server := http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Starting server on port: %s\n", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Could not start server: %v", err)
	}
}
