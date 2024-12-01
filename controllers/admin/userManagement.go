package admin

import (
	"fmt"
	"horizon/config"
	"horizon/models"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

func ViewUsers(c *fiber.Ctx) error {
	query := `SELECT id, name, email, COALESCE(phone, '') AS phone, verified, blocked FROM users`
	rows, err := config.DB.Query(query)
	if err != nil {
		fmt.Println(err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch users"})
	}
	defer rows.Close()

	var users []models.UserView
	for rows.Next() {
		var user models.UserView

		if err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.Phone, &user.Verified, &user.Blocked); err != nil {
			fmt.Println(err)
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to parse users"})
		}
		users = append(users, user)
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{"users": users})
}

func BlockUser(c *fiber.Ctx) error {
	req := new(models.BlockRequest)
	if err := c.BodyParser(req); err != nil {
		fmt.Println(err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	query := `UPDATE users SET blocked=true WHERE id=$1`
	_, err := config.DB.Exec(query, req.ID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to block user"})
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{"message": "User blocked successfully"})
}

func UnblockUser(c *fiber.Ctx) error {
	req := new(models.BlockRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	query := `UPDATE users SET blocked=false WHERE id=$1`
	_, err := config.DB.Exec(query, req.ID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to unblock user"})
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{"message": "User unblocked successfully"})
}
