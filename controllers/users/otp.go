package users

import (
	"horizon/config"
	"horizon/models"
	"horizon/utils"
	"log"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

func ResendOTP(c *fiber.Ctx) error {
	req := new(models.OTP)
	if err := c.BodyParser(req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	err := utils.ResendOTP(req.Email)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{"message": "OTP resent successfully"})
}

func VerifyOTP(c *fiber.Ctx) error {
	req := new(models.OTP)
	if err := c.BodyParser(req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	if !utils.ValidateOTP(req.Email, req.OTP) {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid or expired OTP"})
	}

	query := `UPDATE users SET verified=true WHERE email=$1`
	if _, err := config.DB.Exec(query, req.Email); err != nil {
		log.Printf("Failed to update user verification: %v", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to verify account"})
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{"message": "OTP verified successfully"})
}
