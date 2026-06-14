package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/tom-xs/learn-server-go/internal/auth"
	"github.com/tom-xs/learn-server-go/internal/database"
)

type userResponse struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func (cfg *apiConfig) handleUserCreation(w http.ResponseWriter, r *http.Request) {

	type userRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	defer r.Body.Close()

	var userInfo userRequest
	decoder := json.NewDecoder(r.Body)

	if err := decoder.Decode(&userInfo); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Error decoding request json: %v", err)
		return
	}

	if userInfo.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Password missing from request")
		return
	}

	hashedPassword, err := auth.HashPassword(userInfo.Password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Unable to hash password: %v", err)
		return
	}

	createUserParams := database.CreateUserParams{
		Email:          userInfo.Email,
		HashedPassword: hashedPassword,
	}

	user, err := cfg.dbQuery.CreateUser(r.Context(), createUserParams)
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
