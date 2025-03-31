package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/cloudsmyth/chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileServerHits atomic.Int32
	dbQueries      *database.Queries
	platform       string
}

func main() {
	godotenv.Load()
	platform := os.Getenv("PLATFORM")
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
		platform:  platform,
	}

	handler := http.StripPrefix("/app", http.FileServer(http.Dir(".")))
	mux.Handle("/app/", apiCfg.incrementHits(handler))
	mux.HandleFunc("GET /api/healthz", healthCheckHandler)
	mux.HandleFunc("GET /admin/metrics", apiCfg.metricShowHandler)
	mux.HandleFunc("POST /admin/reset", apiCfg.metricResetHandler)
	mux.HandleFunc("POST /api/chirps", apiCfg.chirpsHandler)
	mux.HandleFunc("POST /api/users", apiCfg.addUserHandler)
	mux.HandleFunc("GET /api/chirps", apiCfg.getChirpsHandler)
	mux.HandleFunc("GET /api/chirps/{chirpId}", apiCfg.getChirpByIdHandler)
	mux.HandleFunc("POST /api/login", apiCfg.loginHandler)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Starting server on port: %s\n", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Could not start server: %v", err)
	}
}
