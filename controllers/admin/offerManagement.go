package admin

import (
	"horizon/config"
	"time"

	"github.com/gofiber/fiber/v2"
)

func AddOffer(c *fiber.Ctx) error {
	var offer struct {
		ProductID          int     `json:"product_id"`
		DiscountPercentage float64 `json:"discount_percentage"`
		StartDate          string  `json:"start_date"`
		EndDate            string  `json:"end_date"`
	}

	if err := c.BodyParser(&offer); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	startDate, err := time.Parse(time.RFC3339, offer.StartDate)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid start_date format"})
	}
	endDate, err := time.Parse(time.RFC3339, offer.EndDate)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid end_date format"})
	}

	if endDate.Before(startDate) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "end_date must be after start_date"})
	}

	var existingOfferCount int
	checkOfferQuery := `
		SELECT COUNT(*) 
		FROM offers 
		WHERE product_id = $1 
		AND NOW() BETWEEN start_date AND end_date
	`
	err = config.DB.QueryRow(checkOfferQuery, offer.ProductID).Scan(&existingOfferCount)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to check existing offer"})
	}

	if existingOfferCount > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Product already has an active offer"})
	}

	insertOfferQuery := `
		INSERT INTO offers (product_id, discount_percentage, start_date, end_date)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`
	var offerID int
	err = config.DB.QueryRow(insertOfferQuery, offer.ProductID, offer.DiscountPercentage, startDate, endDate).Scan(&offerID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to add offer"})
	}

	updateProductQuery := `
		UPDATE products
		SET discounted_price = price - (price * $1 / 100)
		WHERE id = $2
	`
	_, err = config.DB.Exec(updateProductQuery, offer.DiscountPercentage, offer.ProductID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update product price"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":  "Offer added successfully",
		"offer_id": offerID,
	})
}

func RemoveOffer(c *fiber.Ctx) error {

	productID := c.Params("product_id")
	if productID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Product ID is required",
		})
	}

	var offerID int
	err := config.DB.QueryRow("SELECT id FROM offers WHERE product_id = $1 AND end_date >= NOW()", productID).Scan(&offerID)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "No active offer found for the product",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to check offer",
		})
	}

	_, err = config.DB.Exec("DELETE FROM offers WHERE id = $1", offerID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to remove the offer",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Offer removed successfully",
	})
}

type Offer struct {
	ID                 int     `json:"id"`
	ProductID          int     `json:"product_id"`
	DiscountPercentage float64 `json:"discount_percentage"`
	StartDate          string  `json:"start_date"`
	EndDate            string  `json:"end_date"`
}

func ViewOffer(c *fiber.Ctx) error {
	var offers []struct {
		ID                 int     `json:"id"`
		ProductID          int     `json:"product_id"`
		ProductName        string  `json:"product_name"`
		CategoryName       string  `json:"category_name"`
		DiscountPercentage float64 `json:"discount_percentage"`
		StartDate          string  `json:"start_date"`
		EndDate            string  `json:"end_date"`
	}

	query := `
		SELECT o.id, o.product_id, p.name as product_name, c.name as category_name,
		       o.discount_percentage, o.start_date, o.end_date
		FROM offers o
		JOIN products p ON o.product_id = p.id
		JOIN categories c ON p.category_id = c.id
	`
	rows, err := config.DB.Query(query)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch offers",
		})
	}
	defer rows.Close()

	for rows.Next() {
		var offer struct {
			ID                 int     `json:"id"`
			ProductID          int     `json:"product_id"`
			ProductName        string  `json:"product_name"`
			CategoryName       string  `json:"category_name"`
			DiscountPercentage float64 `json:"discount_percentage"`
			StartDate          string  `json:"start_date"`
			EndDate            string  `json:"end_date"`
		}
		if err := rows.Scan(&offer.ID, &offer.ProductID, &offer.ProductName, &offer.CategoryName,
			&offer.DiscountPercentage, &offer.StartDate, &offer.EndDate); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to parse offer details",
			})
		}
		offers = append(offers, offer)
	}

	if len(offers) == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "No offers found",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Offers fetched successfully",
		"offers":  offers,
	})
}
