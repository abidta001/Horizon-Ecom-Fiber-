package users

import (
	"horizon/config"
	"horizon/models"
	responsemodels "horizon/models/responsemodels"

	"github.com/gofiber/fiber/v2"
)

func AddToCart(c *fiber.Ctx) error {
	userID := c.Locals("userID").(int)
	cartItem := new(models.CartItem)

	if err := c.BodyParser(cartItem); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	const maxQtyPerPerson = 10

	var availableStock int
	query := `SELECT stock FROM products WHERE id=$1 AND deleted=false`
	err := config.DB.QueryRow(query, cartItem.ProductID).Scan(&availableStock)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Product not found or unavailable"})
	}

	var currentCartQty int
	cartQuery := `SELECT quantity FROM cart WHERE user_id=$1 AND product_id=$2`
	_ = config.DB.QueryRow(cartQuery, userID, cartItem.ProductID).Scan(&currentCartQty)

	if cartItem.Quantity+currentCartQty > maxQtyPerPerson {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error":                       "Requested quantity exceeds the maximum allowed per person",
			"Maximum Quantity per person": maxQtyPerPerson,
			"In your Cart":                currentCartQty,
		})
	}

	if cartItem.Quantity+currentCartQty > availableStock {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error":   "Requested quantity exceeds available stock",
			"stock":   availableStock,
			"current": currentCartQty,
		})
	}

	query = `
        INSERT INTO cart (user_id, product_id, quantity) 
        VALUES ($1, $2, $3)
        ON CONFLICT (user_id, product_id) 
        DO UPDATE SET quantity = cart.quantity + $3, updated_at = NOW()`
	_, err = config.DB.Exec(query, userID, cartItem.ProductID, cartItem.Quantity)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to add to cart"})
	}

	return c.JSON(fiber.Map{"message": "Product added to cart successfully"})
}

func RemoveFromCart(c *fiber.Ctx) error {
	userID := c.Locals("userID").(int)
	productID := c.Params("product_id")

	query := `DELETE FROM cart WHERE user_id=$1 AND product_id=$2`
	result, err := config.DB.Exec(query, userID, productID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to remove from cart"})
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Product not found in cart"})
	}

	return c.JSON(fiber.Map{"message": "Product removed from cart successfully"})
}
func ViewCart(c *fiber.Ctx) error {
	userID := c.Locals("userID").(int)
	var cart []responsemodels.ViewCartItem

	query := `
        SELECT p.id, p.name, p.price, c.quantity, 
               (p.price * c.quantity) AS subtotal
        FROM cart c
        JOIN products p ON c.product_id = p.id
        WHERE c.user_id=$1`
	rows, err := config.DB.Query(query, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch cart items"})
	}
	defer rows.Close()

	for rows.Next() {
		var item responsemodels.ViewCartItem
		if err := rows.Scan(&item.ID, &item.Name, &item.Price, &item.Quantity, &item.Subtotal); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to parse cart item"})
		}
		cart = append(cart, item)
	}

	return c.JSON(fiber.Map{"message": "Cart fetched successfully", "cart": cart})
}
func ClearCart(c *fiber.Ctx) error {
	userID := c.Locals("userID").(int)

	query := `DELETE FROM cart WHERE user_id=$1`
	_, err := config.DB.Exec(query, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to clear cart"})
	}

	return c.JSON(fiber.Map{"message": "Cart cleared successfully"})
}
