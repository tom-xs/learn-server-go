package main

import (
	"encoding/json"
	"log"
	"net/http"
	"slices"
	"strings"

	"github.com/tom-xs/learn-server-go/internal/database"
)

type chirpRequest struct {
	Body string `json:"body"`
}

type errorResponse struct {
	Error string `json:"error"`
}

type jsonResponse struct {
	CleanedBody string `json:"cleaned_body"`
}

func respondJsonPost(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)
	var req chirpRequest

	if err := decoder.Decode(&req); err != nil {
		log.Printf("Error decoding payload: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	respondWithJson(w, http.StatusOK, jsonResponse{
		CleanedBody: filterProfane(req.Body),
	})
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	respondWithJson(w, code, errorResponse{
		Error: msg,
	})
}

func filterProfane(msg string) string {
	profaneWords := []string{"kerfuffle", "sharbert", "fornax"}

	splitMsg := strings.Split(msg, " ")
	var filteredMsg []string

	for idx, word := range splitMsg {
		lowerCasedWord := strings.ToLower(word)
		if slices.Contains(profaneWords, lowerCasedWord) {
			filteredMsg = append(filteredMsg, "****")
		} else {
			filteredMsg = append(filteredMsg, splitMsg[idx])
		}
	}
	return strings.Join(filteredMsg, " ")
}

func respondWithJsonArray(w http.ResponseWriter, code int, payload []database.Chirp) {
	responseArray := make([]chirpResponse, 0, len(payload))

	for _, p := range payload {
		response := chirpResponse{
			ID:        p.ID,
			Body:      p.Body,
			CreatedAt: p.CreatedAt,
			UpdatedAt: p.UpdatedAt,
			UserID:    p.UserID,
		}
		responseArray = append(responseArray, response)
	}

	data, err := json.Marshal(responseArray)
	if err != nil {
		log.Printf("Error marshaling JSON: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(data)
}

func respondWithJson(w http.ResponseWriter, code int, payload any) {
	response, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshaling JSON: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
