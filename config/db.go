package config

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/plutov/paypal/v4"
)

var DB *sqlx.DB

func InitDB() {
	dsn := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_USER"), os.Getenv("DB_PASS"), os.Getenv("DB_NAME"))

	var err error
	DB, err = sqlx.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	if err := DB.Ping(); err != nil {
		log.Fatalf("Database is unreachable: %v", err)
	}

	log.Println("Database connected successfully")
	AutoMigrate()
}

func AutoMigrate() {

	if err := executeSQLFile("sql/users.sql"); err != nil {
		log.Fatalf("Failed to create users table: %v", err)
	}
	if err := executeSQLFile("sql/categories.sql"); err != nil {
		log.Fatalf("Failed to create categories table: %v", err)
	}
	if err := executeSQLFile("sql/product.sql"); err != nil {
		log.Fatalf("Failed to create product table: %v", err)
	}
	if err := executeSQLFile("sql/address.sql"); err != nil {
		log.Fatalf("Failed to create address table: %v", err)
	}
	if err := executeSQLFile("sql/wishlist.sql"); err != nil {
		log.Fatalf("Failed to create wishlist table: %v", err)
	}
	if err := executeSQLFile("sql/cart.sql"); err != nil {
		log.Fatalf("Failed to create cart table: %v", err)
	}
	if err := executeSQLFile("sql/order.sql"); err != nil {
		log.Fatalf("Failed to create order table: %v", err)
	}
	if err := executeSQLFile("sql/order_items.sql"); err != nil {
		log.Fatalf("Failed to create order_items table: %v", err)
	}
	if err := executeSQLFile("sql/offer.sql"); err != nil {
		log.Fatalf("Failed to create offers table: %v", err)
	}
	if err := executeSQLFile("sql/coupons.sql"); err != nil {
		log.Fatalf("Failed to create coupons table: %v", err)
	}
	if err := executeSQLFile("sql/wallet.sql"); err != nil {
		log.Fatalf("Failed to create wallet table: %v", err)
	}

}
func executeSQLFile(filePath string) error {
	query, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("could not read SQL file %s: %v", filePath, err)
	}

	_, err = DB.Exec(string(query))
	if err != nil {
		return fmt.Errorf("could not execute SQL file %s: %v", filePath, err)
	}

	log.Printf("Executed SQL file: %s", filePath)
	return nil
}

var (
	JWT_SECRET_ADMIN = os.Getenv("JWT_SECRET_ADMIN")
)
var JWTSecret = os.Getenv("JWT_SECRET")

func GetPayPalClient() *paypal.Client {
	clientID := os.Getenv("PAYPAL_CLIENT")     
	clientSecret := os.Getenv("PAYPAL_SECRET") 

	client, err := paypal.NewClient(clientID, clientSecret, paypal.APIBaseSandBox) 
	if err != nil {
		log.Fatalf("Failed to create PayPal client: %v", err)
	}

	return client
}
