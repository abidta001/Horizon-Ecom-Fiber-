package users

import (
	"horizon/config"
	responsemodels "horizon/models/responsemodels"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func AddAddress(c *fiber.Ctx) error {
	address := new(responsemodels.AddressUser)
	userID := c.Locals("userID").(int)

	if err := c.BodyParser(address); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	address.AddressLine = strings.TrimSpace(address.AddressLine)
	address.City = strings.TrimSpace(address.City)
	address.ZipCode = strings.TrimSpace(address.ZipCode)

	if len(address.AddressLine) < 5 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Address Line must be at least 5 characters long"})
	}

	if len(address.City) < 3 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "City must be at least 3 characters long"})
	}

	if len(address.ZipCode) != 6 || !isNumeric(address.ZipCode) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Zip Code must contain exactly 6 digits"})
	}

	query := `INSERT INTO addresses (user_id, address_line, city, zip_code) VALUES ($1, $2, $3, $4) RETURNING id`
	err := config.DB.QueryRow(query, userID, address.AddressLine, address.City, address.ZipCode).Scan(&address.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to add address"})
	}

	return c.JSON(fiber.Map{"message": "Address added successfully", "address": address})
}

func ViewAddresses(c *fiber.Ctx) error {
	userID := c.Locals("userID").(int)
	var addresses []responsemodels.AddressUser

	query := `SELECT id, address_line, city, zip_code FROM addresses WHERE user_id=$1`
	rows, err := config.DB.Query(query, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch addresses"})
	}
	defer rows.Close()

	for rows.Next() {
		var address responsemodels.AddressUser
		if err := rows.Scan(&address.ID, &address.AddressLine, &address.City, &address.ZipCode); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to scan address"})
		}
		addresses = append(addresses, address)
	}

	return c.JSON(fiber.Map{"message": "Addresses fetched successfully", "addresses": addresses})
}

func EditAddress(c *fiber.Ctx) error {
	address := new(responsemodels.AddressUser)
	userID := c.Locals("userID").(int)

	if err := c.BodyParser(address); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	address.AddressLine = strings.TrimSpace(address.AddressLine)
	address.City = strings.TrimSpace(address.City)
	address.ZipCode = strings.TrimSpace(address.ZipCode)

	if len(address.AddressLine) < 5 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Address Line must be at least 5 characters long"})
	}

	if len(address.City) < 3 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "City must be at least 3 characters long"})
	}

	if len(address.ZipCode) != 6 || !isNumeric(address.ZipCode) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Zip Code must contain exactly 6 digits"})
	}

	query := `UPDATE addresses SET address_line=$1, city=$2, zip_code=$3, updated_at=NOW() WHERE id=$4 AND user_id=$5`
	result, err := config.DB.Exec(query, address.AddressLine, address.City, address.ZipCode, address.ID, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update address"})
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Address not found or unauthorized"})
	}

	return c.JSON(fiber.Map{"message": "Address updated successfully"})
}

func DeleteAddress(c *fiber.Ctx) error {
	addressID := c.Params("id")
	userID := c.Locals("userID").(int)

	query := `DELETE FROM addresses WHERE id=$1 AND user_id=$2`
	result, err := config.DB.Exec(query, addressID, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete address"})
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Address not found or unauthorized"})
	}

	return c.JSON(fiber.Map{"message": "Address deleted successfully"})
}
