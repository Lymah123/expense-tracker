package models

type Expense struct {
	Id int `json:"id"`
	Amount float64 `json:"amount"`
	Category string `json:"category"`
	Description string `json:"description"`
	Date string `json:"date"`
}

type ExpenseDistribution struct {
	Category string `json:"category"`
	Total float64 `json:"total"`
}

type DashboardData struct {
	TotalExpenses float64 		`json:"totalExpenses"`
	BudgetRemaining float64		`json:"budgetRemaining"`
	Distribution []ExpenseDistribution `json:"distribution"`
	RecentTransactions []Expense `json:"recentTransactions"`
}
// These JSON tags are crucial for ensuring that the struct fields are correctly mapped to their respetive keys when the struct is converted to or from JSON, making it easier to work with JSON data.
