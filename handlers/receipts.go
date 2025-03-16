package handlers

import (
	"database/sql"
	"expense-tracker/models"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"fmt"
)

type ReceiptsPageData struct {
	Receipts []models.Receipt
}

func CreateReceiptsTable(db *sql.DB) error {
	// Start transaction
	tx, err := db.Begin()
	if err != nil {
			return fmt.Errorf("failed to start transaction: %v", err)
	}
	defer tx.Rollback()

	// First check if the table exists and has the right schema
	if _, err := tx.Exec(`SELECT uploaded_at FROM receipts LIMIT 1`); err != nil {
			// If there's an error, the table might not exist or might not have the uploaded_at column
			// Drop the table and recreate it
			if _, err := tx.Exec(`DROP TABLE IF EXISTS receipts`); err != nil {
					return fmt.Errorf("failed to drop old receipts table: %v", err)
			}

			// Create the table with the correct schema
			tableQuery := `
			CREATE TABLE receipts (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					expense_id INTEGER,
					file_path TEXT NOT NULL,
					uploaded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
					FOREIGN KEY (expense_id) REFERENCES expenses(id)
			);`

			if _, err := tx.Exec(tableQuery); err != nil {
					return fmt.Errorf("failed to create receipts table: %v", err)
			}
	}

	// Create indexes
	if _, err := tx.Exec(`CREATE INDEX IF NOT EXISTS idx_receipts_expense_id ON receipts(expense_id);`); err != nil {
			return fmt.Errorf("failed to create expense_id index: %v", err)
	}

	if _, err := tx.Exec(`CREATE INDEX IF NOT EXISTS idx_receipts_uploaded_at ON receipts(uploaded_at);`); err != nil {
			return fmt.Errorf("failed to create uploaded_at index: %v", err)
	}

	// Commit all changes
	if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}

// ReceiptsHandler handles the /receipts route
func ReceiptsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles(filepath.Join("templates", "receipts.html"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		receiptsData, err := GetReceipts(db)
		if err != nil {
			log.Println("Error retrieving receipts:", err)
			http.Error(w, "Error retrieving receipts", http.StatusInternalServerError)
			return
		}

		// Create a data structure with a Receipts field
		pageData := ReceiptsPageData{
			Receipts: receiptsData,
		}

		// Pass the structured data to the template
		if err = tmpl.Execute(w, pageData); err != nil {
			log.Printf("Template execution error: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func GetReceipts(db *sql.DB) ([]models.Receipt, error) {
	rows, err := db.Query(`
	SELECT id, expense_id, file_path, uploaded_at
	FROM receipts
	ORDER BY uploaded_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var receipts []models.Receipt
	for rows.Next() {
		var r models.Receipt
		err := rows.Scan(&r.Id, &r.ExpenseId, &r.FilePath, &r.UploadedAt)
		if err != nil {
			return nil, err
		}
		receipts = append(receipts, r)
	}
	return receipts, nil
}
