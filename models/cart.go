package models

type CartItem struct {
	ID        int     `json:"id"`
	UserID    int     `json:"user_id"`
	ProductID int     `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	Subtotal  float64 `json:"subtotal"`
}
