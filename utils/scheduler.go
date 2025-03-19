package utils

import (
	"database/sql"
	"expense-tracker/models"
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

			if now.After(expense.NextDueDate) {
				err := models.AddExpense(db, expense.Amount, expense.Category, expense.Description, now.Format("2006-01-02"), expense.CurrencyCode)
				if err != nil {
					log.Println("Error parsing date:", err)
					continue
				}

			}

				// Calculate next due date
				expense.NextDueDate = getNextDueDate(expense.NextDueDate, expense.Frequency)

				err = models.UpdateRecurringExpense(db, expense)
				if err != nil {
					log.Println("Error updating recurring expense:", err)
				}
			}
			// Sleep before the next iteration
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
