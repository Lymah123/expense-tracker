package handlers

import (
	"database/sql"
	"encoding/json"
	"expense-tracker/models"
	"expense-tracker/utils"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type RecurringExpensePageData struct {
	Expenses   []models.RecurringExpense
	Currencies []models.Currency
}

type AddRecurringExpensePageData struct {
	Currencies []models.Currency
}

func ListRecurringExpensesHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		expenses, err := models.GetRecurringExpenses(db)
		if err != nil {
			http.Error(w, "Error fetching recurring expenses", http.StatusInternalServerError)
			return
		}

		currencies, err := models.GetCurrencies(db)
		if err != nil {
			http.Error(w, "Error fetching currencies", http.StatusInternalServerError)
			return
		}

		data := RecurringExpensePageData{
			Expenses:   expenses,
			Currencies: currencies,
		}

		utils.RenderTemplate(w, "recurring", data)
	}
}

// AddRecurringExpenseHandler handles adding a new recurring expense
func AddRecurringExpenseHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			currencies, err := models.GetCurrencies(db)
			if err != nil {
				http.Error(w, "Error fetching currencies", http.StatusInternalServerError)
				return
			}
			data := AddRecurringExpensePageData{
				Currencies: currencies,
			}
			utils.RenderTemplate(w, "add_recurring", data)
			return
		}

		if r.Method == http.MethodPost {
			amount := utils.ParseFloat(r.FormValue("amount"))
			category := r.FormValue("category")
			description := r.FormValue("description")
			frequency := r.FormValue("frequency")
			nextDueDate := r.FormValue("next_due_date")
			currencyCode := r.FormValue("currency_code")

			// Get currency symbol from the database
			var currencySymbol string
			err := db.QueryRow("SELECT symbol FROM currencies WHERE code = ?", currencyCode).Scan(&currencySymbol)
			if err != nil {
				http.Error(w, "Error fetching currency symbol", http.StatusInternalServerError)
				return
			}

			err = models.AddRecurringExpense(db, amount, category, description, frequency, nextDueDate, currencyCode, currencySymbol)
			if err != nil {
				http.Error(w, "Error adding recurring expense", http.StatusInternalServerError)
				return
			}

			http.Redirect(w, r, "/recurring", http.StatusSeeOther)
			return
		}
	}
}

// ToggleRecurringExpenseHandler handles toggling a recurring expense active/inactive
func ToggleRecurringExpenseHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		vars := mux.Vars(r)
		id, err := strconv.Atoi(vars["id"])
		if err != nil {
			http.Error(w, "Invalid expense ID", http.StatusBadRequest)
			return
		}

		// Get current expense
		var expense models.RecurringExpense
		err = db.QueryRow("SELECT Id, is_active FROM recurring_expenses WHERE Id = ?", id).Scan(&expense.Id, &expense.IsActive)
		if err != nil {
			http.Error(w, "Error fetching expense", http.StatusInternalServerError)
			return
		}

		// Toggle is_active
		expense.IsActive = !expense.IsActive
		err = models.UpdateRecurringExpense(db, expense)
		if err != nil {
			http.Error(w, "Error updating expense", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/recurring", http.StatusSeeOther)
	}
}

// DeleteRecurringExpenseHandler handles deleting a recurring expense
func DeleteRecurringExpenseHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		vars := mux.Vars(r)
		id, err := strconv.Atoi(vars["id"])
		if err != nil {
			http.Error(w, "Invalid expense ID", http.StatusBadRequest)
			return
		}

		_, err = db.Exec("DELETE FROM recurring_expenses WHERE Id = ?", id)
		if err != nil {
			http.Error(w, "Error deleting expense", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/recurring", http.StatusSeeOther)
	}
}

// GetRecurringExpenseHandler handles getting a single recurring expense (for API)
func GetRecurringExpenseHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, err := strconv.Atoi(vars["id"])
		if err != nil {
			http.Error(w, "Invalid expense ID", http.StatusBadRequest)
			return
		}

		var expense models.RecurringExpense
		err = db.QueryRow(`
            SELECT Id, amount, category, description, frequency, next_due_date,
                   currency_code, currency_symbol, is_active, created_at
            FROM recurring_expenses
            WHERE id = ?`, id).Scan(
			&expense.Id, &expense.Amount, &expense.Category, &expense.Description,
			&expense.Frequency, &expense.NextDueDate, &expense.CurrencyCode,
			&expense.CurrencySymbol, &expense.IsActive, &expense.CreatedAt)

		if err == sql.ErrNoRows {
			http.Error(w, "Expense not found", http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, "Error fetching expense", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expense)
	}
}
