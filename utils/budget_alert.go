package utils

import (
	"database/sql"
	"expense-tracker/models"
	"fmt"
)

// Check if a category exceeds budget and return an alert message
func CheckBudgetAlert(db *sql.DB, category string) (string, error) {
	budget, err := models.GetBudget(db, category)
	if err != nil {
		return "", err
	}

	expense, err := models.GetCategoryExpenses(db, category)
	if err != nil {
		return "", err
	}

	if expense > budget {
		return fmt.Sprintf("⚠️ Alert: You have exceeded the budget for %s! Budget: %.2f, Spent: %.2f", category, budget, expense), nil
	}
	return "", nil
}
