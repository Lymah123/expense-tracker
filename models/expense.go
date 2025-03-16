package models

import (
	"database/sql"
	"log"
	"time"
	"fmt"
)

type Expense struct {
	Id           int      `json:"id"`
	Amount       float64  `json:"amount"`
	Category     string   `json:"category"`
	Description  string   `json:"description"`
	Date         string   `json:"date"`
	CurrencyCode string   `json:"currency_code"`
	Receipt      *Receipt `json:"receipt,omitempty"`
	ReceiptPath  string   `json:"receipt_path"`
}

type ExpenseDistribution struct {
	Category string  `json:"category"`
	Total    float64 `json:"total"`
}

type DashboardData struct {
	TotalExpenses      float64               `json:"totalExpenses"`
	BudgetRemaining    float64               `json:"budgetRemaining"`
	Distribution       []ExpenseDistribution `json:"distribution"`
	RecentTransactions []Expense             `json:"recentTransactions"`
}

type MonthlyBudget struct {
	Id            int      `json:"id"`
	MonthlyBudget float64  `json:"monthly_budget"`
	Currency      Currency `json:"currency"`
	CreatedAt     string   `json:"created_at"`
}

// These JSON tags are crucial for ensuring that the struct fields are correctly mapped to their respetive keys when the struct is converted to or from JSON, making it easier to work with JSON data.

type Currency struct {
	Code   string `json:"code"`
	Symbol string `json:"symbol"`
}

type RecurringExpense struct {
	Id             int       `json:"id" db:"id"`
	Amount         float64   `json:"amount" db:"amount"`
	Category       string    `json:"category" db:"category"`
	Description    string    `json:"description" db:"description"`
	Frequency      string    `json:"frequency" db:"frequency"`
	NextDueDate    time.Time `json:"next_due_date" db:"next_due_date"`
	CurrencyCode   string    `json:"currency_code" db:"currency"`
	CurrencySymbol string    `json:"currency_symbol" db:"currency_symbol"`
	IsActive       bool      `json:"is_active" db:"is_active"`
	CreatedAt      string    `json:"created_at"`
}

type Receipt struct {
	Id        int64    `json:"id"`
	ExpenseId int64    `json:"expense_id"`
	FilePath  string `json:"file_path"`
	UploadedAt  time.Time `json:"uploaded_at"`
}

type ExchangeRate struct {
	FromCurrency string  `json:"from_currency"`
	ToCurrency   string  `json:"to_currency"`
	Rate         float64 `json:"rate"`
}

// Add a new expense to the database
func AddExpense(db *sql.DB, amount float64, category, description, date, currencyCode string) error {
	query := `INSERT INTO expenses (amount, category, description, date, currency_code) VALUES (?, ?, ?, ?, ?)`
	_, err := db.Exec(query, amount, category, description, date, currencyCode)
	if err != nil {
		log.Println("Error adding expense: ", err)
		return err
	}
	return nil
}

// GetExpenses retrieves all expenses from the database.
func GetExpenses(db *sql.DB) ([]Expense, error) {
	rows, err := db.Query("SELECT id, amount, category, description, date, currency_code FROM expenses")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var expenses []Expense
	for rows.Next() {
		var expense Expense
		if err := rows.Scan(&expense.Id, &expense.Amount, &expense.Category, &expense.Description, &expense.Date, &expense.CurrencyCode); err != nil {
			return nil, err
		}
		expenses = append(expenses, expense)
	}
	return expenses, nil
}

func CreateExpensesTable(db *sql.DB) error {
	tableQuery := `
	CREATE TABLE IF NOT EXISTS expenses (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		amount REAL NOT NULL CHECK (amount >= 0),
		category TEXT NOT NULL,
		description TEXT,
		date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		currency_code TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (category) REFERENCES budgets(category),
		FOREIGN KEY (currency_code) REFERENCES currencies(code)
		);`

		_, err := db.Exec(tableQuery)
		if err != nil {
			return fmt.Errorf("failed to create expenses table: %v", err)
		}

		// Create indexes
		indexQueries := []string{
			`CREATE INDEX IF NOT EXISTS idx_expenses_category ON expenses(category)`,
			`CREATE INDEX IF NOT EXISTS idx_expenses_date ON expenses(date)`,
			`CREATE INDEX IF NOT EXISTS idx_expenses_currency_currency ON expenses(currency_code)`,
		}

		for _, query := range indexQueries {
			_, err := db.Exec(query)
			if err != nil {
				return fmt.Errorf("failed to create expenses indexes: %v", err)
			}
		}

		return nil
		}
