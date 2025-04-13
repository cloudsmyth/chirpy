package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/cloudsmyth/chirpy/internal/api"
	"github.com/cloudsmyth/chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	godotenv.Load()

	platform := os.Getenv("PLATFORM")
	dbUrl := os.Getenv("DB_URL")
	jwtSecret := os.Getenv("JWT_SECRET")
	polkaKey := os.Getenv("POLKA_KEY")

	db, err := sql.Open("postgres", dbUrl)
	if err != nil {
		log.Fatalf("Database connection not made: %v\n", err)
	}
	dbQueries := database.New(db)

	port := "8080"
	mux := http.NewServeMux()

	apiCfg := &api.ApiConfig{
		DbQueries: dbQueries,
		Platform:  platform,
		Secret:    jwtSecret,
		Polka:     polkaKey,
	}

	handler := http.StripPrefix("/app", http.FileServer(http.Dir(".")))
	mux.Handle("/app/", apiCfg.IncrementHits(handler))
	mux.HandleFunc("GET /api/healthz", healthCheckHandler)
	mux.HandleFunc("GET /admin/metrics", apiCfg.MetricShowHandler)
	mux.HandleFunc("POST /admin/reset", apiCfg.MetricResetHandler)
	mux.HandleFunc("POST /api/chirps", apiCfg.CreateChirpsHandler)
	mux.HandleFunc("POST /api/users", apiCfg.AddUserHandler)
	mux.HandleFunc("PUT /api/users", apiCfg.UpdateUserHandler)
	mux.HandleFunc("GET /api/chirps", apiCfg.GetChirpsHandler)
	mux.HandleFunc("GET /api/chirps/{chirpId}", apiCfg.GetChirpByIdHandler)
	mux.HandleFunc("POST /api/login", apiCfg.LoginHandler)
	mux.HandleFunc("POST /api/refresh", apiCfg.RefreshHandler)
	mux.HandleFunc("POST /api/revoke", apiCfg.RevokeHandler)
	mux.HandleFunc("DELETE /api/chirps/{chirpId}", apiCfg.DeleteChirpsHandler)
	mux.HandleFunc("POST /api/polka/webhooks", apiCfg.UpgradeChirpyRedHandler)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Starting server on port: %s\n", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Could not start server: %v", err)
	}
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
