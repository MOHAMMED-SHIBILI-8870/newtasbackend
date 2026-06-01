package usecase

import (
	"errors"
	"fmt"
	"net/smtp"
	"os"
)

//SentToEmail via using SMTP

func SentOTPEmail(toEmail, otp, purpose string) error {
	from := os.Getenv("EMAIL_FROM")
	password := os.Getenv("EMAIL_PASSWORD")
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")

	if from == "" || password == "" || host == "" || port == "" {
		return errors.New("email configuration missing")
	}

	auth := smtp.PlainAuth("", from, password, host)

	subject := "your OTP Code"
	body := fmt.Sprintf(
		"Your OTP for %s is: %s\n\nThis OTP expires in 5 minutes.",
		purpose,
		otp,
	)

	message := []byte(
		"To: " + toEmail + "\r\n" +
			"Subject: " + subject + "\r\n" +
			"\r\n" +
			body,
	)

	return smtp.SendMail(
		host+":"+port,
		auth,
		from,
		[]string{toEmail},
		message,
	)
}
