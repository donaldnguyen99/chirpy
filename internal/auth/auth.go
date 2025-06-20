package auth

import (
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	hashed_password_byte_arr, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashed_password_byte_arr), nil
}

func CheckPasswordHash(hash, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err
}