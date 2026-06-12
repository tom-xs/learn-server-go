package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/tom-xs/learn-server-go/internal/database"
)

func handleHealthz(writer http.ResponseWriter, req *http.Request) {
	req.Header.Set("Content-Type", "text/plain; charset=utf-8")
	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte("OK"))
}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Printf("Unable to connect DB: %v", err)
	}
	dbQueries := database.New(db)

	const port = "8080"
	apiCfg := apiConfig{
		atomic.Int32{},
		dbQueries,
	}
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
