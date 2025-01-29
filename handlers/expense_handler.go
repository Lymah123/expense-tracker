package handlers

import (
    "database/sql"
    "expense-tracker/models"
)

// AddExpenses inserts a new expense into the database.
func AddExpenses(db *sql.DB, expense models.Expense) error {
    query := `INSERT INTO expenses (amount, category, description, date) VALUES (?, ?, ?, ?)`
    _, err := db.Exec(query, expense.Amount, expense.Category, expense.Description, expense.Date)
    return err
}

// GetExpenses retrieves all expenses from the database, ordered by date in descending order.
func GetExpenses(db *sql.DB) ([]models.Expense, error) {
    query := `SELECT id, amount, category, description, date FROM expenses ORDER BY date DESC`
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
    query := `SELECT id, amount, category, description, date
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

// GenerateReport generates an expense report for a given time period.
func GenerateReport(db *sql.DB, startDate, endDate string) ([]models.Expense, error) {
    query := `SELECT id, amount, category, description, date
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
