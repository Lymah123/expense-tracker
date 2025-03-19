package models

import (
	"database/sql"
	"fmt"
	"time"
)

type RecurringExpense struct {
	Id             int       `json:"id"`
	Amount         float64   `json:"amount"`
	Category       string    `json:"category"`
	Description    string    `json:"description"`
	Frequency      string    `json:"frequency"`
	NextDueDate    time.Time `json:"next_due_date"`
	CurrencyCode   string    `json:"currency_code"`
	CurrencySymbol string    `json:"currency_symbol"`
	IsActive       bool      `json:"is_active"`
	CreatedAt      time.Time `json:"created_at"`
}

func CreateRecurringExpensesTable(db *sql.DB) error {
	if _, err := db.Exec("PRAGMA foreign_keys = OFF"); err != nil {
		return fmt.Errorf("failed to disable foreign keys: %v", err)
	}
	defer db.Exec("PRAGMA foreign_keys = ON")

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	tableQuery := `
	CREATE TABLE IF NOT EXISTS recurring_expenses (
	Id INTEGER PRIMARY KEY AUTOINCREMENT,
	amount REAL NOT NULL CHECK (amount >= 0),
	category TEXT NOT NULL,
	description TEXT,
	frequency TEXT NOT NULL,
	next_due_date TIMESTAMP NOT NULL,
	currency_code TEXT,
	currency_symbol TEXT,
	is_active INTEGER DEFAULT 1,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (currency_code) REFERENCES currencies (code)
	);`

	if _, err = tx.Exec(tableQuery); err != nil {
		return fmt.Errorf("failed to create recurring_expenses table: %v", err)
	}

	indexQueries := []string{
		`CREATE INDEX IF NOT EXISTS idx_recurring_expenses_category ON recurring_expenses(category)`,
		`CREATE INDEX IF NOT EXISTS idx_recurring_expenses_next_due_date ON recurring_expenses(next_due_date)`,
		`CREATE INDEX IF NOT EXISTS idx_recurring_expenses_is_active ON recurring_expenses(is_active)`,
	}

	for _, query := range indexQueries {
		if _, err = tx.Exec(query); err != nil {
			return fmt.Errorf("failed to create recurring_expenses indexes: %v", err)
		}
	}

	return tx.Commit()
}

func GetRecurringExpenses(db *sql.DB) ([]RecurringExpense, error) {
	rows, err := db.Query("SELECT Id, category, amount, description, frequency,  next_due_date, currency_code, currency_symbol, is_active, created_at FROM recurring_expenses WHERE is_active = 1")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var expenses []RecurringExpense
	for rows.Next() {
		var expense RecurringExpense
		if err := rows.Scan(&expense.Id, &expense.Category, &expense.Amount, &expense.Description, &expense.Frequency, &expense.NextDueDate, &expense.CurrencyCode, &expense.CurrencySymbol, &expense.IsActive, &expense.CreatedAt); err != nil {
			return nil, err
		}
		expenses = append(expenses, expense)
	}

	return expenses, nil
}

func AddRecurringExpense(db *sql.DB, amount float64, category, description, frequency, nextDueDate, currencyCode, currencySymbol string) error {
	_, err := db.Exec("INSERT INTO recurring_expenses (amount, category, description, frequency, next_due_date, currency_code, currency_symbol) VALUES (?, ?, ?, ?, ?, ?, ?)",
		amount, category, description, frequency, nextDueDate, currencyCode, currencySymbol,
	)
	return err
}

func UpdateRecurringExpense(db *sql.DB, expense RecurringExpense) error {
	_, err := db.Exec("UPDATE recurring_expenses SET is_active = ? WHERE Id = ?", expense.IsActive, expense.Id)
	return err
}
