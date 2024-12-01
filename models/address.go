package models

type Address struct {
	ID          int    `json:"id"`
	UserID      int    `json:"user_id"`
	AddressLine string `json:"address_line"`
	City        string `json:"city"`
	ZipCode     string `json:"zip_code"`
}
