package utils

import (
	"fmt"

	"gopkg.in/gomail.v2"
)

func SendResetEmail(email string, resetToken string) error {
	resetLink := fmt.Sprintf("http://yourdomain.com/reset-password?token=%s", resetToken)

	m := gomail.NewMessage()
	m.SetHeader("From", "noreply@yourdomain.com")
	m.SetHeader("To", email)
	m.SetHeader("Subject", "Password Reset Request")
	m.SetBody("text/html", fmt.Sprintf(`
		<h1>Password Reset</h1>
		<p>Click the link below to reset your password:</p>
		<a href="%s">%s</a>
	`, resetLink, resetLink))

	d := gomail.NewDialer("smtp.your-email-provider.com", 587, "your-email@example.com", "your-email-password")

	return d.DialAndSend(m)
}
