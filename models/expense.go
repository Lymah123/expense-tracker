package models

type Expense struct {
	Id int `json:"id"`
	Amount float64 `json:"amount"`
	Category string `json:"category"`
	Description string `json:"description"`
	Date string `json:"date"`
}

// These JSON tags are crucial for ensuring that the struct fields are correctly mapped to their respetive keys when the struct is converted to or from JSON, making it easier to work with JSON data.
