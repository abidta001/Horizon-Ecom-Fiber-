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

	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	config.InitDB()

	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendFile("./portfolio/index.html")
	})

	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Content-Type,Authorization",
	}))

	routes.UserRoutes(app)
	routes.AdminRoutes(app)

	config.InitGoogleOAuth()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Fatal(app.Listen(":" + port))
}
