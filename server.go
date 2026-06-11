package main

import (
	"log"
	"net/http"
)

func handleHealthz(writer http.ResponseWriter, req *http.Request) {
	req.Header.Set("Content-Type", "text/plain; charset=utf-8")
	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte("OK"))
}

func main() {
	const port = "8080"
	apiCfg := apiConfig{}
	mux := http.NewServeMux()
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir(".")))))
	mux.HandleFunc("POST /admin/reset/", apiCfg.handleMetricsReset)
	mux.HandleFunc("GET /admin/metrics/", apiCfg.handleMetrics)
	mux.HandleFunc("GET /api/healthz/", handleHealthz)
	mux.HandleFunc("POST /api/validate_chirp/", respondJsonPost)
	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}
	server.ListenAndServe()
	log.Printf("Serving on port: %s", port)
}
