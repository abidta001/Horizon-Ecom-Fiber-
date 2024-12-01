package models

type DashboardData struct {
	TotalSales        float64
	TotalOrders       int
	NewUsers          int
	ProductStatistics map[string]int 
	Revenue           float64
	PendingOrders     int
	CompletedOrders   int
	CancelledOrders   int
	ReturnedOrders    int
}
type DashboardReport struct {
	TopProducts   []ProductReport
	TopCategories []CategoryReport
}

type ProductReport struct {
	ProductName string `db:"product_name"`
	Category    string `db:"category_name"`
	TotalSold   int    `db:"total_sold"`
}

type CategoryReport struct {
	CategoryName string `db:"category_name"`
	TotalSold    int    `db:"total_sold"`
}
