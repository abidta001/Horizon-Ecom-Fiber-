package models

import (
	"horizon/config"
)

type User struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Phone    string `json:"phone"`
	Verified bool   `json:"verified"`
	Blocked  bool   `json:"blocked"`
}
type UserDet struct {
	ID       int    `db:"id"`
	Name     string `db:"name"`
	Email    string `db:"email"`
	Verified bool   `db:"verified"`
}

func (u *UserDet) SaveOrUpdate() error {
	var existingUser UserDet
	query := "SELECT id FROM users WHERE email = $1"
	err := config.DB.Get(&existingUser, query, u.Email)

	if err != nil {
		query = "INSERT INTO users (name, email, verified) VALUES ($1, $2, $3) RETURNING id"
		err = config.DB.QueryRow(query, u.Name, u.Email, u.Verified).Scan(&u.ID)
		if err != nil {
			return err
		}
	} else {
		query = "UPDATE users SET name = $1 WHERE email = $2"
		_, err = config.DB.Exec(query, u.Name, u.Email)
	}
	return err
}
func (u *UserDet) FindByEmail(email string) error {
	query := "SELECT id, name, email, verified FROM users WHERE email = $1"
	return config.DB.QueryRow(query, email).Scan(&u.ID, &u.Name, &u.Email, &u.Verified)
}
func (u *UserDet) Create() error {
	query := "INSERT INTO users (name, email, verified) VALUES ($1, $2, $3)"
	_, err := config.DB.Exec(query, u.Name, u.Email, u.Verified)
	return err
}
func (u *UserDet) Update() error {
	query := "UPDATE users SET verified = $1 WHERE email = $2"
	_, err := config.DB.Exec(query, u.Verified, u.Email)
	return err
}
