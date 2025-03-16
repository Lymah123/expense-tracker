package handlers

import (
	"database/sql"
	"expense-tracker/models"
	"expense-tracker/utils"
	"log"
	"net/http"
	"strconv"
)

// Add budget handler
func AddBudgetHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			category := r.FormValue("category")
			amount, _ := strconv.ParseFloat(r.FormValue("amount"), 64)

			err := models.SetBudget(db, category, amount)
			if err != nil {
				http.Error(w, "Error setting budget", http.StatusInternalServerError)
				return
			}

			alert, err := utils.CheckBudgetAlert(db, category)
			if err != nil {
				http.Error(w, "Error checking budget alert", http.StatusInternalServerError)
				return
			}
			if alert != "" {
				log.Println(alert)
			}

			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		utils.RenderTemplate(w, "add_budget", nil)
	}
}

// GetCategoryBudgets retrieves akk category budget from database
func GetCategoryBudgets(db *sql.DB) ([]models.Budget, error) {
	query := `SELECT category, budget_amount FROM category_budgets`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var budgets []models.Budget
	for rows.Next() {
		var budget models.Budget
		if err := rows.Scan(&budget.Category, &budget.BudgetAmount); err != nil {
			return nil, err
		}
		budgets = append(budgets, budget)
	}
	return budgets, nil
}
