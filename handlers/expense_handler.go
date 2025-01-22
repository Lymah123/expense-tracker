package handler

import (
	"database/sql"
	// "fmt"
	"expense-tracker/models"
)

func AddExpenses(db *sql.DB, expense models.Expense) error {
	query := `INSERT INTO expenses (amount, category, description, date) VALUES (?, ?, ?, ?)`
	_, err := db.Exec(query, expense.Amount, expense.Category, expense.Description, expense.Date)
	return err
}

func GetExpenses(db *sql.DB) ([]models.Expense, error) {
	query := `SELECT id, amount, category, description, date FROM expenses ORDER BY date DESC`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var expenses []models.Expense
	for rows.Next() {
		var expense models.Expense
		if err := rows.Scan(&expense.Id, &expense.Amount, &expense.Category, &expense.Description, &expense.Date); err != nil {
			return nil, err
		}
		expenses = append(expenses, expense)
	}
	return expenses, nil
}
