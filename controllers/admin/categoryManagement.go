package admin

import (
	"fmt"
	"horizon/config"
	"horizon/models"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func AddCategory(c *fiber.Ctx) error {
	category := new(models.Category)

	if err := c.BodyParser(category); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}
	category.Name = strings.TrimSpace(category.Name)

	if len(category.Name) < 3 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Category name must be at least 3 characters long"})
	}

	query := `INSERT INTO categories (name, description) VALUES ($1, $2) RETURNING id`
	err := config.DB.QueryRow(query, category.Name, category.Description).Scan(&category.ID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to add category"})
	}

	return c.Status(http.StatusCreated).JSON(fiber.Map{"message": "Category added successfully", "category": category})
}
func EditCategory(c *fiber.Ctx) error {
	category := new(models.Category)

	if err := c.BodyParser(category); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	category.Name = strings.TrimSpace(category.Name)

	if len(category.Name) < 3 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Category name must be at least 3 characters long and cannot be just spaces"})
	}

	query := `UPDATE categories SET name=$1, description=$2 WHERE id=$3 AND deleted=false`
	result, err := config.DB.Exec(query, category.Name, category.Description, category.ID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update category"})
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Category not found or already deleted"})
	}

	return c.JSON(fiber.Map{"message": "Category updated successfully"})
}

func SoftDeleteCategory(c *fiber.Ctx) error {
	categoryID := c.Params("id")
	query := `UPDATE categories SET deleted=true WHERE id=$1`
	_, err := config.DB.Exec(query, categoryID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete category"})
	}

	return c.JSON(fiber.Map{"message": "Category deleted successfully"})
}
func RecoverCategory(c *fiber.Ctx) error {
	categoryID := c.Params("id")
	query := `UPDATE categories SET deleted=false WHERE id=$1`
	_, err := config.DB.Exec(query, categoryID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to recover category"})
	}
	return c.JSON(fiber.Map{"message": "Category retreived successfully!"})
}

func AdminViewCategories(c *fiber.Ctx) error {
	query := `SELECT id, name, description,deleted FROM categories `
	rows, err := config.DB.Query(query)
	if err != nil {
		fmt.Println(err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch categories"})
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var category models.Category
		if err := rows.Scan(&category.ID, &category.Name, &category.Description, &category.Deleted); err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to parse categories"})
		}

		categories = append(categories, category)

	}

	return c.Status(http.StatusOK).JSON(fiber.Map{"categories": categories})
}
