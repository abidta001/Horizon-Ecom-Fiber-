package admin

import (
	"fmt"
	"horizon/config"
	"horizon/models"
	"horizon/sql"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func AddProduct(c *fiber.Ctx) error {
	product := new(models.Product)
	if err := c.BodyParser(product); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	if product.Price < 0 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Price cannot be negative"})
	}

	if product.Stock < 0 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Stock cannot be negative"})
	}

	product.Name = strings.TrimSpace(product.Name)
	if len(product.Name) < 3 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Product name must be at least 3 characters long"})
	}

	product.Description = strings.TrimSpace(product.Description)
	if len(product.Description) < 5 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Product description must be at least 5 characters long"})
	}

	query := `INSERT INTO products (name, description, price, category_id, stock) VALUES ($1, $2, $3, $4, $5) RETURNING id`
	err := config.DB.QueryRow(query, product.Name, product.Description, product.Price, product.CategoryID, product.Stock).Scan(&product.ID)
	if err != nil {
		fmt.Println("er", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to add product"})
	}

	return c.Status(http.StatusCreated).JSON(fiber.Map{"message": "Product added successfully", "product": product})
}

func EditProduct(c *fiber.Ctx) error {
	product := new(models.Product)
	if err := c.BodyParser(product); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	if product.Price < 0 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Price cannot be negative"})
	}
	if product.Stock < 0 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Stock cannot be negative"})
	}

	if len(product.Name) < 3 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Product name must be at least 3 characters long"})
	}
	if len(product.Description) < 5 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Product description must be at least 5 characters long"})
	}

	var exists bool
	checkQuery := `SELECT EXISTS (SELECT 1 FROM products WHERE id=$1 AND deleted=false)`
	err := config.DB.QueryRow(checkQuery, product.ID).Scan(&exists)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to check product existence"})
	}
	if !exists {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Product not found or unavailable"})
	}

	updateQuery := `UPDATE products SET name=$1, description=$2, price=$3, category_id=$4, stock=$5 WHERE id=$6 AND deleted=false`
	_, err = config.DB.Exec(updateQuery, product.Name, product.Description, product.Price, product.CategoryID, product.Stock, product.ID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update product"})
	}

	return c.JSON(fiber.Map{"message": "Product updated successfully"})
}

func SoftDeleteProduct(c *fiber.Ctx) error {
	productID := c.Params("id")
	query := `UPDATE products SET deleted=true WHERE id=$1`
	_, err := config.DB.Exec(query, productID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete product"})
	}

	return c.JSON(fiber.Map{"message": "Product deleted successfully"})
}
func RecoverProduct(c *fiber.Ctx) error {
	productID := c.Params("id")
	query := `UPDATE products SET deleted=false WHERE id=$1`
	_, err := config.DB.Exec(query, productID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to recover product"})
	}

	return c.JSON(fiber.Map{"message": "Product retrieved successfully"})
}

func AdminViewProducts(c *fiber.Ctx) error {
	rows, err := config.DB.Query(sql.AdminViewProductsQuery)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch products",
		})
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var product models.Product
		if err := rows.Scan(&product.ID, &product.Name, &product.Description, &product.Price, &product.CategoryID, &product.Stock, &product.Deleted); err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to parse products",
			})
		}
		products = append(products, product)
	}

	return c.JSON(fiber.Map{
		"message":  "Products fetched successfully",
		"products": products,
	})
}
func UpdateProductStock(c *fiber.Ctx) error {
	var stockUpdate struct {
		ProductID int `json:"product_id"`
		Stock     int `json:"stock"`
	}

	if err := c.BodyParser(&stockUpdate); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	if stockUpdate.Stock < 0 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Stock cannot be negative"})
	}

	var exists bool
	checkQuery := `SELECT EXISTS (SELECT 1 FROM products WHERE id=$1 AND deleted=false)`
	err := config.DB.QueryRow(checkQuery, stockUpdate.ProductID).Scan(&exists)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to check product existence"})
	}
	if !exists {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Product not found or unavailable"})
	}

	updateQuery := `UPDATE products SET stock = $1 WHERE id = $2 AND deleted=false`
	_, err = config.DB.Exec(updateQuery, stockUpdate.Stock, stockUpdate.ProductID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update product stock"})
	}

	return c.JSON(fiber.Map{"message": "Stock updated successfully"})
}
