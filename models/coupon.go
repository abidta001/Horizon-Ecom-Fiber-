package models

type Coupon struct {
	ID                 int     `json:"id"`
	Code               string  `json:"code"`
	DiscountPercentage float64 `json:"discount_percentage"`
	MaxDiscountAmount  float64 `json:"max_discount_amount"`
	MinOrderAmount     float64 `json:"min_order_amount"`
	StartDate          string  `json:"start_date"`
	EndDate            string  `json:"end_date"`
	UsageLimit         int     `json:"usage_limit"`
	UsedCount          int     `json:"used_count"`
	CreatedAt          string  `json:"created_at"`
	UpdatedAt          string  `json:"updated_at"`
}
