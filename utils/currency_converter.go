package utils

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "time"
)

// API Response stores exchange rates
type APIResponse struct {
    Rates map[string]float64 `json:"rates"`
    Base  string             `json:"base"`
}

// CachedRates stores the latest exchange rates to avoid repeated API calls
var CachedRates map[string]float64

// UpdateExchangeRates fetches latest exchange rates every 24 hours
func UpdateExchangeRates() {
    // Initialize the cache
    CachedRates = make(map[string]float64)

    // Immediately fetch rates once
    fetchRates()

    // Then schedule regular updates
    ticker := time.NewTicker(24 * time.Hour)
    for range ticker.C {
        fetchRates()
    }
}

// fetchRates gets the latest exchange rates using USD as base
func fetchRates() {
    log.Println("Updating exchange rates...")
    apiURL := "https://api.exchangerate-api.com/v4/latest/USD"

    resp, err := http.Get(apiURL)
    if err != nil {
        log.Printf("Error fetching exchange rates: %v", err)
        return
    }
    defer resp.Body.Close()

    var data APIResponse
    if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
        log.Printf("Error decoding exchange rates: %v", err)
        return
    }

    // Cache the rates
    CachedRates = data.Rates
    log.Println("Exchange rates updated successfully")
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
        return 0, fmt.Errorf("conversion rate not found")
    }
    return amount * rate, nil
}
