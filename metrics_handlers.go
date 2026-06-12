package main

import (
	"fmt"
	"net/http"
	"sync/atomic"

	"github.com/tom-xs/learn-server-go/internal/database"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	dbQuery        *database.Queries
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) handleMetricsReset(writer http.ResponseWriter, req *http.Request) {
	writer.WriteHeader(http.StatusOK)
	cfg.fileserverHits.Store(0)
	writer.Write([]byte("Hits reset"))
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
