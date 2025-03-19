package models

import (
	"database/sql"
	"time"
)

// GetCurrencies retrieves all available currencies from the database
func GetCurrencies(db *sql.DB) ([]Currency, error) {
	rows, err := db.Query("SELECT Id, code, symbol, name, created_at FROM currencies")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var currencies []Currency
	for rows.Next() {
		var currency Currency
		if err := rows.Scan(&currency.Id, &currency.Code, &currency.Symbol, &currency.Name, &currency.CreatedAt); err != nil {
			return nil, err
		}
		currencies = append(currencies, currency)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// If no currencies found, add default USD
	if len(currencies) == 0 {
		now := time.Now()
		currencies = append(currencies, Currency{
			Id:        0,
			Code:      "USD",
			Symbol:    "$",
			Name:      "US Dollar",
			CreatedAt: now,
		})
	}

	return currencies, nil
}

type Currency struct {
	Id int `json:"id"`
	Code string `json:"code"`
	Symbol string `json:"symbol"`
	Name string `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}
// CreateCurrenciesTable creates the currencies table if it doesn't exist
func CreateCurrenciesTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS currencies (
		Id INTEGER PRIMARY KEY AUTOINCREMENT,
		code TEXT UNIQUE NOT NULL,
		symbol TEXT NOT NULL,
		name TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`

	_, err := db.Exec(query)
	if err != nil {
		return err
	}

	// Insert default currencies if table is empty
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM currencies").Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		defaultCurrencies := []struct {
			code string
			symbol string
			name string
		}{
			{"USD", "$", "US Dollar"},
			{"EUR", "€", "Euro"},
			{"GBP", "£", "British Pound"},
			{"JPY", "¥", "Japanese Yen"},
			{"NGN", "₦", "Nigerian Naira"},
		}

		for _, c := range defaultCurrencies {
			_, err := db.Exec("INSERT INTO currencies (code, symbol, name) VALUES (?, ?, ?)",
				c.code, c.symbol, c.name)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// AddCurrency adds a new currency to the database
func AddCurrency(db *sql.DB, code, symbol, name string) error {
	_, err := db.Exec("INSERT INTO currencies (code, symbol, name) VALUES (?, ?, ?)", code, symbol, name)
	return err
}

// UpdateCurrency updates an existing currency's symbol
func UpdateCurrency(db *sql.DB, code, symbol string) error {
	_, err := db.Exec("UPDATE currencies SET symbol = ? WHERE code = ?", symbol, code)
	return err
}

// DeleteCurrency deletes a currency from the database
func DeleteCurrency(db *sql.DB, code string) error {
	_, err := db.Exec("DELETE FROM currencies WHERE code = ?", code)
	return err
}

// GetCurrencyByCode retrieves a specific currency by its code
func GetCurrencyByCode(db *sql.DB, code string) (*Currency, error) {
	var currency Currency
	err := db.QueryRow("SELECT Id, code, symbol, name, created_at FROM currencies WHERE code = ?", code).
		Scan(&currency.Id, &currency.Code, &currency.Symbol, &currency.Name, &currency.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &currency, nil
}
