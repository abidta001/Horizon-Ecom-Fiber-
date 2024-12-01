package users

import (
	"horizon/config"
	"horizon/models"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

func ApplyCoupon(c *fiber.Ctx) error {
	couponCode := c.Query("code")
	orderAmount := c.Query("order_amount")
	if couponCode == "" || orderAmount == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Coupon code and order amount are required"})
	}

	var coupon models.Coupon
	query := `
        SELECT id, discount_percentage, max_discount_amount, min_order_amount, usage_limit, used_count, end_date 
        FROM coupons 
        WHERE code = $1 AND end_date >= NOW()
    `
	err := config.DB.QueryRow(query, couponCode).Scan(&coupon.ID, &coupon.DiscountPercentage, &coupon.MaxDiscountAmount, &coupon.MinOrderAmount, &coupon.UsageLimit, &coupon.UsedCount, &coupon.EndDate)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Invalid or expired coupon"})
	}

	orderAmountFloat, err := strconv.ParseFloat(orderAmount, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid order amount"})
	}

	if orderAmountFloat < coupon.MinOrderAmount {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Order amount does not meet the minimum requirement for this coupon"})
	}

	if coupon.UsedCount >= coupon.UsageLimit {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Coupon usage limit reached"})
	}

	discount := (coupon.DiscountPercentage / 100) * orderAmountFloat
	if discount > coupon.MaxDiscountAmount {
		discount = coupon.MaxDiscountAmount
	}

	return c.JSON(fiber.Map{
		"message":    "Coupon applied successfully",
		"discount":   discount,
		"finalPrice": orderAmountFloat - discount,
	})
}
func ViewCoupons(c *fiber.Ctx) error {
	query := `
        SELECT id, code, discount_percentage, max_discount_amount, min_order_amount, start_date, end_date, usage_limit, used_count, created_at, updated_at
        FROM coupons
        WHERE end_date >= NOW()
    `
	rows, err := config.DB.Query(query)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch coupons"})
	}
	defer rows.Close()

	var coupons []models.Coupon
	for rows.Next() {
		var coupon models.Coupon
		if err := rows.Scan(&coupon.ID, &coupon.Code, &coupon.DiscountPercentage, &coupon.MaxDiscountAmount, &coupon.MinOrderAmount, &coupon.StartDate, &coupon.EndDate, &coupon.UsageLimit, &coupon.UsedCount, &coupon.CreatedAt, &coupon.UpdatedAt); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to parse coupons"})
		}
		coupons = append(coupons, coupon)
	}

	return c.JSON(fiber.Map{
		"message": "Coupons fetched successfully",
		"coupons": coupons,
	})
}
