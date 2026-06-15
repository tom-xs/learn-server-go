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

type userRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type userCreationResponse struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func (cfg *apiConfig) handleUserUpdate(w http.ResponseWriter, r *http.Request) {
	accessToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		log.Printf("User Unauthorized: %v", err)
		return
	}

	var user userRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&user); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Error while decoding request: %v", err)
		return
	}

	userID, err := auth.ValidateJWT(accessToken, cfg.jwtSecret)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		log.Printf("User not authenticated: %v", err)
		return
	}

	hashedPassword, err := auth.HashPassword(user.Password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Unable to hash password: %v", err)
		return
	}

	updateUserParams := database.UpdateUserParams{
		ID:             userID,
		Email:          user.Email,
		HashedPassword: hashedPassword,
		UpdatedAt:      time.Now(),
	}
	updatedUser, err := cfg.dbQuery.UpdateUser(r.Context(), updateUserParams)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Error while updating DB: %v", err)
		return
	}

	respondWithJson(w, http.StatusOK, userCreationResponse{
		ID:        updatedUser.ID,
		Email:     updatedUser.Email,
		CreatedAt: updatedUser.CreatedAt,
		UpdatedAt: updatedUser.UpdatedAt,
	})
}

func (cfg *apiConfig) handleUserCreation(w http.ResponseWriter, r *http.Request) {

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

	respondWithJson(w, http.StatusCreated, userCreationResponse{
		user.ID,
		user.CreatedAt,
		user.UpdatedAt,
		user.Email,
	})
}
