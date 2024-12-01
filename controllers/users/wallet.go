package users

import (
	"horizon/config"
	"time"

	"github.com/gofiber/fiber/v2"
)

func ViewWalletBalance(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(int)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user session"})
	}

	var walletBalance float64
	query := `SELECT wallet_balance FROM users WHERE id = $1`
	err := config.DB.QueryRow(query, userID).Scan(&walletBalance)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch wallet balance"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"wallet_balance": walletBalance,
	})
}
func ViewWalletTransactions(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(int)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user session"})
	}

	transactionsQuery := `
		SELECT id, order_id, amount, transaction_type, created_at
		FROM wallet_transactions
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := config.DB.Query(transactionsQuery, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch transactions"})
	}
	defer rows.Close()

	var transactions []struct {
		ID              int       `json:"id"`
		OrderID         int       `json:"order_id"`
		Amount          float64   `json:"amount"`
		TransactionType string    `json:"transaction_type"`
		CreatedAt       time.Time `json:"created_at"`
	}

	for rows.Next() {
		var transaction struct {
			ID              int       `json:"id"`
			OrderID         int       `json:"order_id"`
			Amount          float64   `json:"amount"`
			TransactionType string    `json:"transaction_type"`
			CreatedAt       time.Time `json:"created_at"`
		}

		err := rows.Scan(
			&transaction.ID,
			&transaction.OrderID,
			&transaction.Amount,
			&transaction.TransactionType,
			&transaction.CreatedAt,
		)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to parse transactions"})
		}

		transactions = append(transactions, transaction)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"transactions": transactions,
	})
}
