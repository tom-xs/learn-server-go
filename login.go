package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/tom-xs/learn-server-go/internal/auth"
)

func (cfg *apiConfig) handleLogin(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	type LoginRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var loginInfo LoginRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&loginInfo); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Unable to decode JSON request: %v", err)
		return
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

	respondWithJson(w, http.StatusOK, userResponse{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	})
}
