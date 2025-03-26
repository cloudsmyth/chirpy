package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/cloudsmyth/chirpy/internal/database"
	_ "github.com/lib/pq"
)

func main() {
	dbUrl := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbUrl)
	if err != nil {
		log.Fatalf("Database connection not made: %v\n", err)
	}
	dbQueries := database.New(db)

	port := "8080"
	mux := http.NewServeMux()

	apiCfg := &apiConfig{
		dbQueries: dbQueries,
	}

	handler := http.StripPrefix("/app", http.FileServer(http.Dir(".")))
	mux.Handle("/app/", apiCfg.incrementHits(handler))
	mux.HandleFunc("GET /api/healthz", healthCheckHandler)
	mux.HandleFunc("GET /admin/metrics", apiCfg.metricShowHandler)
	mux.HandleFunc("POST /admin/reset", apiCfg.metricResetHandler)
	mux.HandleFunc("POST /api/validate_chirp", validateChirpHandler)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Starting server on port: %s\n", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Could not start server: %v", err)
	}
}
