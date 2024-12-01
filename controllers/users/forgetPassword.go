package users

import (
	"horizon/config"
	"horizon/middlewares"
	"horizon/utils"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

func ForgotPassword(c *fiber.Ctx) error {
	email := c.FormValue("email")

	var userID int
	query := `SELECT id FROM users WHERE email=$1`
	err := config.DB.QueryRow(query, email).Scan(&userID)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Email not registered"})
	}

	resetToken, err := middleware.GenerateResetToken(userID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate reset token"})
	}

	err = utils.SendResetEmail(email, resetToken)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to send reset email"})
	}

	return c.JSON(fiber.Map{"message": "Reset email sent successfully"})
}
func ResetPassword(c *fiber.Ctx) error {
	token := c.Query("token")
	newPassword := c.FormValue("password")

	userID, err := middleware.ValidateResetToken(token)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid or expired token"})
	}

	query := `UPDATE users SET password=$1 WHERE id=$2`
	_, err = config.DB.Exec(query, utils.HashPassword1(newPassword), userID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to reset password"})
	}

	return c.JSON(fiber.Map{"message": "Password reset successfully"})
}
