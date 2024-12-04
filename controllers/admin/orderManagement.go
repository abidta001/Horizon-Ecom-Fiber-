package admin

import (
	"horizon/config"
	"horizon/models"
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

func AdminListOrder(c *fiber.Ctx) error {
	query := `
		SELECT 
			o.id AS order_id,
			o.order_id AS reference_id,
			o.user_id,
			u.name AS user_name,
			u.email AS user_email,   -- Added user email
			p.id AS product_id,
			p.name AS product_name,
			cat.name AS category_name,
			oi.quantity,
			oi.subtotal,
			o.payment_status,
			o.payment_method,
			o.status AS order_status,
			o.order_date::DATE AS order_date,
			o.address_line,
			o.city,
			o.zip_code,
			o.total_amount,
			o.coupon_discount,      -- Added coupon discount
			o.offer_discount,       -- Added offer discount
			(oi.subtotal - o.coupon_discount - o.offer_discount) AS final_amount -- Calculated final amount
		FROM orders o
		JOIN order_items oi ON o.id = oi.order_id
		JOIN products p ON oi.product_id = p.id
		JOIN categories cat ON p.category_id = cat.id
		JOIN users u ON o.user_id = u.id
	`

	rows, err := config.DB.Query(query)
	if err != nil {
		log.Printf("Database query error: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve order details"})
	}
	defer rows.Close()

	var orderDetails []models.OrderDetail

	for rows.Next() {
		var detail models.OrderDetail
		err := rows.Scan(
			&detail.OrderID,
			&detail.ReferenceID,
			&detail.UserID,
			&detail.UserName,
			&detail.UserEmail,
			&detail.ProductID,
			&detail.ProductName,
			&detail.CategoryName,
			&detail.Quantity,
			&detail.Subtotal,
			&detail.PaymentStatus,
			&detail.PaymentMethod,
			&detail.OrderStatus,
			&detail.OrderDate,
			&detail.AddressLine,
			&detail.City,
			&detail.ZipCode,
			&detail.TotalAmount,
			&detail.CouponDiscount,
			&detail.OfferDiscount,
			&detail.FinalAmount,
		)
		if err != nil {
			log.Printf("Row scan error: %v\n", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to parse order details"})
		}
		orderDetails = append(orderDetails, detail)
	}

	if len(orderDetails) == 0 {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "No orders found",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":       "Orders retrieved successfully",
		"order_details": orderDetails,
	})
}

func AdminChangeOrderStatus(c *fiber.Ctx) error {
	orderIDParam := c.Params("order_id")
	orderID, err := strconv.Atoi(orderIDParam)
	if err != nil {
		log.Printf("Invalid order ID: %v\n", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid order ID"})
	}

	type StatusUpdateRequest struct {
		Status string `json:"status"`
	}
	var statusUpdate StatusUpdateRequest
	if err := c.BodyParser(&statusUpdate); err != nil {
		log.Printf("Error parsing request body: %v\n", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	validStatuses := map[string]bool{
		"Pending":   true,
		"Shipped":   true,
		"Delivered": true,
		"Cancelled": true,
		"Returned":  true,
	}
	if !validStatuses[statusUpdate.Status] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid status"})
	}

	var currentStatus, paymentStatus string
	var userID int
	var totalAmount float64

	query := `
		SELECT status, payment_status, user_id, total_amount
		FROM orders
		WHERE id = $1
	`
	err = config.DB.QueryRow(query, orderID).Scan(&currentStatus, &paymentStatus, &userID, &totalAmount)
	if err != nil {
		log.Printf("Failed to fetch current order details: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch order details"})
	}

	if statusUpdate.Status == "Cancelled" && currentStatus == "Delivered" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Order cannot be canceled after being delivered",
		})
	}

	if (currentStatus == "Delivered" || currentStatus == "Returned" || currentStatus == "Cancelled") &&
		statusUpdate.Status != "Returned" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Order status cannot be changed from Delivered, Returned, or Canceled except to Returned",
		})
	}

	tx, err := config.DB.Begin()
	if err != nil {
		log.Printf("Failed to start transaction: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to process request"})
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
		}
	}()

	if (statusUpdate.Status == "Cancelled" || statusUpdate.Status == "Returned") &&
		(paymentStatus == "Paid" || paymentStatus == "Completed") {
		updateWalletQuery := `
			UPDATE users
			SET wallet_balance = wallet_balance + $1
			WHERE id = $2
		`
		_, err = tx.Exec(updateWalletQuery, totalAmount, userID)
		if err != nil {
			tx.Rollback()
			log.Printf("Failed to update wallet balance: %v\n", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update wallet balance"})
		}

		insertWalletTransactionQuery := `
		INSERT INTO wallet_transactions (user_id, order_id, amount, transaction_type, created_at)
		VALUES ($1, $2, $3, $4, NOW())
		`
		_, err = tx.Exec(insertWalletTransactionQuery, userID, orderID, totalAmount, "Refund")
		if err != nil {
			tx.Rollback()
			log.Printf("Failed to insert wallet transaction: %v\n", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to record wallet transaction"})
		}
	}

	updateOrderQuery := `
		UPDATE orders
		SET status = $1
		WHERE id = $2
	`
	result, err := tx.Exec(updateOrderQuery, statusUpdate.Status, orderID)
	if err != nil {
		tx.Rollback()
		log.Printf("Failed to update order status: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update order status"})
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		tx.Rollback()
		log.Printf("Error checking rows affected: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update order status"})
	}
	if rowsAffected == 0 {
		tx.Rollback()
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Order not found"})
	}

	if err := tx.Commit(); err != nil {
		log.Printf("Failed to commit transaction: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to finalize transaction"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Order status updated successfully",
	})
}
