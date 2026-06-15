package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/tom-xs/learn-server-go/internal/auth"
)

type userLoginResponse struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
	Token     string    `json:"token"`
}

func (cfg *apiConfig) handleLogin(w http.ResponseWriter, r *http.Request) {
	type LoginRequest struct {
		Email     string `json:"email"`
		Password  string `json:"password"`
		ExpiresIn int    `json:"expires_in_seconds"`
	}

	var loginInfo LoginRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&loginInfo); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Unable to decode JSON request: %v", err)
		return
	}

	expirationTime := loginInfo.ExpiresIn
	if loginInfo.ExpiresIn == 0 || loginInfo.ExpiresIn > 60*60 {
		expirationTime = 60 * 60
	}

	user, err := cfg.dbQuery.SearchUser(r.Context(), loginInfo.Email)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Error while searching for user email on DB: %v", err)
		return
	}

	userExists, err := auth.CheckPasswordHash(loginInfo.Password, user.HashedPassword)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Error while validating/hashing password: %v", err)
		return
	}

	if !userExists {
		w.WriteHeader(http.StatusUnauthorized)
		log.Printf("Requested user not registered")
		return
	}

	token, err := auth.MakeJWT(user.ID, cfg.jwtSecret, time.Duration(expirationTime*int(time.Second)))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Unable to generate JWT Token: %v", err)
		return
	}

	respondWithJson(w, http.StatusOK, userLoginResponse{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
		Token:     token,
	})
}
