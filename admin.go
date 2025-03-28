package main

import (
	"fmt"
	"net/http"
	"sync/atomic"

	"github.com/cloudsmyth/chirpy/internal/database"
)

type apiConfig struct {
	fileServerHits atomic.Int32
	dbQueries      *database.Queries
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (cfg *apiConfig) metricShowHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	metric := cfg.fileServerHits.Load()
	metrics := fmt.Sprintf(`<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>
`, metric)
	w.Write([]byte(metrics))
}

func (cfg *apiConfig) metricResetHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	cfg.fileServerHits.Store(0)
	w.Write([]byte("Counter reset to 0"))
}

func (cfg *apiConfig) incrementHits(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileServerHits.Add(1)
		next.ServeHTTP(w, r)
	})
}
