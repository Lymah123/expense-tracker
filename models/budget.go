package models

import (
	"database/sql"
	"errors"
)

// Set budget for a category
func SetBudget(db *sql.DB, category string, amount float64) error {
	_, err := db.Exec("INSERT INTO budgets (category, amount) VALUES (?, ?) ON CONFLICT(category) DO UPDATE SET amount = ?", category, amount, amount)
	return err
}

// Get budget for a category
func GetBudget(db *sql.DB, category string) (float64, error) {
	var amount float64
	err := db.QueryRow("SELECT amount FROM budgets WHERE category = ?", category).Scan(&amount)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, err
	}
	return amount, nil
}

// Get category Budget
type Budget struct {
	Category     string
	BudgetAmount float64
}

// GetCategoryBudgets function
func GetCategoryBudgets(db *sql.DB) ([]Budget, error) {
	query := `SELECT category, amount FROM budgets`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var budgets []Budget
	for rows.Next() {
		var budget Budget
		if err := rows.Scan(&budget.Category, &budget.BudgetAmount); err != nil {
			return nil, err
		}
		budgets = append(budgets, budget)
	}
	return budgets, nil
}

// Get total expenses for a category
func GetCategoryExpenses(db *sql.DB, category string) (float64, error) {
	var total float64
	err := db.QueryRow("SELECT COALESCE(SUM(amount), 0) FROM expenses WHERE category = ?", category).Scan(&total)
	if err != nil {
		return 0, err
	}
	return total, nil
}
