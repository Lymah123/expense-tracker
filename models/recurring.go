package models

import (
	"database/sql"
	// "times"
)

// Add a new recurring expense to the database
func GetRecurringExpenses(db *sql.DB) ([]RecurringExpense, error) {
	rows, err := db.Query("SELECT id, category, amount, description, next_due_date, currency_code, currency_symbol, is_active, created_at FROM recurring_expenses WHERE is_active = 1")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var expenses []RecurringExpense
	for rows.Next() {
		var expense RecurringExpense
		if err := rows.Scan(&expense.Id, &expense.Amount, &expense.Category, &expense.Description, &expense.Frequency, &expense.NextDueDate, &expense.CurrencyCode, &expense.CurrencySymbol, &expense.IsActive, &expense.CreatedAt); err != nil {
			return nil, err
		}
		expenses = append(expenses, expense)
	}
	return expenses, nil
}

func AddRecurringExpense(db *sql.DB, amount float64, category, description, frequency, nextDueDate, currencyCode, currencySymbol string) error {
	_, err := db.Exec("INSERT INTO recurring_expenses (amount, category, description, frequency, next_due_date, currency_code, currency_symbol) VALUES (?, ?, ?, ?, ?, ?, ?)", amount, category, description, frequency, nextDueDate, currencyCode, currencySymbol)
	return err
}

func UpdateRecurringExpense(db *sql.DB, expense RecurringExpense) error {
	_, err := db.Exec("UPDATE recurring_expenses SET next_due_date = ?, is_active = ? WHERE id = ?", expense.NextDueDate, expense.IsActive, expense.Id)
	return err
}
