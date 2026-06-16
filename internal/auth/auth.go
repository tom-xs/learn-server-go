package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var hashParameters = argon2id.DefaultParams
var tokenTypeAccess = "chirpy-access"
var refreshTokenLenght = 32 // 32 Bytes = 256 Bits

func MakeRefreshToken() string {
	key := make([]byte, refreshTokenLenght)
	rand.Read(key)
	return hex.EncodeToString(key)
}

func HashPassword(password string) (string, error) {
	hashedPassword, err := argon2id.CreateHash(password, hashParameters)
	if err != nil {
		return "", err
	}
	return hashedPassword, nil
}

func CheckPasswordHash(password, hash string) (bool, error) {
	return argon2id.ComparePasswordAndHash(password, hash)
}

func GetAPIKey(headers http.Header) (string, error) {
	headerString := headers.Get("Authorization")
	key, found := strings.CutPrefix(headerString, "ApiKey ")
	if !found {
		return "", errors.New("ApiKey not found")
	}
	return key, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	headerString := headers.Get("Authorization")
	token, found := strings.CutPrefix(headerString, "Bearer ")
	if !found {
		return "", errors.New("Bearer token not found")
	}
	return token, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	claimStruct := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claimStruct, func(token *jwt.Token) (any, error) {
		return []byte(tokenSecret), nil
	})
	if err != nil {
		log.Printf("Error while parsing JWT Token")
		return uuid.Nil, err
	}

	userIdString, err := token.Claims.GetSubject()
	if err != nil {
		log.Printf("Error while obtaining JWT token subject: %v", err)
		return uuid.Nil, err
	}

	issuer, err := token.Claims.GetIssuer()
	if err != nil {
		log.Printf("Error while obtaining Issuer from JWT: %v", err)
		return uuid.Nil, err
	}

	if issuer != tokenTypeAccess {
		return uuid.Nil, errors.New("invalid issuer")
	}

	id, err := uuid.Parse(userIdString)
	if err != nil {
		log.Printf("Error while parsing UUID: %v", err)
		return uuid.Nil, err
	}
	return id, nil
}

func MakeJWT(userID uuid.UUID, tokenSecret string) (string, error) {
	signingKey := []byte(tokenSecret)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    tokenTypeAccess,
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		Subject:   userID.String(),
	})
	return token.SignedString(signingKey)
}
