package models

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type BlockRequest struct {
	ID int `json:"id"`
}

type Admin struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
}
type UserView struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Verified bool   `json:"verified"`
	Blocked  bool   `json:"blocked"`
}
