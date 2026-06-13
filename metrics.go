package main

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"

	"github.com/tom-xs/learn-server-go/internal/database"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	dbQuery        *database.Queries
	platform       string
}

func handleHealthz(writer http.ResponseWriter, req *http.Request) {
	req.Header.Set("Content-Type", "text/plain; charset=utf-8")
	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte("OK"))
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) handleReset(w http.ResponseWriter, r *http.Request) {
	if cfg.platform != "dev" {
		w.WriteHeader(http.StatusForbidden)
		log.Printf("Platform not available for request")
		return
	}

	if err := cfg.dbQuery.DeleteUsers(r.Context()); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Something wrong happened while reseting users db: %v", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	cfg.fileserverHits.Store(0)
	w.Write([]byte("Server fully reset"))
}

func (cfg *apiConfig) handleMetrics(writer http.ResponseWriter, req *http.Request) {
	req.Header.Set("Content-Type", "text/html")
	writer.WriteHeader(http.StatusOK)

	hits := cfg.fileserverHits.Load()
	msg := fmt.Sprintf(`<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`, hits)
	writer.Write([]byte(msg))
}
