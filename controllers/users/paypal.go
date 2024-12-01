package users

import (
	"context"
	"horizon/config"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/plutov/paypal/v4"
)

var paypalClient *paypal.Client

func SetPayPalClient(client *paypal.Client) {
	paypalClient = client
}

func ExecutePayPalPayment(c *fiber.Ctx) error {
	orderID := c.Query("order_id")

	_, err := paypalClient.CaptureOrder(context.Background(), orderID, paypal.CaptureOrderRequest{})
	if err != nil {
		log.Printf("Failed to capture PayPal order: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Payment capture failed"})
	}

	updateOrderQuery := `UPDATE orders SET payment_status = 'Completed' WHERE id = $1`
	_, err = config.DB.Exec(updateOrderQuery, orderID)
	if err != nil {
		log.Printf("Database update error: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update order status in database"})
	}

	return c.JSON(fiber.Map{"message": "Payment successful"})
}

func PayPalSuccess(c *fiber.Ctx) error {
	paymentID := c.Query("token")
	payerID := c.Query("PayerID")

	if paymentID == "" || payerID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing payment information"})
	}

	client, err := paypal.NewClient(os.Getenv("PAYPAL_CLIENT"), os.Getenv("PAYPAL_SECRET"), paypal.APIBaseSandBox)
	if err != nil {
		log.Printf("PayPal client creation error: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create PayPal client"})
	}

	captureResp, err := client.CaptureOrder(context.Background(), paymentID, paypal.CaptureOrderRequest{})
	if err != nil {
		log.Printf("PayPal CaptureOrder error: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to capture PayPal payment"})
	}

	log.Printf("PayPal Capture Response Status: %s\n", captureResp.Status)
	if captureResp.Status != "COMPLETED" {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Payment not completed"})
	}

	var orderID string
	for _, purchaseUnit := range captureResp.PurchaseUnits {
		if purchaseUnit.ReferenceID != "" {
			orderID = purchaseUnit.ReferenceID
			break
		}
	}

	if orderID == "" {
		log.Println("Reference ID missing in PayPal response")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Order reference missing from PayPal response"})
	}

	updateOrderQuery := `UPDATE orders SET payment_status = 'Completed' WHERE id = $1`
	res, err := config.DB.Exec(updateOrderQuery, orderID)
	if err != nil {
		log.Printf("Failed to update payment status for order ID %s: %v\n", orderID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update order status"})
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		log.Printf("No rows updated for order ID %s. Payment status might already be 'Completed'.\n", orderID)
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "Order status update failed"})
	}

	deductStockQuery := `
		UPDATE products
		SET stock = stock - oi.quantity
		FROM order_items oi
		WHERE products.id = oi.product_id AND oi.order_id = $1
	`
	_, err = config.DB.Exec(deductStockQuery, orderID)
	if err != nil {
		log.Printf("Failed to deduct stock for order ID %s: %v\n", orderID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to deduct stock"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":  "Payment successful and order status updated",
		"order_id": orderID,
	})
}
func PayPalCancel(c *fiber.Ctx) error {

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Payment was canceled by the user.",
	})
}
