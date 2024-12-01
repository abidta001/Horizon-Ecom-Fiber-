package models

type OTP struct {
	Email string `json:"email"`
	OTP   string `json:"otp"`
}
