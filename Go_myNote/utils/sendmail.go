package utils

import (
	"gopkg.in/gomail.v2"
	"log"
	"os"
)

func SendEmail(to, subject, body string) error {
	message := gomail.NewMessage()
	fromEmail := os.Getenv("MAIL_EMAIL")
	fromPassword := os.Getenv("MAIL_PASSWORD")

	message.SetHeader("From", fromEmail)
	message.SetHeader("To", to)
	message.SetHeader("Subject", subject)
	message.SetBody("text/plain", body)

	d := gomail.NewDialer("smtp.gmail.com", 587, fromEmail, fromPassword)

	log.Printf("Sending email...\nFrom: %s\nTo: %s\nSubject: %s\nBody: %s\n", fromEmail, to, subject, body)

	if err := d.DialAndSend(message); err != nil {
		log.Printf("Failed to send email to %s: %v", to, err)
		return err
	}
	log.Printf("Email successfully sent to %s", to)
	return nil
}
