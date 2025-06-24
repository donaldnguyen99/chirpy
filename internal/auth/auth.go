package auth

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
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

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	nowUTC := time.Now().UTC()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer: "chirpy",
		Subject: userID.String(),
		ExpiresAt: jwt.NewNumericDate(nowUTC.Add(expiresIn)),
		IssuedAt: jwt.NewNumericDate(nowUTC),
	})

	signedJWT, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", err
	}
	 return signedJWT, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(tokenSecret), nil
	})
	if err != nil {
		return uuid.Nil, err
	}
	if !token.Valid {
		return uuid.Nil, fmt.Errorf("invalid token")
	}

	uuidString, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.Nil, err
	}

	uuidRes, err := uuid.Parse(uuidString)
	if err != nil {
		return uuid.Nil, err
	}
	return uuidRes, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	tokenString, found := strings.CutPrefix(authHeader, "Bearer ")
	if !found {
		return "", fmt.Errorf("error: missing 'Bearer ' prefix")
	}

	return strings.TrimSpace(tokenString), nil
}