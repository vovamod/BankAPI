package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/vovamod/BankAPI/utils"
	"strings"
)

// TokenClaims represents the expected JWT claims
type TokenClaims struct {
	UserID   int    `json:"userId"`
	UserRole string `json:"userRole"`
	jwt.RegisteredClaims
}

func AuthMiddleware(allowedRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		tokenString := c.Get("Authorization")
		if tokenString == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Missing authentication token"})
		}

		// The token should be prefixed with "Bearer "
		tokenParts := strings.Split(tokenString, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Missing or invalid Authorization header"})
		}

		tokenString = tokenParts[1]

		claims := utils.VerifyToken(tokenString)
		for _, b := range allowedRoles {
			if b == claims["role"] {
				return c.Next()
			}
		}
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid authentication token"})
	}
}
