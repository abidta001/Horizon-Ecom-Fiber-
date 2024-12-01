package users

import (
	"horizon/config"
	"horizon/models"
	responsemodels "horizon/models/responsemodels"

	"github.com/gofiber/fiber/v2"
)

func AddToWishlist(c *fiber.Ctx) error {
	wishlist := new(models.Wishlist)
	userID := c.Locals("userID").(int)

	if err := c.BodyParser(wishlist); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	query := `INSERT INTO wishlists (user_id, product_id) VALUES ($1, $2) ON CONFLICT (user_id, product_id) DO NOTHING`
	_, err := config.DB.Exec(query, userID, wishlist.ProductID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to add product to wishlist"})
	}

	return c.JSON(fiber.Map{"message": "Product added to wishlist successfully"})
}

func RemoveFromWishlist(c *fiber.Ctx) error {
	productID := c.Params("product_id")
	userID := c.Locals("userID").(int)

	query := `DELETE FROM wishlists WHERE user_id=$1 AND product_id=$2`
	result, err := config.DB.Exec(query, userID, productID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to remove product from wishlist"})
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Product not found in wishlist"})
	}

	return c.JSON(fiber.Map{"message": "Product removed from wishlist successfully"})
}
func ClearWishlist(c *fiber.Ctx) error {
	userID := c.Locals("userID").(int)

	query := `DELETE FROM wishlists WHERE user_id=$1`
	_, err := config.DB.Exec(query, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to clear wishlist"})
	}

	return c.JSON(fiber.Map{"message": "Wishlist cleared successfully"})
}
func ViewWishlist(c *fiber.Ctx) error {
	userID := c.Locals("userID").(int)
	var wishlist []responsemodels.ViewProducts

	query := `SELECT p.id, p.name, p.description, p.price, c.name AS category_name, 
              CASE WHEN p.stock > 0 THEN 'Available' ELSE 'Out of Stock' END AS status
              FROM wishlists w
              JOIN products p ON w.product_id = p.id
              JOIN categories c ON p.category_id = c.id
              WHERE w.user_id=$1`
	rows, err := config.DB.Query(query, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch wishlist"})
	}
	defer rows.Close()

	for rows.Next() {
		var product responsemodels.ViewProducts
		if err := rows.Scan(&product.ID, &product.Name, &product.Description, &product.Price, &product.CategoryName, &product.Status); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to scan product"})
		}
		wishlist = append(wishlist, product)
	}

	return c.JSON(fiber.Map{"message": "Wishlist fetched successfully", "wishlist": wishlist})
}
