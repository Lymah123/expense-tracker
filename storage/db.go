package storage

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

func InitDB(filepath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", filepath)
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
		date TEXT
		);`
		if _, err := db.Exec(expenseQuery); err != nil {
			return nil, err
		}

// Create category_budgets table if not exists
budgetsQuery := `
CREATE TABLE IF NOT EXISTS category_budgets (
   category VAR CHAR(50) PRIMARY KEY,
	 budget_amount DECIMAL(10, 2)
	 );`

	 if _, err := db.Exec(budgetsQuery); err != nil {
		return nil, err
	 }

	 return db, nil
}


// The InitDB function initializes a SQLite database connection, creates an expensees table if it doesn't exist, returns the database connection.
