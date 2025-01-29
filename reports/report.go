package reports

import (
	"database/sql"
	"fmt"
)
func formatCurrency(amount float64) string {
	return fmt.Sprintf("$%.2f", amount)
}

func GenerateMonthlyReport(db *sql.DB, month string) error {
	query := `SELECT category, SUM(amount) FROM expenses WHERE date LIKE ? GROUP BY category`
	rows, err := db.Query(query, month+"%")
	if err != nil {
			return err
	}
	defer rows.Close()

	fmt.Println("Monthly Report:")
	for rows.Next() {
			var category string
			var total float64
			if err := rows.Scan(&category, &total); err != nil {
					return err
			}
			fmt.Printf("Category: %s, Total: %s\n", category, formatCurrency(total))
	}
	return nil
}
