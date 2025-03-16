package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// API Response stores exchange rates
type APIResponse struct {
	Rates map[string]float64 `json:"rates"`
	Base  string             `json:"base"`
}

// ConvertCurrency converts one currency to another
func ConvertCurrency(amount float64, from string, to string) (float64, error) {
	apiURL := fmt.Sprintf("https://api.exchangerate-api.com/v4/latest/%s", from)
	resp, err := http.Get(apiURL)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var data APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, err
	}

	rate, exists := data.Rates[to]
	if !exists {
		return 0, fmt.Errorf("conversation rate found")
	}
	return amount * rate, nil
}
