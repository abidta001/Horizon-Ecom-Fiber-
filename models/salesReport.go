package models

type SalesItem struct {
	ProductID           int     `db:"product_id"`
	ProductName         string  `db:"product_name"`
	ProductDescription  string  `db:"product_description"`
	TotalQuantity       int     `db:"total_quantity"`
	TotalAmount         float64 `db:"total_amount"`
	TotalOfferDiscount  float64 `db:"total_offer_discount"`
	TotalCouponDiscount float64 `db:"total_coupon_discount"`
	TotalRevenue        float64 `db:"total_revenue"`
}

type SalesReport struct {
	CompanyName     string
	PreparedBy      string
	DateRange       string
	TotalSalesCount int
	TotalRevenue    float64
	TotalDiscount   float64
	Items           []SalesItem
}
