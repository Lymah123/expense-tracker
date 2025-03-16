package handlers

import (
	"fmt"
	"os"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"github.com/joho/godotenv"
	"log"
)

func SendBudgetAlertNotifications(email string, message string) error {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}

	apiKey := os.Getenv("SENDGRID_API_KEY")
	if apiKey == "" {
		log.Println("Missing SENDGRID_API_KEY")
		return fmt.Errorf("missing SendGrid API key")
	}

	from := mail.NewEmail("Expense Tracker", "fimihanodunola625@gmail.com")
	subject := "Budget Alert!"
	to := mail.NewEmail("User", email)
	content := mail.NewContent("text/plain", message)

	m := mail.NewV3MailInit(from, subject, to, content)
	client := sendgrid.NewSendClient(apiKey)
	_, err = client.Send(m)

	return err
}
