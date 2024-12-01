package users

import (
	"horizon/config"
	responsemodels "horizon/models/responsemodels"
	"regexp"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func ShowUserProfile(c *fiber.Ctx) error {
	userID := c.Locals("userID").(int)

	var profile responsemodels.UserProfile
	query := `SELECT  name, email , phone FROM users WHERE id=$1`
	err := config.DB.QueryRow(query, userID).Scan(&profile.Name, &profile.Email, &profile.Phone)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch user profile"})
	}

	return c.JSON(fiber.Map{
		"message": "Profile fetched successfully",
		"user":    profile,
	})
}

func EditUserProfile(c *fiber.Ctx) error {
	userID := c.Locals("userID").(int)

	var updatedProfile struct {
		Name  string `json:"name"`
		Phone string `json:"phone"`
	}

	if err := c.BodyParser(&updatedProfile); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	updatedProfile.Name = strings.TrimSpace(updatedProfile.Name)
	updatedProfile.Phone = strings.TrimSpace(updatedProfile.Phone)

	if len(updatedProfile.Name) < 3 || len(updatedProfile.Name) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Name must be at least 3 characters long and cannot be empty or spaces"})
	}
	if len(updatedProfile.Phone) != 10 || !isNumeric(updatedProfile.Phone) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Phone must be 10 digits long and contain only numbers"})
	}

	updateQuery := `
		UPDATE users 
		SET name = $1, phone = $2 
		WHERE id = $3
	`

	_, err := config.DB.Exec(updateQuery, updatedProfile.Name, updatedProfile.Phone, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update profile"})
	}

	return c.JSON(fiber.Map{
		"message": "Profile updated successfully",
	})
}

func isNumeric(str string) bool {
	re := regexp.MustCompile(`^[0-9]+$`)
	return re.MatchString(str)
}
