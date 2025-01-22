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

	// Create table if not exist
	query := `
	CREATE TABLE IF NOT EXISTS expenses (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		amount REAL,
		category TEXT,
		description TEXT,
		date TEXT
		);`
		if _, err := db.Exec(query); err != nil {
			return nil, err
		}

		return db, nil
}

// The InitDB function initializes a SQLite database connection, creates an expensees table if it doesn't exist, returns the database connection.
