package utils

import (
	"context"
	"errors"
	"github.com/gofiber/fiber/v2/log"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
	"time"
)

// jwtSecret needed for JWT encryption (make it about 256 char long with some stuff)
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

// MongoDatabase create a mongoDatabase pointer and off you go!
func MongoDatabase() *mongo.Database {
	log.Info("Connecting to MongoDB")
	// Init client, mongo driver Client.
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(os.Getenv("MONGODB_URI")))
	if err != nil {
		panic(err)
	}
	// Check whether DB is alive and ping still works
	if c := client.Ping(nil, nil); c != nil {
		log.Fatal("Cannot connect to MongoDB. Error: " + c.Error())
	}
	log.Info("Connected to MongoDB")
	return client.Database(os.Getenv("MONGODB_DATABASE"))
}

// CheckEnv .env for var
func CheckEnv() {
	log.Info("Checking environment variables")
	// Create sorta List obj to store our things
	vars := map[string]string{
		"MONGODB_URI":      "MongoDB URI (MONGODB_URI)",
		"JWT_SECRET":       "JWT Secret (JWT_SECRET)",
		"BANK_256_CODE":    "Bank 256 Code (BANK_256_CODE)",
		"MONGODB_DATABASE": "MongoDB Database (MONGODB_DATABASE)",
		"ADDR":             "Address (ADDR)",
	}
	var missingVars []string

	for key, name := range vars {
		if _, exists := os.LookupEnv(key); !exists {
			missingVars = append(missingVars, name)
		}
	}
	if len(missingVars) > 0 {
		log.Fatal("Missing environment variables:", missingVars)
	}
}
