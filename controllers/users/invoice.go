package users

import (
	"fmt"
	"horizon/config"
	"horizon/models"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/jung-kurt/gofpdf"
)

func fetchInvoiceData(orderID int) (models.Invoice, error) {
	query := `
		SELECT 
			o.id AS invoice_id, 
			u.name AS user_name, 
			u.email AS user_email,
			concat(a.address_line ,', ',a.city ,' ',a.zip_code )AS user_address,   -- Get address line from address table
			u.phone AS user_phone,     -- Get user phone from users table
			o.order_date, 
			o.payment_method, 
			o.total_amount, 
			o.offer_discount, 
			o.coupon_discount,
			(o.offer_discount + o.coupon_discount) AS total_discount,
			oi.quantity, 
			p.name AS product_name, 
			p.description AS product_desc,
			oi.price AS price_per_unit, 
			oi.subtotal
		FROM orders o
		JOIN users u ON o.user_id = u.id
		JOIN addresses a ON u.id = a.user_id  -- Join the address table with users table
		JOIN order_items oi ON oi.order_id = o.id
		JOIN products p ON oi.product_id = p.id
		WHERE o.id = $1
	`

	rows, err := config.DB.Queryx(query, orderID)
	if err != nil {
		return models.Invoice{}, err
	}
	defer rows.Close()

	var invoice models.Invoice
	var items []models.InvoiceItem
	var totalInvoiceSubtotal float64

	for rows.Next() {
		var item models.InvoiceItem

		err := rows.Scan(&invoice.InvoiceID, &invoice.UserName, &invoice.UserEmail,
			&invoice.UserAddress, &invoice.UserPhoneNumber, &invoice.OrderDate,
			&invoice.PaymentMethod, &invoice.TotalAmount, &invoice.OfferDiscount,
			&invoice.CouponDiscount, &invoice.TotalDiscount, &item.Quantity, &item.ProductName,
			&item.ProductDesc, &item.PricePerUnit, &item.Subtotal)
		if err != nil {
			return models.Invoice{}, err
		}

		totalInvoiceSubtotal += item.Subtotal

		items = append(items, item)
	}

	invoice.Subtotal = totalInvoiceSubtotal

	invoice.Items = items
	return invoice, nil
}

func generateInvoicePDF(invoice models.Invoice) error {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(190, 10, "Invoice")
	pdf.Ln(10)

	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(190, 10, "Company Name: Horizon Ecommerce")
	pdf.Ln(6)
	pdf.Cell(190, 10, "Address: 325-A, Sector 7, Noida")
	pdf.Ln(6)
	pdf.Cell(190, 10, "Email: horizonecom@gmail.com")
	pdf.Ln(6)
	pdf.Cell(190, 10, "Phone: 8078921231")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 12)
	pdf.Cell(100, 10, fmt.Sprintf("Customer: %s", invoice.UserName))
	pdf.Ln(6)
	pdf.Cell(100, 10, fmt.Sprintf("Email: %s", invoice.UserEmail))
	pdf.Ln(6)

	pdf.Cell(100, 10, fmt.Sprintf("Address: %s", invoice.UserAddress))
	pdf.Ln(6)
	pdf.Cell(100, 10, fmt.Sprintf("Phone: %s", invoice.UserPhoneNumber))
	pdf.Ln(6)

	pdf.Cell(100, 10, fmt.Sprintf("Invoice Date: %s", invoice.OrderDate.Format("2006-01-02")))
	pdf.Ln(15)

	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(40, 10, "Product Name")
	pdf.Cell(40, 10, "Description")
	pdf.Cell(25, 10, "Quantity")
	pdf.Cell(25, 10, "Price")
	pdf.Cell(25, 10, "Subtotal")
	pdf.Cell(30, 10, "Payment Method")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 10)
	for _, item := range invoice.Items {
		pdf.Cell(40, 10, item.ProductName)
		pdf.Cell(40, 10, item.ProductDesc)
		pdf.Cell(25, 10, fmt.Sprintf("%d", item.Quantity))
		pdf.Cell(25, 10, fmt.Sprintf("%.2f", item.PricePerUnit))
		pdf.Cell(25, 10, fmt.Sprintf("%.2f", item.Subtotal))
		pdf.Cell(30, 10, invoice.PaymentMethod)
		pdf.Ln(10)
	}

	pdf.Ln(10)
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(100, 15, fmt.Sprintf("Subtotal: %.2f", invoice.TotalAmount))
	pdf.Ln(6)
	pdf.Cell(100, 15, fmt.Sprintf("Offer Discount: %.2f", invoice.OfferDiscount))
	pdf.Ln(6)
	pdf.Cell(100, 15, fmt.Sprintf("Coupon Discount: %.2f", invoice.CouponDiscount))
	pdf.Ln(6)
	pdf.Cell(100, 15, fmt.Sprintf("Total Discount: %.2f", invoice.TotalDiscount))
	pdf.Ln(6)
	pdf.Cell(100, 15, fmt.Sprintf("Total Amount: %.2f", invoice.TotalAmount))

	pdf.Ln(50)
	pdf.SetFont("Arial", "I", 8)
	pdf.SetX(55)
	pdf.MultiCell(100, 5, "horizonecom@gmail.com  |  www.horizonecom.tech.com", "", "C", false)
	return pdf.OutputFileAndClose("invoice.pdf")
}
func GetInvoice(c *fiber.Ctx) error {
	orderID := c.Params("orderID")
	if orderID == "" {
		return c.Status(fiber.StatusBadRequest).SendString("Order ID is required")
	}

	id, err := strconv.Atoi(orderID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid Order ID")
	}

	invoice, err := fetchInvoiceData(id)
	if err != nil {
		fmt.Println(err)
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to fetch invoice data")
	}

	err = generateInvoicePDF(invoice)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to generate invoice PDF")
	}

	return c.SendFile("invoice.pdf")
}
