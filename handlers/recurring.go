package handlers

import (
	"database/sql"
	"expense-tracker/models"
	"expense-tracker/utils"
	"net/http"
)

func AddRecurringExpenseHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			amount := utils.ParseFloat(r.FormValue("amount"))
			category := r.FormValue("category")
			description := r.FormValue("description")
			frequency := r.FormValue("frequency")
			nextDueDate := r.FormValue("next_due_date")
			currencyCode := r.FormValue("currency_code")
			currencySymbol := r.FormValue("currency_symbol")

			err := models.AddRecurringExpense(db, amount, category, description, frequency, nextDueDate, currencyCode, currencySymbol)
			if err != nil {
				http.Error(w, "Error adding recurring expense", http.StatusInternalServerError)
				return
			}

			http.Redirect(w, r, "/recurring", http.StatusSeeOther)
			return
		}
		utils.RenderTemplate(w, "add_recurring", nil)
	}
}
