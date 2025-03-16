package handlers

import (
	"database/sql"
	"expense-tracker/models"
)

// Fetch all available currencies from the database
func GetCurrencies(db *sql.DB) ([]models.Currency, error) {
	rows, err := db.Query("SELECT code, symbol FROM currencies")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var currencies []models.Currency
	for rows.Next() {
		var c models.Currency
		if err := rows.Scan(&c.Code, &c.Symbol); err != nil {
			return nil, err
		}
		currencies = append(currencies, c)
	}
	return currencies, nil
}
