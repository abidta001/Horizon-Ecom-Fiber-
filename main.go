package main

import (
	"github.com/joho/godotenv"

	"horizon/config"

	"horizon/routes"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	config.InitDB()

	app := fiber.New()

	routes.UserRoutes(app)
	routes.AdminRoutes(app)
	config.InitGoogleOAuth()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Fatal(app.Listen(":" + port))
}
