package utils

import (
	"errors"
	"github.com/gofiber/fiber/v2/log"
	"github.com/golang-jwt/jwt/v5"
	"os"
	"time"
)

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

// VerifyToken verifies a token JWT validate
func VerifyToken(tokenString string) jwt.MapClaims {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtSecret, nil
	})
	if err != nil {
		return nil
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims
	}
	return nil
}

// GenerateToken used for JWT token, pass userID, role (in case of bank it is the name of it), and time (prefer 72 hours)
func GenerateToken(userID, role string, tm time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"exp":     time.Now().Add(tm * time.Hour).Unix(),
	})
	t, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}
	return t, nil
}

// CheckEnv .env for var
func CheckEnv() {
	errA := os.Getenv("MONGODB_URI")
	errB := os.Getenv("JWT_SECRET")
	errC := os.Getenv("BANK_256_CODE")
	if errA == "" || errB == "" || errC == "" {
		log.Fatal("Missing required environment variables.")
	}
}
