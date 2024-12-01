package models

type OrderDetail struct {
	OrderID        int     `json:"order_id"`
	ReferenceID    string  `json:"reference_id"`
	UserID         int     `json:"user_id"`
	UserName       string  `json:"user_name"`
	UserEmail      string  `json:"user_email"` 
	ProductID      int     `json:"product_id"`
	ProductName    string  `json:"product_name"`
	CategoryName   string  `json:"category_name"`
	Quantity       int     `json:"quantity"`
	Subtotal       float64 `json:"subtotal"`
	CouponDiscount float64 `json:"coupon_discount"`
	OfferDiscount  float64 `json:"offer_discount"`
	TotalAmount    float64 `json:"total_amount"` 
	FinalAmount    float64 `json:"final_amount"` 
	PaymentStatus  string  `json:"payment_status"`
	PaymentMethod  string  `json:"payment_method"`
	OrderStatus    string  `json:"order_status"`
	OrderDate      string  `json:"order_date"`
	AddressLine    string  `json:"address_line"`
	City           string  `json:"city"`
	ZipCode        string  `json:"zip_code"`
}
