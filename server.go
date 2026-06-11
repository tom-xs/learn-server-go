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
	mux := http.NewServeMux()
	mux.Handle("/app/", http.StripPrefix("/app/", http.FileServer(http.Dir("."))))
	mux.HandleFunc("/healthz/", handleHealthz)
	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}
	server.ListenAndServe()
	log.Printf("Serving on port: %s", port)
}
