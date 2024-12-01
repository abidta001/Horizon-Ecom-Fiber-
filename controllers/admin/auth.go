package admin

import (
	"horizon/config"
	"horizon/models"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

func AdminLogin(c *fiber.Ctx) error {
	req := new(models.LoginRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	var admin models.Admin
	query := `SELECT username, password FROM admins WHERE username=$1`
	err := config.DB.QueryRow(query, req.Username).Scan(&admin.Username, &admin.Password)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	err = bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(req.Password))
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	claims := jwt.StandardClaims{
		Subject:   "admin",
		Issuer:    "horizon-app",
		ExpiresAt: time.Now().Add(time.Hour * 3).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(config.JWT_SECRET_ADMIN))
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate token"})
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{"token": tokenString})
}
