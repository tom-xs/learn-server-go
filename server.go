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
	jwtSecret := os.Getenv("SECRET")
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
		jwtSecret,
	}
	mux := http.NewServeMux()
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir(".")))))
	mux.HandleFunc("GET /api/healthz/", handleHealthz)
	mux.HandleFunc("GET /admin/metrics/", apiCfg.handleMetrics)
	mux.HandleFunc("GET /api/chirps/", apiCfg.handleAllChirpRequest)
	mux.HandleFunc("GET /api/chirps/{id}", apiCfg.handleChirpRequest)
	mux.HandleFunc("POST /admin/reset/", apiCfg.handleReset)
	mux.HandleFunc("POST /api/validate_chirp/", respondJsonPost)
	mux.HandleFunc("POST /api/users/", apiCfg.handleUserCreation)
	mux.HandleFunc("PUT /api/users/", apiCfg.handleUserUpdate)
	mux.HandleFunc("POST /api/chirps/", apiCfg.handleChirpCreation)
	mux.HandleFunc("POST /api/login/", apiCfg.handleLogin)
	mux.HandleFunc("POST /api/refresh/", apiCfg.handleTokenRefresh)
	mux.HandleFunc("POST /api/revoke/", apiCfg.handleTokenRevoke)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}
	server.ListenAndServe()
	log.Printf("Serving on port: %s", port)
}
