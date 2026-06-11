package main

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

func (cfg *apiConfig) handleMetricsReset(writer http.ResponseWriter, req *http.Request) {
	writer.WriteHeader(http.StatusOK)
	cfg.fileserverHits.Store(0)
	writer.Write([]byte("Hits reset"))
}

func (cfg *apiConfig) handleMetrics(writer http.ResponseWriter, req *http.Request) {
	req.Header.Set("Content-Type", "text/plain; charset=utf-8")
	writer.WriteHeader(http.StatusOK)

	hits := cfg.fileserverHits.Load()
	msg := fmt.Sprintf("Hits: %d", hits)
	writer.Write([]byte(msg))
}

func handleHealthz(writer http.ResponseWriter, req *http.Request) {
	req.Header.Set("Content-Type", "text/plain; charset=utf-8")
	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte("OK"))
}

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func main() {
	const port = "8080"
	apiCfg := apiConfig{}
	mux := http.NewServeMux()
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir(".")))))
	mux.HandleFunc("/metrics/", apiCfg.handleMetrics)
	mux.HandleFunc("/reset/", apiCfg.handleMetricsReset)
	mux.HandleFunc("/healthz/", handleHealthz)
	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}
	server.ListenAndServe()
	log.Printf("Serving on port: %s", port)
}
