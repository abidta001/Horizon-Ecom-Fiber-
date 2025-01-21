package main

import (
	"fmt"
	"net/http"

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
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight request
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Example response
		fmt.Fprintf(w, "CORS enabled for all origins!")

	})

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
