package users

import (
	"fmt"
	"horizon/config"
	responsemodels "horizon/models/responsemodels"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
)

func ViewOrder(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(int)
	if !ok {
		log.Println("Invalid or missing userID in request context")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	query := `
		SELECT 
			o.id AS order_id,
			o.order_id AS reference_id,
			p.name AS product_name,
			p.price AS product_price,
			cat.name AS category_name,
			oi.quantity,
			oi.subtotal,
			o.coupon_discount,
			o.offer_discount,
			o.total_amount AS amount_paid,
			o.payment_status,
			o.status AS order_status
		FROM orders o
		JOIN order_items oi ON o.id = oi.order_id
		JOIN products p ON oi.product_id = p.id
		JOIN categories cat ON p.category_id = cat.id
		WHERE o.user_id = $1
	`

	rows, err := config.DB.Query(query, userID)
	if err != nil {
		log.Printf("Database query error: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve order details"})
	}
	defer rows.Close()

	var orderDetails []responsemodels.OrderDetail

	for rows.Next() {
		var detail responsemodels.OrderDetail
		err := rows.Scan(
			&detail.OrderID,
			&detail.ReferenceID,
			&detail.ProductName,
			&detail.ProductPrice,
			&detail.CategoryName,
			&detail.Quantity,
			&detail.Subtotal,
			&detail.CouponDiscount,
			&detail.OfferDiscount,
			&detail.AmountPaid,
			&detail.PaymentStatus,
			&detail.OrderStatus,
		)
		if err != nil {
			log.Printf("Row scan error: %v\n", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to parse order details"})
		}
		orderDetails = append(orderDetails, detail)
	}

	if len(orderDetails) == 0 {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "No orders found for the user",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":       "Orders retrieved successfully",
		"order_details": orderDetails,
	})
}

func CancelOrder(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(int)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user session"})
	}

	orderID := c.Params("order_id")
	if orderID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing order ID"})
	}

	tx, err := config.DB.Begin()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to start transaction"})
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
		}
	}()

	var orderAmount float64
	var status, paymentStatus string
	checkOrderQuery := `
		SELECT total_amount, status, payment_status
		FROM orders
		WHERE id = $1 AND user_id = $2
	`
	err = tx.QueryRow(checkOrderQuery, orderID, userID).Scan(&orderAmount, &status, &paymentStatus)
	if err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid order"})
	}

	// Check if the order is already canceled
	if status == "Cancelled" {
		tx.Rollback()
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Order has already been cancelled"})
	}

	// Check if the order is eligible for cancellation
	if status == "Delivered" || status == "Returned" {
		tx.Rollback()
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Order cannot be canceled as it is either Delivered or Returned"})
	}

	// Refund amount to wallet only if payment status is Paid or Completed
	if paymentStatus == "Paid" || paymentStatus == "Completed" {
		updateWalletQuery := `
			UPDATE users
			SET wallet_balance = wallet_balance + $1
			WHERE id = $2
		`
		_, err = tx.Exec(updateWalletQuery, orderAmount, userID)
		if err != nil {
			tx.Rollback()
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update wallet"})
		}

		insertTransactionQuery := `
			INSERT INTO wallet_transactions (user_id, order_id, amount, transaction_type)
			VALUES ($1, $2, $3, 'refund')
		`
		_, err = tx.Exec(insertTransactionQuery, userID, orderID, orderAmount)
		if err != nil {
			tx.Rollback()
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to log wallet transaction"})
		}
	}

	// Update the order status to Cancelled
	updateOrderQuery := `
		UPDATE orders
		SET status = 'Cancelled'
		WHERE id = $1
	`
	_, err = tx.Exec(updateOrderQuery, orderID)
	if err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update order status"})
	}

	if err := tx.Commit(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to commit transaction"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Order cancelled successfully",
	})
}
func UseWalletForPurchase(c *fiber.Ctx) error {

	userID, ok := c.Locals("userID").(int)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user session"})
	}

	addressID := c.Query("address_id")
	if addressID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing address ID"})
	}

	var address struct {
		AddressLine string
		City        string
		ZipCode     string
	}
	addressQuery := `
		SELECT address_line, city, zip_code
		FROM addresses
		WHERE id = $1 AND user_id = $2
	`
	err := config.DB.QueryRow(addressQuery, addressID, userID).Scan(&address.AddressLine, &address.City, &address.ZipCode)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid or unauthorized address"})
	}

	var walletBalance float64
	err = config.DB.QueryRow(`
		SELECT wallet_balance
		FROM users
		WHERE id = $1
	`, userID).Scan(&walletBalance)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch wallet balance"})
	}

	var cartTotal float64
	cartQuery := `
		SELECT SUM(c.quantity * p.price)
		FROM cart c
		JOIN products p ON c.product_id = p.id
		WHERE c.user_id = $1
	`
	err = config.DB.QueryRow(cartQuery, userID).Scan(&cartTotal)
	if err != nil || cartTotal <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Failed to calculate cart total"})
	}

	if walletBalance < cartTotal {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Insufficient wallet balance"})
	}

	tx, err := config.DB.Begin()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to start transaction"})
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
		}
	}()

	cartItemsQuery := `
		SELECT p.id, p.stock, c.quantity, p.price
		FROM cart c
		JOIN products p ON c.product_id = p.id
		WHERE c.user_id = $1
	`
	rows, err := tx.Query(cartItemsQuery, userID)
	if err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch cart items"})
	}
	defer rows.Close()

	var cartItems []struct {
		ProductID int
		Stock     int
		Quantity  int
		Price     float64
	}
	for rows.Next() {
		var item struct {
			ProductID int
			Stock     int
			Quantity  int
			Price     float64
		}
		if err := rows.Scan(&item.ProductID, &item.Stock, &item.Quantity, &item.Price); err != nil {
			tx.Rollback()
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to parse cart items"})
		}
		if item.Quantity > item.Stock {
			tx.Rollback()
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error":   "Insufficient stock for product",
				"product": item.ProductID,
			})
		}
		cartItems = append(cartItems, item)
	}

	if len(cartItems) == 0 {
		tx.Rollback()
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cart is empty"})
	}

	newBalance := walletBalance - cartTotal
	_, err = tx.Exec(`
		UPDATE users
		SET wallet_balance = $1
		WHERE id = $2
	`, newBalance, userID)
	if err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update wallet balance"})
	}

	uniqueOrderID := fmt.Sprintf("ORD-%d", time.Now().UnixNano())
	var orderID int
	createOrderQuery := `
		INSERT INTO orders 
		(order_id, user_id, total_amount, payment_method, payment_status, status, address_line, city, zip_code) 
		VALUES 
		($1, $2, $3, $4, $5, $6, $7, $8, $9) 
		RETURNING id
	`
	err = tx.QueryRow(createOrderQuery, uniqueOrderID, userID, cartTotal, "wallet", "Paid", "Confirmed", address.AddressLine, address.City, address.ZipCode).Scan(&orderID)
	if err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create order"})
	}

	for _, item := range cartItems {
		subtotal := item.Price * float64(item.Quantity)
		_, err := tx.Exec(`
			INSERT INTO order_items (order_id, product_id, quantity, price, subtotal)
			VALUES ($1, $2, $3, $4, $5)`,
			orderID, item.ProductID, item.Quantity, item.Price, subtotal,
		)
		if err != nil {
			tx.Rollback()
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to add order items"})
		}
	}

	_, err = tx.Exec(`DELETE FROM cart WHERE user_id = $1`, userID)
	if err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to clear cart"})
	}

	if err := tx.Commit(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to finalize transaction"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":        "Order placed successfully using wallet",
		"order_id":       uniqueOrderID,
		"wallet_balance": newBalance,
	})
}
