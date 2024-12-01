package users

import (
	"fmt"
	"horizon/config"
	"horizon/models"
	responsemodels "horizon/models/responsemodels"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

type ProductView struct {
	ID                 int      `json:"id"`
	Name               string   `json:"name"`
	Description        string   `json:"description"`
	Price              float64  `json:"price"`
	FinalPrice         float64  `json:"final_price"`
	DiscountPercentage *float64 `json:"discount_percentage,omitempty"`
	CategoryName       string   `json:"category_name"`
	Status             string   `json:"status"`
}

func ViewProducts(c *fiber.Ctx) error {
	query := `
		SELECT 
			p.id, 
			p.name, 
			p.description, 
			p.price, 
			COALESCE(
				CASE 
					WHEN o.start_date <= NOW() AND o.end_date >= NOW() THEN 
						p.price - (p.price * o.discount_percentage / 100) 
				END, p.price
			) AS final_price,
			o.discount_percentage,
			c.name AS category_name, 
			CASE 
				WHEN p.stock > 0 THEN 'Available'
				ELSE 'Out of Stock' 
			END AS status
		FROM products p
		LEFT JOIN offers o ON p.id = o.product_id AND o.start_date <= NOW() AND o.end_date >= NOW()
		JOIN categories c ON p.category_id = c.id
		WHERE p.deleted = false`

	rows, err := config.DB.Query(query)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch products",
		})
	}
	defer rows.Close()

	var products []ProductView
	for rows.Next() {
		var product ProductView
		if err := rows.Scan(&product.ID, &product.Name, &product.Description, &product.Price, &product.FinalPrice, &product.DiscountPercentage, &product.CategoryName, &product.Status); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
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

func ViewCategories(c *fiber.Ctx) error {
	query := `SELECT  name, description FROM categories WHERE deleted=false`
	rows, err := config.DB.Query(query)
	if err != nil {
		fmt.Println(err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch categories"})
	}
	defer rows.Close()

	var categories []responsemodels.ViewCategory
	for rows.Next() {
		var category responsemodels.ViewCategory
		if err := rows.Scan(&category.Name, &category.Description); err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to parse categories"})
		}

		categories = append(categories, category)

	}

	return c.Status(http.StatusOK).JSON(fiber.Map{"categories": categories})
}

func SearchProducts(c *fiber.Ctx) error {
	// Get sorting criteria from query parameters
	sortBy := c.Query("sort_by", "price_asc") // Default sorting: price low to high
	var orderBy string

	// Determine sorting order based on query parameter
	switch sortBy {
	case "price_desc":
		orderBy = "ORDER BY price DESC"
	case "name_asc":
		orderBy = "ORDER BY name ASC"
	case "name_desc":
		orderBy = "ORDER BY name DESC"
	default:
		orderBy = "ORDER BY price ASC" // Default sorting: price low to high
	}

	// Fetch products based on the sorting criteria
	query := `SELECT id, name, description, price, category_id, stock, deleted FROM products WHERE deleted = false ` + orderBy

	rows, err := config.DB.Query(query)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch products"})
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var product models.Product
		if err := rows.Scan(&product.ID, &product.Name, &product.Description, &product.Price, &product.CategoryID, &product.Stock, &product.Deleted); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to scan product"})
		}
		products = append(products, product)
	}

	return c.JSON(fiber.Map{
		"message":  "Products fetched successfully",
		"products": products,
	})
}
