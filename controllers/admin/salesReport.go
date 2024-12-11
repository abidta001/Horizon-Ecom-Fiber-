package admin

import (
	"fmt"
	"horizon/config"
	"horizon/models"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jung-kurt/gofpdf"
	"github.com/xuri/excelize/v2"
)

func fetchSalesData(startDate, endDate time.Time) ([]models.SalesItem, int, float64, float64, error) {
	query := `
		SELECT 
			p.id AS product_id,
			p.name AS product_name,
			p.description AS product_description,
			SUM(oi.quantity) AS total_quantity,
			SUM(oi.subtotal) AS total_amount,
			SUM(o.offer_discount) AS total_offer_discount,
			SUM(o.coupon_discount) AS total_coupon_discount,
			SUM(oi.subtotal - o.offer_discount - o.coupon_discount) AS total_revenue
		FROM order_items oi
		JOIN orders o ON oi.order_id = o.id
		JOIN products p ON oi.product_id = p.id
		WHERE o.order_date BETWEEN $1 AND $2
		  AND o.status IN ('Delivered', 'Returned', 'Canceled')
		GROUP BY p.id, p.name, p.description
	`
	rows, err := config.DB.Queryx(query, startDate, endDate)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	defer rows.Close()

	var items []models.SalesItem
	var totalSalesCount int
	var totalRevenue float64
	var totalDiscount float64

	for rows.Next() {
		var item models.SalesItem
		if err := rows.StructScan(&item); err != nil {
			return nil, 0, 0, 0, err
		}

		totalSalesCount += item.TotalQuantity
		totalRevenue += item.TotalRevenue
		totalDiscount += item.TotalOfferDiscount + item.TotalCouponDiscount

		items = append(items, item)
	}

	return items, totalSalesCount, totalRevenue, totalDiscount, nil
}

func GenerateSalesReport(c *fiber.Ctx) error {
	timeFilter := c.Query("timeFilter")
	fileType := c.Query("fileType")

	customStart := c.Query("customStart")
	customEnd := c.Query("customEnd")
	var startDate, endDate time.Time
	var err error

	switch timeFilter {
	case "daily":
		startDate = time.Now().Truncate(24 * time.Hour)
		endDate = time.Now()
	case "this week":
		startDate = time.Now().AddDate(0, 0, -int(time.Now().Weekday()))
		endDate = time.Now()
	case "this month":
		startDate = time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.Local)
		endDate = time.Now()
	case "this year":
		startDate = time.Date(time.Now().Year(), 1, 1, 0, 0, 0, 0, time.Local)
		endDate = time.Now()
	case "custom":
		startDate, err = time.Parse("2006-01-02", customStart)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid custom start date")
		}
		endDate, err = time.Parse("2006-01-02", customEnd)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid custom end date")
		}
	default:
		return c.Status(fiber.StatusBadRequest).SendString("Invalid time filter")
	}

	items, totalSalesCount, totalRevenue, totalDiscount, err := fetchSalesData(startDate, endDate)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to fetch sales data")
	}

	report := models.SalesReport{
		CompanyName:     "HORIZON ECOM",
		PreparedBy:      "ABID T A",
		DateRange:       fmt.Sprintf("%s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")),
		TotalSalesCount: totalSalesCount,
		TotalRevenue:    totalRevenue,
		TotalDiscount:   totalDiscount,
		Items:           items,
	}

	switch fileType {
	case "pdf":
		err = generateSalesReportPDF(report)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to generate PDF report")
		}
		return c.SendFile("sales_report.pdf")
	case "excel":
		err = generateSalesReportExcel(report)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to generate Excel report")
		}
		return c.SendFile("sales_report.xlsx")
	default:
		return c.Status(fiber.StatusBadRequest).SendString("Invalid file type")
	}
}

func generateSalesReportPDF(report models.SalesReport) error {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(200, 10, fmt.Sprintf("Sales Report for %s", report.CompanyName))
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 12)
	pdf.Cell(100, 10, fmt.Sprintf("Prepared By: %s", report.PreparedBy))
	pdf.Ln(6)
	pdf.Cell(100, 10, fmt.Sprintf("Date Range: %s", report.DateRange))
	pdf.Ln(10)

	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(20, 10, "Item No")
	pdf.Cell(50, 10, "Item Name")
	pdf.Cell(25, 10, "Quantity")
	pdf.Cell(25, 10, "Amount")
	pdf.Cell(25, 10, "Offer")
	pdf.Cell(25, 10, "Coupon")
	pdf.Cell(25, 10, "Total")
	pdf.Ln(6)

	// Group data by product and display only unique products
	for _, item := range report.Items {
		pdf.SetFont("Arial", "", 10)
		pdf.Cell(20, 10, fmt.Sprintf("%d", item.ProductID))
		pdf.Cell(50, 10, item.ProductName)
		pdf.Cell(25, 10, fmt.Sprintf("%d", item.TotalQuantity))
		pdf.Cell(25, 10, fmt.Sprintf("%.2f", item.TotalAmount))
		pdf.Cell(25, 10, fmt.Sprintf("%.2f", item.TotalOfferDiscount))
		pdf.Cell(25, 10, fmt.Sprintf("%.2f", item.TotalCouponDiscount))
		pdf.Cell(25, 10, fmt.Sprintf("%.2f", item.TotalRevenue))
		pdf.Ln(6)
	}

	pdf.Ln(10)

	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(100, 10, fmt.Sprintf("Total Sales Count: %d", report.TotalSalesCount))
	pdf.Ln(6)
	pdf.Cell(100, 10, fmt.Sprintf("Total Discount: %.2f", report.TotalDiscount))
	pdf.Ln(6)
	pdf.Cell(100, 10, fmt.Sprintf("Total Revenue: %.2f", report.TotalRevenue))

	err := pdf.OutputFileAndClose("sales_report.pdf")
	if err != nil {
		return err
	}

	log.Println("Sales report generated successfully")
	return nil
}
func generateSalesReportExcel(report models.SalesReport) error {
	f := excelize.NewFile()

	sheetName := "Sales Report"
	f.NewSheet(sheetName)
	f.SetCellValue(sheetName, "A1", "Sales Report")
	f.SetCellValue(sheetName, "A2", "Company Name: "+report.CompanyName)
	f.SetCellValue(sheetName, "A3", "Prepared By: "+report.PreparedBy)
	f.SetCellValue(sheetName, "A4", "Date Range: "+report.DateRange)

	f.SetCellValue(sheetName, "A6", "Item Number")
	f.SetCellValue(sheetName, "B6", "Item Name")
	f.SetCellValue(sheetName, "C6", "Quantity")
	f.SetCellValue(sheetName, "D6", "Amount")

	row := 7

	for _, item := range report.Items {
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), item.ProductID)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), item.ProductName)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), item.TotalQuantity)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), item.TotalRevenue)
		row++
	}

	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row+1), fmt.Sprintf("Total Sales Count: %d", report.TotalSalesCount))
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row+2), fmt.Sprintf("Total Revenue: %.2f", report.TotalRevenue))

	err := f.SaveAs("sales_report.xlsx")
	if err != nil {
		return err
	}

	return nil
}
