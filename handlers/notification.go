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
		log.Println("Error loading .env file:", err)
	}

	apiKey := os.Getenv("SENDGRID_API_KEY")
	if apiKey == "" {
		log.Println("Missing SENDGRID_API_KEY in .env file")
		return fmt.Errorf("missing SENDGRID_API_KEY in .env file")
	}
	from := mail.NewEmail("Expense Tracker", "fimihanodunola625@gmail.com")
	subject := "Budget Alert!"
	to := mail.NewEmail("User", email)
	content := mail.NewContent("text/plain", message)

	m := mail.NewV3MailInit(from, subject, to, content)
	client := sendgrid.NewSendClient(apiKey)
	response, err := client.Send(m)

	if err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}

	log.Printf("Budget alert email sent to %s (status code: %d)", email, response.StatusCode)
	return nil
}
