package admin

import (
	"horizon/config"
	"horizon/models"
	"time"

	"github.com/gofiber/fiber/v2"
)

func CreateCoupon(c *fiber.Ctx) error {
	var coupon models.Coupon
	if err := c.BodyParser(&coupon); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	if coupon.Code == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Coupon code is required"})
	}

	if coupon.DiscountPercentage <= 0 || coupon.DiscountPercentage > 100 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Discount percentage must be between 1 and 100"})
	}

	if coupon.MaxDiscountAmount < 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Max discount amount must be non-negative"})
	}

	if coupon.MinOrderAmount < 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Min order amount must be non-negative"})
	}

	startDate, err := time.Parse(time.RFC3339, coupon.StartDate)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid start date format"})
	}
	endDate, err := time.Parse(time.RFC3339, coupon.EndDate)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid end date format"})
	}
	if endDate.Before(startDate) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "End date must be after start date"})
	}

	if coupon.UsageLimit < 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Usage limit must be non-negative"})
	}

	var existingCouponCount int
	checkCouponQuery := `SELECT COUNT(*) FROM coupons WHERE code = $1`
	err = config.DB.QueryRow(checkCouponQuery, coupon.Code).Scan(&existingCouponCount)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to check coupon code uniqueness"})
	}
	if existingCouponCount > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Coupon code already exists"})
	}

	query := `
        INSERT INTO coupons (code, discount_percentage, max_discount_amount, min_order_amount, start_date, end_date, usage_limit) 
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        RETURNING id
    `
	var couponID int
	err = config.DB.QueryRow(query, coupon.Code, coupon.DiscountPercentage, coupon.MaxDiscountAmount, coupon.MinOrderAmount, startDate, endDate, coupon.UsageLimit).Scan(&couponID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create coupon"})
	}

	return c.JSON(fiber.Map{
		"message":  "Coupon created successfully",
		"couponID": couponID,
	})
}
func ViewCouponsAdmin(c *fiber.Ctx) error {
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

func RemoveCoupon(c *fiber.Ctx) error {
	couponID := c.Params("id")
	if couponID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Coupon ID is required",
		})
	}

	query := `DELETE FROM coupons WHERE id = $1`

	result, err := config.DB.Exec(query, couponID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to remove coupon",
		})
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to determine the result of deletion",
		})
	}

	if rowsAffected == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Coupon not found or already deleted",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Coupon removed successfully",
	})
}
