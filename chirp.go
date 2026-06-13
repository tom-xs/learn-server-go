package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/tom-xs/learn-server-go/internal/database"
)

var ErrLongChirp = errors.New("Chirp exceeds 140 length")

type chirpResponse struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) handleChirpCreation(w http.ResponseWriter, r *http.Request) {
	type chirpRequest struct {
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}

	defer r.Body.Close()

	var chirp chirpRequest
	decoder := json.NewDecoder(r.Body)

	if err := decoder.Decode(&chirp); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Error decoding request json: %v", err)
		return
	}

	err := validateChirp(chirp.Body)
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

	parameters := database.CreateChirpParams{
		Body:   chirp.Body,
		UserID: chirp.UserID,
	}

	user, err := cfg.dbQuery.CreateChirp(r.Context(), parameters)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Error adding user to DB: %v", err)
		return
	}

	respondWithJson(w, http.StatusCreated, chirpResponse{
		user.ID,
		user.CreatedAt,
		user.UpdatedAt,
		user.Body,
		user.UserID,
	})
}

func (cfg *apiConfig) handleChirpRequest(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	chirps, err := cfg.dbQuery.GetAllChirps(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Error requesting chirps to DB: %v", err)
		return
	}

	respondWithJsonArray(w, http.StatusOK, chirps)
}

func validateChirp(body string) error {
	if len(body) > 140 {
		return ErrLongChirp
	}
	return nil
}
