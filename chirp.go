package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/tom-xs/learn-server-go/internal/auth"
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

func (cfg *apiConfig) handleChirpDeletion(w http.ResponseWriter, r *http.Request) {
	chirpID := r.PathValue("id")

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		log.Printf("User not authenticated: %v", err)
		return
	}

	chirpUUID, err := uuid.Parse(chirpID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Unable to parse UUID: %v", err)
		return
	}
	chirp, err := cfg.dbQuery.GetChirp(r.Context(), chirpUUID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Unable to query DB for requested Chirp ID: %v", err)
		return
	}

	jwtUUID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Unable to retrieve user UUID from JWT: %v", err)
		return
	}

	if jwtUUID != chirp.UserID {
		w.WriteHeader(http.StatusForbidden)
		log.Printf("JWT doesn't refer to user that created requested chirp: %v", err)
		return
	}

	cfg.dbQuery.DeleteChirp(r.Context(), chirpUUID)
	w.WriteHeader(http.StatusNoContent)
}

func (cfg *apiConfig) handleChirpCreation(w http.ResponseWriter, r *http.Request) {
	type chirpRequest struct {
		Body string `json:"body"`
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT")
		return
	}
	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT")
		return
	}

	var chirp chirpRequest
	decoder := json.NewDecoder(r.Body)

	if err := decoder.Decode(&chirp); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Error decoding request json: %v", err)
		return
	}

	err = validateChirp(chirp.Body)
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
		UserID: userID,
	}

	user, err := cfg.dbQuery.CreateChirp(r.Context(), parameters)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Error adding chirp to DB: %v", err)
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

	uuidString := r.PathValue("id")

	if uuidString == "" {
		w.WriteHeader(http.StatusNotFound)
		log.Printf("Chirp related to given UUID not found")
		return
	}

	uuid, err := uuid.Parse(uuidString)
	if err != nil {
		return
	}

	chirp, err := cfg.dbQuery.GetChirp(r.Context(), uuid)
	switch {
	case err == sql.ErrNoRows:
		w.WriteHeader(http.StatusNotFound)
		log.Printf("Chirp with corresponding UUID not found: %v", err)
		return
	case err != nil:
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Error requesting chirp to DB: %v", err)
		return
	}

	respondWithJson(w, http.StatusOK, chirpResponse{
		ID:        chirp.ID,
		Body:      chirp.Body,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		UserID:    chirp.UserID,
	})
}

func (cfg *apiConfig) handleAllChirpRequest(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var chirps []database.Chirp
	var err error

	if authorID := r.URL.Query().Get("author_id"); authorID != "" {
		authorUUID, err := uuid.Parse(authorID)
		if err != nil {
			log.Printf("Error obtaining user UUID: %v", err)
			return
		}
		chirps, err = cfg.dbQuery.GetChirpsFrom(r.Context(), authorUUID)
	} else {
		chirps, err = cfg.dbQuery.GetAllChirps(r.Context())
	}

	if parameter := r.URL.Query().Get("sort"); parameter == "desc" {
		sort.Slice(chirps, func(i, j int) bool {
			return chirps[i].CreatedAt.After(chirps[j].CreatedAt)
		})
	}

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
