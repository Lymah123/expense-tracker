package reports

import (
	"database/sql"
	"fmt"
	"log"
)

func formatCurrency(amount float64) string {
	return fmt.Sprintf("$%.2f", amount)
}

// GenerateReport generates a report based on the report type, start date, end date, and export format.
func GenerateReport(db *sql.DB, reportType, startDate, endDate, exportFormat string) error {
	switch reportType {
	case "monthly":
		return GenerateMonthlyReport(db, startDate, endDate, exportFormat)
	case "weekly":
		return GenerateWeeklyReport(db, startDate, endDate, exportFormat)
	case "yearly":
		return GenerateYearlyReport(db, startDate, endDate, exportFormat)
	default:
		return fmt.Errorf("invalid report type: %s", reportType)
	}
}

func GenerateMonthlyReport(db *sql.DB, startDate, endDate, exportFormat string) error {
	query := `SELECT category, SUM(amount) FROM expenses WHERE date >= ? AND date <= ? GROUP BY category`
	rows, err := db.Query(query, startDate, endDate)
	if err != nil {
		log.Println("Error querying database:", err)
		return err
	}
	defer rows.Close()

	fmt.Println("Monthly Report:")
	for rows.Next() {
		var category string
		var total float64
		if err := rows.Scan(&category, &total); err != nil {
			log.Println("Error scanning row:", err)
			return err
		}
		fmt.Printf("Category: %s, Total: %s\n", category, formatCurrency(total))
	}
	return nil
}

func GenerateWeeklyReport(db *sql.DB, startDate, endDate, exportFormat string) error {
	query := `SELECT category, SUM(amount) FROM expenses WHERE date >= ? AND date <= ? GROUP BY category`
	rows, err := db.Query(query, startDate, endDate)
	if err != nil {
		log.Println("Error querying database:", err)
		return err
	}
	defer rows.Close()

	fmt.Println("Weekly Report:")
	for rows.Next() {
		var category string
		var total float64
		if err := rows.Scan(&category, &total); err != nil {
			log.Println("Error scanning row:", err)
			return err
		}
		fmt.Printf("Category: %s, Total: %s\n", category, formatCurrency(total))
	}
	return nil
}

func GenerateYearlyReport(db *sql.DB, startDate, endDate, exportFormat string) error {
	query := `SELECT category, SUM(amount) FROM expenses WHERE date >= ? AND date <= ? GROUP BY category`
	rows, err := db.Query(query, startDate, endDate)
	if err != nil {
		log.Println("Error querying database:", err)
		return err
	}
	defer rows.Close()

	fmt.Println("Yearly Report:")
	for rows.Next() {
		var category string
		var total float64
		if err := rows.Scan(&category, &total); err != nil {
			log.Println("Error scanning row:", err)
			return err
		}
		fmt.Printf("Category: %s, Total: %s\n", category, formatCurrency(total))
	}
	return nil
}
