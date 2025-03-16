package utils

import (
	"database/sql"
	"expense-tracker/models"
	"fmt"
	"log"
	"time"
)

func ProcessRecurringExpenses(db *sql.DB) {
	for {
		now := time.Now()
		recurringExpenses, err := models.GetRecurringExpenses(db)
		if err != nil {
			log.Println("Error fetching recurring expenses:", err)
			continue
		}

		for _, expense := range recurringExpenses {

			nextDate := expense.NextDueDate

			if now.After(nextDate) {
				err := models.AddExpense(db, expense.Amount, expense.Category, expense.Description, now.Format("2006-01-02"), expense.CurrencyCode)
				if err != nil {
					log.Println("Error adding expense:", err)
					continue
				}

				// Calculate and directly assign the next due date
				expense.NextDueDate = getNextDueDate(nextDate, expense.Frequency)

				err = models.UpdateRecurringExpense(db, expense)
				if err != nil {
					log.Println("Error updating recurring expense:", err)
				}
			}
		}

		time.Sleep(24 * time.Hour)
	}
}

func getNextDueDate(currentDate time.Time, frequency string) time.Time {
	switch frequency {
	case "daily":
		return currentDate.AddDate(0, 0, 1)
	case "weekly":
		return currentDate.AddDate(0, 0, 7)
	case "monthly":
		return currentDate.AddDate(0, 1, 0)
	case "yearly":
		return currentDate.AddDate(1, 0, 0)
	default:
		return currentDate
	}
}

// UpdateExchangeRates fetches latest exchange rates every 24 hours
func UpdateExchangeRates() {
	ticker := time.NewTicker(24 * time.Hour)
	go func() {
		for range ticker.C {
			fmt.Println("Updating exchange rates...")
		}
	}()
}
