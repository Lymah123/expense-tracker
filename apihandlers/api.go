package apihandlers

import (
    "database/sql"
    "encoding/json"
    "net/http"
)

// OverviewData represents the data for the overview section.
type OverviewData struct {
    TotalExpenses   float64 `json:"totalExpenses"`
    BudgetRemaining float64 `json:"budgetRemaining"`
}

// Transaction represents a single transaction.
type Transaction struct {
    Category string  `json:"category"`
    Amount   float64 `json:"amount"`
    Date     string  `json:"date"`
}

// ExpenseDistributionData represents the data for the expense distribution charts
type ExpenseDistributionData struct {
    Categories []string  `json:"categories"`
    Amounts    []float64 `json:"amounts"`
}

func OverviewHandler(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")

			var totalExpenses float64
			var budgetRemaining float64

			// Query to get the tatal expenses
			err := db.QueryRow("SELECT COALESCE(SUM(amount), 0) FROM expenses").Scan(&totalExpenses)
			if err != nil {
				http.Error(w, "Error getting total expenses", http.StatusInternalServerError)
				return
			}

			// Query to get the budget remaining
			err = db.QueryRow("SELECT COALESCE(monthly_budget, 0) FROM budget ORDER BY created_at DESC LIMIT 1 ").Scan(&budgetRemaining)
			if err != nil {
				http.Error(w, "Error fetching budget", http.StatusInternalServerError)
				return
			}

			// Remaining budget is Budget - Total Expenses
			budgetRemaining -= totalExpenses

			data := OverviewData{
				TotalExpenses: totalExpenses,
				BudgetRemaining: budgetRemaining,
			}

			json.NewEncoder(w).Encode(data)
		}
	}

func RecentTransactionsHandler(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")

        // Query to fetch the 5 most recent transactions
        rows, err := db.Query("SELECT category, amount, date FROM expenses ORDER BY date DESC LIMIT 5")
				if err != nil {
					http.Error(w, "Error fetching transactions", http.StatusInternalServerError)
					return
				}
				defer rows.Close()

				var transactions []Transaction
				for rows.Next() {
					var t Transaction
					if err := rows.Scan(&t.Category, &t.Amount, &t.Date); err != nil {
						http.Error(w, "Error scanning transaction", http.StatusInternalServerError)
						return
					}
					transactions = append(transactions, t)
				}
				json.NewEncoder(w).Encode(map[string]interface{}{"transactions": transactions})
    }
}

func ExpenseDistributionHandler(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")

        // Query to get total amount spent per category
        rows, err := db.Query("SELECT category, SUM(amount) FROM expenses GROUP BY category ORDER BY SUM(amount) DESC")
				if err != nil {
					http.Error(w, "Error fetching expense distribution", http.StatusInternalServerError)
					return
				}
				defer rows.Close()

				var categories []string
				var amounts []float64

				for rows.Next() {
					var category string
					var amount float64
					if err := rows.Scan(&category, &amount); err != nil {
						http.Error(w, "Error scanning data", http.StatusInternalServerError)
						return
					}
					if category != "" {
						categories = append(categories, category)
						amounts = append(amounts, amount)
				}
			}
				data := ExpenseDistributionData{
					Categories: categories,
					Amounts: amounts,
				}
				json.NewEncoder(w).Encode(data)
    }
}
