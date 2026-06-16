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

type jwtTokenResponse struct {
	Token string `json:"token"`
}

type userLoginResponse struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
	IsUserRed    bool      `json:"is_chirpy_red"`
}

func (cfg *apiConfig) handleTokenRevoke(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		log.Printf("Error while getting Bearer Token: %v", err)
		return
	}

	if err := cfg.dbQuery.RevokeToken(r.Context(), refreshToken); err != nil {
		w.WriteHeader(http.StatusForbidden)
		log.Printf("Unable to revoke requested token: %v", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (cfg *apiConfig) handleTokenRefresh(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		log.Printf("Error while getting Bearer Token: %v", err)
		return
	}

	token, err := cfg.dbQuery.SelectRefreshToken(r.Context(), refreshToken)
	tokenValid := !token.RevokedAt.Valid || time.Now().After(token.ExpiresAt)

	log.Printf("token: %v", token)

	// Refresh token expired
	if !tokenValid {
		w.WriteHeader(http.StatusUnauthorized)
		log.Printf("Refresh token revoked: %v", err)
		return
	}

	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		log.Printf("Error while getting Bearer Token: %v", err)
		return
	}

	jwt, err := auth.MakeJWT(token.UserID, cfg.jwtSecret)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		log.Printf("Error while getting Bearer Token: %v", err)
		return
	}

	respondWithJson(w, http.StatusOK, jwtTokenResponse{
		Token: jwt,
	})
}

func (cfg *apiConfig) handleLogin(w http.ResponseWriter, r *http.Request) {
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

	token, err := auth.MakeJWT(user.ID, cfg.jwtSecret)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Unable to generate JWT Token: %v", err)
		return
	}

	params := database.CreateRefreshTokenParams{
		Token:  auth.MakeRefreshToken(),
		UserID: user.ID,
	}

	refreshToken, err := cfg.dbQuery.CreateRefreshToken(r.Context(), params)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("unable to generate refresh Token: %v", err)
		return
	}

	respondWithJson(w, http.StatusOK, userLoginResponse{
		ID:           user.ID,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		Token:        token,
		RefreshToken: refreshToken.Token,
		IsUserRed:    user.IsChirpyRed.Bool,
	})
}
