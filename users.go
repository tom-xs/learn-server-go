package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type userRequest struct {
	Email string `json:"email"`
}

type userResponse struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func (cfg *apiConfig) handleUserCreation(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var email userRequest
	decoder := json.NewDecoder(r.Body)

	if err := decoder.Decode(&email); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Error decoding request json: %v", err)
		return
	}

	user, err := cfg.dbQuery.CreateUser(r.Context(), email.Email)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Error adding user to DB: %v", err)
		return
	}

	respondWithJson(w, http.StatusCreated, userResponse{
		user.ID,
		user.CreatedAt,
		user.UpdatedAt,
		user.Email,
	})
}
