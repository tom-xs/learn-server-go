package auth

import (
	"github.com/alexedwards/argon2id"
)

var hashParameters = argon2id.DefaultParams

func HashPassword(password string) (string, error) {
	hashedPassword, err := argon2id.CreateHash(password, hashParameters)
	if err != nil {
		return "", err
	}
	return hashedPassword, nil
}

func CheckPassword(password, hash string) (bool, error) {
	return argon2id.ComparePasswordAndHash(password, hash)
}
