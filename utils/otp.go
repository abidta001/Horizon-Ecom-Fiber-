package utils

import (
	"fmt"
	"log"
	"net/smtp"
	"os"
	"sync"
	"time"

	"golang.org/x/exp/rand"
)

const (
	otpLength   = 6
	otpValidity = 5 * time.Minute
	resendDelay = 30 * time.Second
)

var otpStore = sync.Map{}

type OTPData struct {
	OTP          string
	Expiration   time.Time
	LastSentTime time.Time
}

func GenerateOTP() string {
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}

func StoreOTP(email, otp string) {
	otpData := OTPData{
		OTP:          otp,
		Expiration:   time.Now().Add(otpValidity),
		LastSentTime: time.Now(),
	}
	otpStore.Store(email, otpData)

	time.AfterFunc(otpValidity, func() {
		otpStore.Delete(email)
	})
}

func SendEmail(to, subject, body string) error {
	from := os.Getenv("SMTP_EMAIL")
	password := os.Getenv("SMTP_PASSWORD")
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")

	if from == "" || password == "" || smtpHost == "" || smtpPort == "" {
		return fmt.Errorf("SMTP configuration is incomplete")
	}

	message := fmt.Sprintf("Subject: %s\r\n\r\n%s", subject, body)
	auth := smtp.PlainAuth("", from, password, smtpHost)

	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{to}, []byte(message))
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	return nil
}

func ResendOTP(email string) error {
	value, exists := otpStore.Load(email)
	var otpData OTPData

	if exists {
		otpData = value.(OTPData)

		if time.Since(otpData.LastSentTime) < resendDelay {
			return fmt.Errorf("you can only resend otp after %d seconds", resendDelay/time.Second)
		}

		otpData.LastSentTime = time.Now()
		otpStore.Store(email, otpData)
	} else {

		newOTP := GenerateOTP()
		StoreOTP(email, newOTP)

		otpData = OTPData{
			OTP:          newOTP,
			Expiration:   time.Now().Add(otpValidity),
			LastSentTime: time.Now(),
		}
	}

	body := fmt.Sprintf("Your OTP is: %s. It is valid for 5 minutes.", otpData.OTP)
	err := SendEmail(email, "Resend OTP", body)
	if err != nil {
		log.Printf("Failed to resend OTP: %v", err)
		return fmt.Errorf("failed to resend OTP")
	}

	return nil
}

func ValidateOTP(email, otp string) bool {
	value, exists := otpStore.Load(email)
	if !exists {
		return false
	}

	otpData := value.(OTPData)

	if otpData.OTP == otp && time.Now().Before(otpData.Expiration) {

		otpStore.Delete(email)
		return true
	}

	return false
}
