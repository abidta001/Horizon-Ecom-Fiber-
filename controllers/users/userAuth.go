package users

import (
	"fmt"
	"horizon/config"
	middleware "horizon/middlewares"
	"horizon/models"
	"horizon/utils"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func Signup(c *fiber.Ctx) error {
	user := new(models.User)

	if err := c.BodyParser(user); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	user.Name = strings.TrimSpace(user.Name)
	if len(user.Name) < 3 || len(user.Name) == 0 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Name must be at least 3 characters and cannot be empty or spaces"})
	}

	if !utils.IsValidEmail(user.Email) {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid email format"})
	}

	if len(user.Phone) != 10 || !utils.IsNumeric(user.Phone) {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid phone number"})
	}

	if len(user.Password) < 8 || !utils.IsStrongPassword(user.Password) {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Password must be at least 8 characters long and include letters, numbers, and symbols"})
	}

	var exists bool
	checkQuery := `SELECT EXISTS (SELECT 1 FROM users WHERE email=$1 OR phone=$2)`
	if err := config.DB.QueryRow(checkQuery, user.Email, user.Phone).Scan(&exists); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to check existing user"})
	}

	if exists {
		return c.Status(http.StatusConflict).JSON(fiber.Map{"error": "Email or phone number already exists"})
	}

	hashedPassword, err := utils.HashPassword(user.Password)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to process password"})
	}

	query := `INSERT INTO users (name, email, phone, password, verified) VALUES ($1, $2, $3, $4, $5) RETURNING id`
	var userID int
	if err := config.DB.QueryRow(query, user.Name, user.Email, user.Phone, hashedPassword, false).Scan(&userID); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create user"})
	}

	otp := utils.GenerateOTP()
	go utils.SendEmail(user.Email, "Horizon Ecommerce ", "Get Your Account verified by using OTP: "+otp)

	utils.StoreOTP(user.Email, otp)
	fmt.Println(otp)
	return c.Status(http.StatusCreated).JSON(fiber.Map{"message": "OTP has been sent to your email!"})
}

func Login(c *fiber.Ctx) error {
	login := new(models.User)

	if err := c.BodyParser(login); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	query := `SELECT id, password, verified, blocked FROM users WHERE email=$1`
	var user models.User
	var verified, blocked bool

	if err := config.DB.QueryRow(query, login.Email).Scan(&user.ID, &user.Password, &verified, &blocked); err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid email or password"})
	}

	if !verified {
		return c.Status(http.StatusForbidden).JSON(fiber.Map{"error": "Account not verified. Please verify with OTP."})
	}

	if blocked {
		return c.Status(http.StatusForbidden).JSON(fiber.Map{"error": "Your account is blocked. Contact support for assistance."})
	}

	if err := utils.CheckPassword(user.Password, login.Password); err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid email or password"})
	}

	token, err := middleware.GenerateToken(user.ID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate token"})
	}

	return c.JSON(fiber.Map{"message": "Login successful", "token": token})
}
