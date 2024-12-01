package middleware

import (
	"horizon/config"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
)

func AdminJWT(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Authorization header required"})
	}

	parts := strings.Split(authHeader, "Bearer ")
	if len(parts) != 2 {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Malformed Authorization header"})
	}

	tokenString := parts[1]

	claims := &jwt.StandardClaims{}
	_, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.JWT_SECRET_ADMIN), nil
	})
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid or expired token"})
	}

	if claims.Subject != "admin" {
		return c.Status(http.StatusForbidden).JSON(fiber.Map{"error": "Insufficient permissions"})
	}

	return c.Next()
}
