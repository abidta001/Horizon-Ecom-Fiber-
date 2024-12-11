package users

import (
	"context"
	"fmt"
	"horizon/config"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/plutov/paypal/v4"
)

func Checkout(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(int)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user session"})
	}

	addressID := c.Query("address_id")
	couponCode := c.Query("coupon_code")
	paymentMethod := c.Query("payment_method")

	if addressID == "" || paymentMethod == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing required parameters"})
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
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid or missing address"})
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

	cartQuery := `
		SELECT p.id, p.stock, c.quantity, p.price
		FROM cart c
		JOIN products p ON c.product_id = p.id
		WHERE c.user_id = $1
	`
	rows, err := tx.Query(cartQuery, userID)
	if err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch cart items"})
	}
	defer rows.Close()

	var orderTotal float64
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
		orderTotal += float64(item.Quantity) * item.Price
		cartItems = append(cartItems, item)
	}

	if len(cartItems) == 0 {
		tx.Rollback()
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cart is empty"})
	}

	var couponDiscountFloat float64
	if couponCode != "" && paymentMethod != "cod" {
		var discountPercentage, maxDiscountAmount, minOrderAmount float64
		var usedCount, usageLimit int
		var startDate, endDate time.Time
		couponQuery := `
			SELECT discount_percentage, max_discount_amount, min_order_amount, used_count, usage_limit, start_date, end_date
			FROM coupons
			WHERE code = $1 AND CURRENT_TIMESTAMP BETWEEN start_date AND end_date
		`

		err := config.DB.QueryRow(couponQuery, couponCode).Scan(&discountPercentage, &maxDiscountAmount, &minOrderAmount, &usedCount, &usageLimit, &startDate, &endDate)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid or expired coupon"})
		}

		if usedCount >= usageLimit {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Coupon usage limit reached"})
		}

		if orderTotal < minOrderAmount {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Order total does not meet the minimum amount required for this coupon"})
		}

		discount := (orderTotal * discountPercentage) / 100
		if discount > maxDiscountAmount {
			discount = maxDiscountAmount
		}
		couponDiscountFloat = discount
		orderTotal -= couponDiscountFloat
	}

	var offerDiscountFloat float64

	for _, item := range cartItems {
		var discountPercentage float64

		offerQuery := `
    SELECT discount_percentage
    FROM offers
    WHERE product_id = $1 AND CURRENT_TIMESTAMP BETWEEN start_date AND end_date
	`
		err := config.DB.QueryRow(offerQuery, item.ProductID).Scan(&discountPercentage)
		if err == nil {
			offerDiscountFloat += (float64(item.Quantity) * item.Price * discountPercentage) / 100
		}
		if err != nil && err.Error() != "no rows in result set" {
			fmt.Println(err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch offer data"})
		}
	}

	orderTotal -= offerDiscountFloat

	if orderTotal < 0 {
		orderTotal = 0
	}

	var paymentStatus, status string
	if paymentMethod == "cod" {
		if orderTotal > 1000 {
			tx.Rollback()
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "COD not allowed for orders above 1000"})
		}
		paymentStatus = "Pending"
		status = "Pending COD Verification"
	} else {
		paymentStatus = "Processing"
		status = "Pending"
	}

	uniqueOrderID := fmt.Sprintf("ORD-%d", time.Now().UnixNano())
	var orderID int
	createOrderQuery := `
	INSERT INTO orders 
	(order_id, user_id, total_amount, coupon_discount, offer_discount, payment_method, payment_status, status, address_line, city, zip_code) 
	VALUES 
	($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) 
	RETURNING id
`
	err = tx.QueryRow(createOrderQuery, uniqueOrderID, userID, orderTotal, couponDiscountFloat, offerDiscountFloat, paymentMethod, paymentStatus, status, address.AddressLine, address.City, address.ZipCode).Scan(&orderID)
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

	if couponCode != "" {
		updateCouponQuery := `
		UPDATE coupons
		SET used_count = used_count + 1
		WHERE code = $1
	`
		_, err = tx.Exec(updateCouponQuery, couponCode)
		if err != nil {
			tx.Rollback()
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update coupon usage"})
		}
	}

	if err := tx.Commit(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to finalize transaction"})
	}

	if paymentMethod == "paypal" {
		paypalClient := config.GetPayPalClient()
		fixedExchangeRate := 0.012
		orderTotalInUsd := orderTotal * fixedExchangeRate
		order, err := paypalClient.CreateOrder(
			context.Background(),
			"CAPTURE",
			[]paypal.PurchaseUnitRequest{
				{
					ReferenceID: fmt.Sprintf("%d", orderID),
					Amount: &paypal.PurchaseUnitAmount{
						Currency: "USD",
						Value:    fmt.Sprintf("%.2f", orderTotalInUsd),
					},
				},
			},
			nil,
			&paypal.ApplicationContext{
				ReturnURL: "https://horizonweb.me/paypal/success",
				CancelURL: "https://horizonweb.me/paypal/cancel",
			},
		)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create PayPal order"})
		}
		if len(order.Links) > 1 {
			return c.JSON(fiber.Map{"url": order.Links[1].Href})
		}
	}

	return c.JSON(fiber.Map{"message": "Order placed successfully", "order_id": uniqueOrderID, "total_amount": orderTotal})
}
