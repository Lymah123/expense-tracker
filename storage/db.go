package storage

import (
	"database/sql"
	_ "modernc.org/sqlite"
	"io"
	"os"
	"time"
)

func InitDB(filepath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", filepath)
	if err != nil {
		return nil, err
	}

	// Create expenses table if not exist
	expenseQuery := `
    CREATE TABLE IF NOT EXISTS expenses (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        amount REAL,
        category TEXT,
        description TEXT,
        date TEXT,
        currency_code TEXT,
				receipt_path TEXT
    );`
	if _, err := db.Exec(expenseQuery); err != nil {
		return nil, err
	}

	// Create category_budgets table if not exists
	budgetsQuery := `
    CREATE TABLE IF NOT EXISTS category_budgets (
        category VARCHAR(50) PRIMARY KEY,
        budget_amount DECIMAL(10, 2)
    );`
	if _, err := db.Exec(budgetsQuery); err != nil {
		return nil, err
	}

	// Create currencies table if not exists
	currenciesQuery := `
    CREATE TABLE IF NOT EXISTS currencies (
        code TEXT PRIMARY KEY,
        symbol TEXT NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );`
	if _, err := db.Exec(currenciesQuery); err != nil {
		return nil, err
	}

	// Insert default currencies if table is empty
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM currencies").Scan(&count)
	if err != nil {
		return nil, err
	}

	if count == 0 {
		defaultCurrencies := []struct {
			code   string
			symbol string
		}{
			{"USD", "$"},
			{"EUR", "€"},
			{"GBP", "£"},
			{"JPY", "¥"},
			{"NGN", "₦"},
		}

		for _, currency := range defaultCurrencies {
			_, err := db.Exec("INSERT INTO currencies (code, symbol) VALUES (?, ?)",
				currency.code, currency.symbol)
			if err != nil {
				return nil, err
			}
		}
	}

	// Create recurring_expenses table if not exists
	recurringExpensesQuery := `
    CREATE TABLE IF NOT EXISTS recurring_expenses (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        amount REAL NOT NULL,
        category TEXT NOT NULL,
        description TEXT,
        frequency TEXT NOT NULL CHECK (frequency IN ('daily', 'weekly', 'monthly', 'yearly')),
        next_due_date TEXT NOT NULL,
        currency_code TEXT DEFAULT 'USD',
        currency_symbol TEXT DEFAULT '$',
        is_active BOOLEAN DEFAULT 1,
        created_at TEXT DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (currency_code) REFERENCES currencies(code)
    );`
	if _, err := db.Exec(recurringExpensesQuery); err != nil {
		return nil, err
	}

	return db, nil
}

// Backup function
func BackupDB(dbPath string) error {
	backupPath := dbPath + "." + time.Now().Format("2025-01-02T15:04:05") + ".bak"
	srcFile, err := os.Open(dbPath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(backupPath)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}
