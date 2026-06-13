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

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Printf("Unable to connect DB: %v", err)
	}
	dbQueries := database.New(db)

	const port = "8080"
	apiCfg := apiConfig{
		atomic.Int32{},
		dbQueries,
		platform,
	}
	mux := http.NewServeMux()
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir(".")))))
	mux.HandleFunc("GET /api/healthz/", handleHealthz)
	mux.HandleFunc("GET /admin/metrics/", apiCfg.handleMetrics)
	mux.HandleFunc("POST /admin/reset/", apiCfg.handleReset)
	mux.HandleFunc("POST /api/validate_chirp/", respondJsonPost)
	mux.HandleFunc("POST /api/users/", apiCfg.handleUserCreation)
	mux.HandleFunc("POST /api/chirp/", apiCfg.handleChirpCreation)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}
	server.ListenAndServe()
	log.Printf("Serving on port: %s", port)
}
