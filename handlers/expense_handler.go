package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"expense-tracker/models"
	"net/http"
	"strings"
)

// AddExpenses inserts a new expense into the database.
func AddExpenses(db *sql.DB, expense models.Expense) error {
	query := `INSERT INTO expenses (amount, category, description, date, currency_code) VALUES (?, ?, ?, ?, ?)`
	_, err := db.Exec(query, expense.Amount, expense.Category, expense.Description, expense.Date, expense.CurrencyCode)
	return err
}

// GetExpenses retrieves all expenses from the database, ordered by date in descending order.
func GetExpenses(db *sql.DB) ([]models.Expense, error) {
	query := `SELECT Id, amount, category, description, date FROM expenses ORDER BY date DESC`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var expenses []models.Expense
	for rows.Next() {
		var expense models.Expense
		if err := rows.Scan(&expense.Id, &expense.Amount, &expense.Category, &expense.Description, &expense.Date); err != nil {
			return nil, err
		}
		expenses = append(expenses, expense)
	}
	return expenses, nil
}

// GetRecentTransactions retrieves the 5 most recent expenses.
func GetRecentTransactions(db *sql.DB) ([]models.Expense, error) {
	query := `SELECT Id, amount, category, description, date
              FROM expenses
              ORDER BY date DESC
              LIMIT 5`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var expenses []models.Expense
	for rows.Next() {
		var expense models.Expense
		if err := rows.Scan(&expense.Id, &expense.Amount, &expense.Category, &expense.Description, &expense.Date); err != nil {
			return nil, err
		}
		expenses = append(expenses, expense)
	}
	return expenses, nil
}

// GetExpenseDistribution retrieves the distribution of expenses by category.
func GetExpenseDistribution(db *sql.DB) ([]models.ExpenseDistribution, error) {
	query := `SELECT category, SUM(amount) as total FROM expenses GROUP BY category`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var distribution []models.ExpenseDistribution
	for rows.Next() {
		var dist models.ExpenseDistribution
		if err := rows.Scan(&dist.Category, &dist.Total); err != nil {
			return nil, err
		}
		distribution = append(distribution, dist)
	}
	return distribution, nil
}

// GetDashboardData retrieves all data needed for the dashboard.
func GetDashboardData(db *sql.DB) (models.DashboardData, error) {
	var dashboard models.DashboardData

	// Get recent transactions
	recent, err := GetRecentTransactions(db)
	if err != nil {
		return dashboard, err
	}
	dashboard.RecentTransactions = recent

	// Get distribution
	dist, err := GetExpenseDistribution(db)
	if err != nil {
		return dashboard, err
	}
	dashboard.Distribution = dist

	// Calculate total expenses
	var total float64
	row := db.QueryRow("SELECT SUM(amount) FROM expenses")
	if err := row.Scan(&total); err != nil && err != sql.ErrNoRows {
		return dashboard, err
	}
	dashboard.TotalExpenses = total

	const monthlyBudget float64 = 5000.00 // Example budget
	dashboard.BudgetRemaining = monthlyBudget - total

	return dashboard, nil
}

// Search and filtering functionality
func SearchExpensesHandler(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	filteredExpenses := []models.Expense{}

	expenses, err := GetExpenses(db)
	if err != nil {
		http.Error(w, "Failed to retrieve expenses", http.StatusInternalServerError)
		return
	}
	// Filter expenses by query
	for _, expense := range expenses {
		if strings.Contains(expense.Description, query) {
			filteredExpenses = append(filteredExpenses, expense)
		}
	}

	// Encode the filtered expenses to JSON and write to expense
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(filteredExpenses)
}

// BulkAddExpensesHandler handles bulk addition of expenses
func BulkAddExpensesHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var expenses []models.Expense
		if err := json.NewDecoder(r.Body).Decode(&expenses); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		for _, expense := range expenses {
			if err := AddExpenses(db, expense); err != nil {
				http.Error(w, "Failed to add expense", http.StatusInternalServerError)
				return
			}
		}
		w.WriteHeader(http.StatusCreated)
		json.NewDecoder(r.Body).Decode(&expenses)

		// Save expenses in DB
		json.NewEncoder(w).Encode(expenses)
	}
}

// GenerateReport generates an expense report for a given time period.
func GenerateReport(db *sql.DB, startDate, endDate string) ([]models.Expense, error) {
	query := `SELECT Id, amount, category, description, date
              FROM expenses
              WHERE date BETWEEN ? AND ?
              ORDER BY date DESC`

	rows, err := db.Query(query, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var expenses []models.Expense
	for rows.Next() {
		var expense models.Expense
		if err := rows.Scan(&expense.Id, &expense.Amount, &expense.Category, &expense.Description, &expense.Date); err != nil {
			return nil, err
		}
		expenses = append(expenses, expense)
	}
	return expenses, nil
}

// Validation rules
func validateExpense(expense models.Expense) error {
	if expense.Amount <= 0 {
		return errors.New("amount must be greater than zero")
	}
	if expense.Category == "" {
		return errors.New("category is required")
	}
	if expense.Date == "" {
		return errors.New("date is required")
	}
	return nil
}

// AddExpenseHandler handles the addition of a new expense
func AddExpenseHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var expense models.Expense
		if err := json.NewDecoder(r.Body).Decode(&expense); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		// Validate the expense
		if err := validateExpense(expense); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Insert the expense into the database
		query := `INSERT INTO expenses (amount, category, description, date, currency_code) VALUES (?, ?, ?, ?, ?)`
		_, err := db.Exec(query, expense.Amount, expense.Category, expense.Description, expense.Date, expense.CurrencyCode)
		if err != nil {
			http.Error(w, "Error adding expense", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}

// func AddExpenses(db *sql.DB, expense models.Expense) error {
// 	if err := validateExpense(expense); err != nil {
// 		return err
// 	}
// 	query := `INSERT INTO expenses (amount, category, description, date, currency_code) VALUES (?, ?, ?, ?, ?)`
// 	_, err := db.Exec(query, expense.Amount, expense.Category, expense.Description, expense.Date, expense.CurrencyCode)
// 	return err
// }
