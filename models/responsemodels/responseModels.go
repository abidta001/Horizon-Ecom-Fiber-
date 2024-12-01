package responsemodels

type ViewCategory struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type ViewProducts struct {
	ID           int     `json:"id"`
	Name         string  `json:"name"`
	Description  string  `json:"description"`
	Price        float64 `json:"price"`
	CategoryName string  `json:"category_name"`
	Status       string  `json:"status"`
}
type UserProfile struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone"`
}
type AddressUser struct {
	ID          int    `json:"id"`
	AddressLine string `json:"address_line"`
	City        string `json:"city"`
	ZipCode     string `json:"zip_code"`
}
type ViewCartItem struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	Quantity int     `json:"quantity"`
	Subtotal float64 `json:"subtotal"`
}

type OrderDetail struct {
	OrderID        int     `json:"order_id"`
	ReferenceID    string  `json:"reference_id"`
	ProductName    string  `json:"product_name"`
	ProductPrice   float64 `json:"product_price"`
	CategoryName   string  `json:"category_name"`
	Quantity       int     `json:"quantity"`
	Subtotal       float64 `json:"subtotal"`
	CouponDiscount float64 `json:"coupon_discount"`
	OfferDiscount  float64 `json:"offer_discount"`
	AmountPaid     float64 `json:"amount_paid"`
	PaymentStatus  string  `json:"payment_status"`
	OrderStatus    string  `json:"order_status"`
}
