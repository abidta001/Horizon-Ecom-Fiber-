package users

import (
	"encoding/json"
	"fmt"
	"horizon/config"
	"horizon/models"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/oauth2"
)

func GoogleLogin(c *fiber.Ctx) error {
	url := config.GoogleOAuthConfig.AuthCodeURL("random_state_token", oauth2.AccessTypeOffline)
	return c.Redirect(url)
}

func GoogleCallback(c *fiber.Ctx) error {
	code := c.Query("code")
	if code == "" {
		return c.Status(http.StatusBadRequest).SendString("No code in the callback URL")
	}

	token, err := config.GoogleOAuthConfig.Exchange(c.Context(), code)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Failed to exchange token")
	}

	client := config.GoogleOAuthConfig.Client(c.Context(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Failed to get user info")
	}
	defer resp.Body.Close()

	var userInfo struct {
		Email string `json:"email"`
		Name  string `json:"name"`
		ID    string `json:"id"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Failed to decode user info")
	}

	user := models.UserDet{}
	if err := user.FindByEmail(userInfo.Email); err != nil {
		user.Name = userInfo.Name
		user.Email = userInfo.Email
		user.Verified = true
		if err := user.Create(); err != nil {
			fmt.Println("errr", err)
			return c.Status(http.StatusInternalServerError).SendString("Failed to create user")
		}
	} else {

		user.Verified = true
		if err := user.Update(); err != nil {
			return c.Status(http.StatusInternalServerError).SendString("Failed to update user")
		}
	}

	return c.JSON(fiber.Map{
		"message": "Login successful, Welcome to Horizon!",
	})
}
