package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"slices"
	"strings"
)

var ErrLongChirp = errors.New("Chirp exceeds 140 length")

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

	err := validateChirp(req.Body)
	if err != nil {
		switch {
		case errors.Is(err, ErrLongChirp):
			respondWithError(w, http.StatusBadRequest, "Chirp is too long")
		default:
			log.Printf("Unexpected validation error: %v", err)
			respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		}
		return
	}

	respondWithJSON(w, http.StatusOK, jsonResponse{
		CleanedBody: filterProfane(req.Body),
	})
}

func validateChirp(body string) error {
	if len(body) > 140 {
		return ErrLongChirp
	}
	return nil
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	respondWithJSON(w, code, errorResponse{
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

func respondWithJSON(w http.ResponseWriter, code int, payload any) {
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
