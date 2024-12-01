package models

import "time"

type Invoice struct {
	InvoiceID       int           `json:"invoice_id"`
	UserName        string        `json:"user_name"`
	UserEmail       string        `json:"user_email"`
	UserAddress     string        `json:"user_address"`
	UserPhoneNumber string        `json:"user_phone"`
	OrderDate       time.Time     `json:"order_date"`
	PaymentMethod   string        `json:"payment_method"`
	TotalAmount     float64       `json:"total_amount"`
	OfferDiscount   float64       `json:"offer_discount"`
	CouponDiscount  float64       `json:"coupon_discount"`
	TotalDiscount   float64       `json:"total_discount"`
	Subtotal        float64       `json:"subtotal"`
	Items           []InvoiceItem `json:"items"`
}

type InvoiceItem struct {
	Quantity     int     `json:"quantity"`
	ProductName  string  `json:"product_name"`
	ProductDesc  string  `json:"product_desc"`
	PricePerUnit float64 `json:"price_per_unit"`
	Subtotal     float64 `json:"subtotal"`
}
