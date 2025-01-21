package main

import (
	"log"
	"os"

	"horizon/config"
	"horizon/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Initialize the database
	config.InitDB()

	// Create a new Fiber app
	app := fiber.New()

	// Apply CORS middleware globally to all routes
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*", // Allow all origins
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Content-Type,Authorization",
	}))

	// Initialize routes for users and admins
	routes.UserRoutes(app)
	routes.AdminRoutes(app)

	// Initialize Google OAuth
	config.InitGoogleOAuth()

	// Get the port from the environment or use default 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start the Fiber app
	log.Fatal(app.Listen(":" + port))
}
