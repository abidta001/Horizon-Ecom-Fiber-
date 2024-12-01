package admin

import (
	"fmt"
	"horizon/config"
	"horizon/models"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jung-kurt/gofpdf"
)

func FetchDashboardData(period string) (models.DashboardData, error) {
	var data models.DashboardData
	var interval string

	switch period {
	case "Daily":
		interval = "1 day"
	case "Monthly":
		interval = "1 month"
	case "Yearly":
		interval = "1 year"
	default:
		return data, fmt.Errorf("invalid period: %s", period)
	}

	querySales := fmt.Sprintf(`
		SELECT SUM(total_amount) 
		FROM orders 
		WHERE order_date >= NOW() - INTERVAL '%s'
	`, interval)
	err := config.DB.QueryRow(querySales).Scan(&data.TotalSales)
	if err != nil {
		return data, err
	}

	queryOrders := fmt.Sprintf(`
		SELECT COUNT(id) 
		FROM orders 
		WHERE order_date >= NOW() - INTERVAL '%s'
	`, interval)
	err = config.DB.QueryRow(queryOrders).Scan(&data.TotalOrders)
	if err != nil {
		return data, err
	}

	queryNewUsers := fmt.Sprintf(`
		SELECT COUNT(id) 
		FROM users 
		WHERE created_at >= NOW() - INTERVAL '%s'
	`, interval)
	err = config.DB.QueryRow(queryNewUsers).Scan(&data.NewUsers)
	if err != nil {
		return data, err
	}

	queryProductStats := fmt.Sprintf(`
		SELECT 
   		p.name, 
   		SUM(oi.quantity) AS total_sold
		FROM 
 	    order_items oi
		JOIN 
  		products p ON oi.product_id = p.id
		JOIN 
   		orders o ON oi.order_id = o.id
		WHERE 
    	o.order_date >= NOW() - INTERVAL '%s' 
   		AND (
        o.payment_status IN ('Paid', 'Completed') 
        OR o.status = 'Pending COD Verification'
  	  )
		GROUP BY 
   		 p.name
	`, interval)

	rows, err := config.DB.Query(queryProductStats)
	if err != nil {
		return data, err
	}
	defer rows.Close()

	data.ProductStatistics = make(map[string]int)
	for rows.Next() {
		var productName string
		var quantity int
		err := rows.Scan(&productName, &quantity)
		if err != nil {
			return data, err
		}
		data.ProductStatistics[productName] = quantity
	}

	queryRevenue := fmt.Sprintf(`
		SELECT SUM(total_amount) 
		FROM orders 
		WHERE order_date >= NOW() - INTERVAL '%s'
	`, interval)
	err = config.DB.QueryRow(queryRevenue).Scan(&data.Revenue)
	if err != nil {
		return data, err
	}

	queryPendingOrders := fmt.Sprintf(`
		SELECT COUNT(id)
		FROM orders
		WHERE status IN ('Pending', 'Pending COD Verification')
  		AND payment_status IN ('Completed', 'Paid')
  		AND order_date >= NOW() - INTERVAL '%s'
	`, interval)
	err = config.DB.QueryRow(queryPendingOrders).Scan(&data.PendingOrders)
	if err != nil {
		return data, err
	}
	queryCancelledOrders := fmt.Sprintf(`
		SELECT COUNT(id)
		FROM orders 
		WHERE status='Cancelled'
  		AND order_date >= NOW() - INTERVAL '%s'
	`, interval)
	err = config.DB.QueryRow(queryCancelledOrders).Scan(&data.CancelledOrders)
	if err != nil {
		return data, err
	}
	queryReturnedOrders := fmt.Sprintf(`
		SELECT COUNT(id)
		FROM orders 
		WHERE status='Returned'
  		AND order_date >= NOW() - INTERVAL '%s'
	`, interval)
	err = config.DB.QueryRow(queryReturnedOrders).Scan(&data.ReturnedOrders)
	if err != nil {
		return data, err
	}

	queryCompletedOrders := fmt.Sprintf(`
		SELECT COUNT(id) 
		FROM orders 
		WHERE status = 'Delivered' AND order_date >= NOW() - INTERVAL '%s'
	`, interval)
	err = config.DB.QueryRow(queryCompletedOrders).Scan(&data.CompletedOrders)
	if err != nil {
		return data, err
	}

	return data, nil
}

func createHeader(pdf *gofpdf.Fpdf, period string) {
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(190, 10, fmt.Sprintf("%s Dashboard Report", period))
	pdf.Ln(6)
}

func createDateSection(pdf *gofpdf.Fpdf) {
	pdf.SetFont("Arial", "", 12)
	pdf.Cell(190, 10, fmt.Sprintf("Generated on: %s", time.Now().Format("2006-01-02")))
	pdf.Ln(15)
}

func createSection(pdf *gofpdf.Fpdf, title, value string) {
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(100, 10, title)
	pdf.SetFont("Arial", "", 12)
	pdf.Cell(100, 10, value)
	pdf.Ln(10)
}

func createProductStatistics(pdf *gofpdf.Fpdf, productStats map[string]int) {
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(100, 10, "Product Statistics:")
	pdf.Ln(10)

	nameColumnWidth := 100.0 
	quantityColumnWidth := 40.0

	for product, quantity := range productStats {
		pdf.SetFont("Arial", "", 12)
		pdf.Cell(nameColumnWidth, 10, product)

		pdf.SetFont("Arial", "", 12)
		pdf.Cell(quantityColumnWidth, 10, fmt.Sprintf("%d sold", quantity))
		pdf.Ln(10)
	}
	pdf.Ln(10)
}

func GenerateDashboardPDF(data models.DashboardData, period string) error {

	pdf := gofpdf.New("P", "mm", "A4", "")

	pdf.AddPage()

	createHeader(pdf, period)

	createDateSection(pdf)

	createSection(pdf, "New Users (signups):", fmt.Sprintf("%d", data.NewUsers))

	createProductStatistics(pdf, data.ProductStatistics)

	createSection(pdf, "Pending Orders:", fmt.Sprintf("%d", data.PendingOrders))

	createSection(pdf, "Cancelled Orders:", fmt.Sprintf("%d", data.CancelledOrders))

	createSection(pdf, "Returned Orders:", fmt.Sprintf("%d", data.ReturnedOrders))

	createSection(pdf, "Completed Orders:", fmt.Sprintf("%d", data.CompletedOrders))

	createSection(pdf, "Total Orders:", fmt.Sprintf("%d", data.TotalOrders))

	createSection(pdf, "Total Revenue:", fmt.Sprintf("%.2f", data.Revenue))

	err := pdf.OutputFileAndClose("dashboard_report.pdf")
	if err != nil {
		return fmt.Errorf("error generating PDF: %v", err)
	}

	return nil
}

func GenerateDashboardReport(c *fiber.Ctx) error {

	period := c.Params("period")

	if period != "Daily" && period != "Monthly" && period != "Yearly" {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid period. Valid options are 'daily', 'monthly', or 'yearly'.")
	}

	data, err := FetchDashboardData(period)
	if err != nil {
		log.Printf("Error fetching dashboard data: %v", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Error fetching dashboard data.")
	}

	err = GenerateDashboardPDF(data, period)
	if err != nil {
		log.Printf("Error generating PDF report: %v", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Error generating PDF report.")
	}

	return c.SendFile("dashboard_report.pdf")
}
