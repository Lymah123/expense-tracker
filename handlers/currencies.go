package handlers

import (
	"database/sql"
	"encoding/json"
	"expense-tracker/models"
	"expense-tracker/utils"
	"net/http"
	"log"
)

// CurrencyPageData represents the data needed for the currencies page
type CurrencyPageData struct {
	Currencies []models.Currency
	Message    string
}

// ListCurrenciesHandler handles displaying the currencies page
func ListCurrenciesHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		currencies, err := models.GetCurrencies(db)
		if err != nil {
			http.Error(w, "Error fetching currencies", http.StatusInternalServerError)
			return
		}

		data := CurrencyPageData{
			Currencies: currencies,
		}

		utils.RenderTemplate(w, "currencies", data)
	}
}

// AddCurrencyHandler handles adding a new currency
func AddCurrencyHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		code := r.FormValue("code")
		symbol := r.FormValue("symbol")
		name := r.FormValue("name")

		if code == "" || symbol == "" {
			http.Error(w, "Currency code and symbol are required", http.StatusBadRequest)
			return
		}

		err := models.AddCurrency(db, code, symbol, name)
		if err != nil {
			http.Error(w, "Error adding currency", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/currencies", http.StatusSeeOther)
	}
}

// UpdateCurrencyHandler handles updating a currency's symbol
func UpdateCurrencyHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		code := r.FormValue("code")
		symbol := r.FormValue("symbol")



		if code == "" || symbol == "" {
			http.Error(w, "Currency code and symbol are required", http.StatusBadRequest)
			return
		}

		err := models.UpdateCurrency(db, code, symbol)
		if err != nil {
			http.Error(w, "Error updating currency", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/currencies", http.StatusSeeOther)
	}
}

// DeleteCurrencyHandler handles deleting a currency
func DeleteCurrencyHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		code := r.FormValue("code")
		if code == "" {
			http.Error(w, "Currency code is required", http.StatusBadRequest)
			return
		}

		err := models.DeleteCurrency(db, code)
		if err != nil {
			http.Error(w, "Error deleting currency", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/currencies", http.StatusSeeOther)
	}
}

// GetCurrencyHandler handles retrieving a single currency (API endpoint)
func GetCurrencyHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "Currency code is required", http.StatusBadRequest)
			return
		}

		currency, err := models.GetCurrencyByCode(db, code)
		if err != nil {
			if err == sql.ErrNoRows {
				log.Printf("Error fetching currencies: %v", err)
				http.Error(w, "Currency not found", http.StatusNotFound)
				return
			}

			http.Error(w, "Error fetching currency", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(currency)
	}
}
