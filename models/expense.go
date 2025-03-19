package models

import (
	"database/sql"
	"fmt"
	"log"
	"time"
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

// type RecurringExpense struct {
// 	Id int `json:"id" db:"id"`
// }

// type RecurringExpense struct {
// 	Id             int       `json:"id" db:"id"`
// 	Amount         float64   `json:"amount" db:"amount"`
// 	Category       string    `json:"category" db:"category"`
// 	Description    string    `json:"description" db:"description"`
// 	Frequency      string    `json:"frequency" db:"frequency"`
// 	NextDueDate    time.Time `json:"next_due_date" db:"next_due_date"`
// 	CurrencyCode   string    `json:"currency_code" db:"currency"`
// 	CurrencySymbol string    `json:"currency_symbol" db:"currency_symbol"`
// 	IsActive       bool      `json:"is_active" db:"is_active"`
// 	CreatedAt      string    `json:"created_at"`
// }

type Receipt struct {
	Id         int64     `json:"id"`
	ExpenseId  int64     `json:"expense_id"`
	FilePath   string    `json:"file_path"`
	UploadedAt time.Time `json:"uploaded_at"`
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
	rows, err := db.Query("SELECT Id, amount, category, description, date, currency_code FROM expenses")
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
	rows, err := db.Query("PRAGMA table_info(expenses)")
	if err == nil {
		defer rows.Close()
		fmt.Println("Current expenses table schema:")
		hasColumn := false
		for rows.Next() {
			var cid, notnull, pk int
			var name, type_name string
			var dfItValue interface{}
			if err := rows.Scan(&cid, &name, &type_name, &notnull, &dfItValue, &pk); err != nil {
				fmt.Printf("Error scanning: %v\n", err)
				continue
			}
			fmt.Printf("Column: %s (%s)\n", name, type_name)
			if name == "currency_code" {
				hasColumn = true
			}
		}
		fmt.Printf("Has currency_code column: %v\n", hasColumn)
	}
	if _, err :=db.Exec("PRAGMA foreign_keys = OFF"); err != nil {
		return fmt.Errorf("failed to disable foreign keys: %v", err)
	}
	defer db.Exec("PRAGMA foreign_keys = ON")

	_, err = db.Exec("DROP TABLE IF EXISTS expenses")
	if err != nil {
		return fmt.Errorf("failed to drop expenses table: %v", err)
	}
	// Start a transaction to ensure atomicity
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
	CREATE TABLE IF NOT EXISTS expenses (
		Id INTEGER PRIMARY KEY AUTOINCREMENT,
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

	if _, err = tx.Exec(tableQuery); err != nil {
		return fmt.Errorf("failed to create expenses table: %v", err)
	}

	// Create indexes
	indexQueries := []string{
		`CREATE INDEX IF NOT EXISTS idx_expenses_category ON expenses(category)`,
		`CREATE INDEX IF NOT EXISTS idx_expenses_date ON expenses(date)`,
		`CREATE INDEX IF NOT EXISTS idx_expenses_currency_code ON expenses(currency_code)`,
	}

	for _, query := range indexQueries {
		if _, err = tx.Exec(query); err != nil {
			return fmt.Errorf("failed to create expenses indexes: %v", err)
		}
	}

	return tx.Commit()
}
