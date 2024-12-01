package models

type SalesItem struct {
	ItemNumber      int     `db:"item_number"`
	ItemName        string  `db:"item_name"`
	ItemDescription string  `db:"item_description"`
	ProductID       int     `db:"product_id"`
	Quantity        int     `db:"quantity"`
	Amount          float64 `db:"amount"`
	Offer           float64 `db:"offer_discount"`
	Coupon          float64 `db:"coupon_discount"`
	TotalDiscount   float64 `db:"total_discount"`
	Total           float64 `db:"total"`
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
