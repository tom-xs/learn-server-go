package auth

import (
	"errors"
	"log"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var hashParameters = argon2id.DefaultParams
var tokenTypeAccess = "chirpy-access"

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

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	signingKey := []byte(tokenSecret)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    tokenTypeAccess,
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
		Subject:   userID.String(),
	})
	return token.SignedString(signingKey)
}
