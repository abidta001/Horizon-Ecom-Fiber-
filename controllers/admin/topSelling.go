package admin

import (
	"bytes"
	"fmt"
	"horizon/config"
	"horizon/models"
	"log"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jung-kurt/gofpdf"
)

func FetchTopSellingData() (models.DashboardReport, error) {
	var data models.DashboardReport

	productQuery := `
		SELECT 
			p.name AS product_name, 
			c.name AS category_name, 
			SUM(oi.quantity) AS total_sold
		FROM 
			order_items oi
		JOIN 
			products p ON oi.product_id = p.id
		JOIN 
			categories c ON p.category_id = c.id
		JOIN 
			orders o ON oi.order_id = o.id
		WHERE 
			 o.payment_status IN ('Paid', 'Completed')
		GROUP BY 
			p.id, c.id
		ORDER BY 
			total_sold DESC
		LIMIT 10;
	`

	rows, err := config.DB.Queryx(productQuery)
	if err != nil {
		return data, err
	}
	defer rows.Close()

	for rows.Next() {
		var product models.ProductReport
		if err := rows.StructScan(&product); err != nil {
			return data, err
		}
		data.TopProducts = append(data.TopProducts, product)
	}

	categoryQuery := `
		SELECT 
			c.name AS category_name, 
			SUM(oi.quantity) AS total_sold
		FROM 
			order_items oi
		JOIN 
			products p ON oi.product_id = p.id
		JOIN 
			categories c ON p.category_id = c.id
		JOIN 
			orders o ON oi.order_id = o.id
		WHERE 
		o.payment_status IN ('Paid', 'Completed')
		GROUP BY 
			c.id
		ORDER BY 
			total_sold DESC
		LIMIT 10;
	`

	rows, err = config.DB.Queryx(categoryQuery)
	if err != nil {
		return data, err
	}
	defer rows.Close()

	for rows.Next() {
		var category models.CategoryReport
		if err := rows.StructScan(&category); err != nil {
			return data, err
		}
		data.TopCategories = append(data.TopCategories, category)
	}

	return data, nil
}

func GenerateTopSellingPDF(data models.DashboardReport) (*bytes.Buffer, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")

	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(190, 10, "Top Selling Products and Categories Report")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 12)
	pdf.Cell(190, 10, fmt.Sprintf("Generated on: %s", time.Now().Format("2006-01-02")))
	pdf.Ln(20)

	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(190, 10, "Top 10 Best-Selling Products")
	pdf.Ln(10)

	for _, product := range data.TopProducts {
		pdf.SetFont("Arial", "", 12)
		pdf.Cell(100, 10, product.ProductName)
		pdf.Cell(60, 10, fmt.Sprintf("Category: %s", product.Category))
		pdf.Cell(30, 10, fmt.Sprintf("Sold: %d", product.TotalSold))
		pdf.Ln(10)
	}

	pdf.Ln(10)

	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(190, 10, "Top 10 Best-Selling Categories")
	pdf.Ln(10)

	for _, category := range data.TopCategories {
		pdf.SetFont("Arial", "", 12)
		pdf.Cell(100, 10, category.CategoryName)
		pdf.Cell(30, 10, fmt.Sprintf("Sold: %d", category.TotalSold))
		pdf.Ln(10)
	}

	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, err
	}

	return &buf, nil
}

func GenerateTopSellingReport(c *fiber.Ctx) error {
	data, err := FetchTopSellingData()
	if err != nil {
		log.Printf("Error fetching top-selling data: %v", err)
		return c.Status(http.StatusInternalServerError).SendString("Failed to fetch top-selling data.")
	}

	pdf, err := GenerateTopSellingPDF(data)
	if err != nil {
		log.Printf("Error generating PDF: %v", err)
		return c.Status(http.StatusInternalServerError).SendString("Failed to generate PDF report.")
	}

	c.Set("Content-Type", "application/pdf")
	return c.Send(pdf.Bytes())
}
