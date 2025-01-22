package main

import (
	handlers "expense-tracker/handlers"
	"expense-tracker/models"
	"expense-tracker/reports"
	"expense-tracker/storage"
	"fmt"
	"log"
)

func main() {
	db, err := storage.InitDB("data/expenses.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// CLI Interface
	fmt.Println("Expense Tracker")
	for {
		fmt.Println("\nChoose an option:")
		fmt.Println("1. Add expense")
		fmt.Println("2. View expense")
		fmt.Println("3. Generate monthly report")
		fmt.Println("4. Exit")

		var choice int
		fmt.Scan(&choice)

		switch choice {
		case 1:
			var exp models.Expense
			fmt.Print("Amount: ")
			fmt.Scan(&exp.Amount)
			fmt.Print("Category: ")
			fmt.Scan(&exp.Category)
			fmt.Print("Description: ")
			fmt.Scan(&exp.Description)
			fmt.Print("Date(YYYY-MM-DD): ")
			fmt.Scan(&exp.Date)

			if err := handlers.AddExpenses(db, exp); err != nil {
				log.Println("Error adding expense:", err)
			} else {
				fmt.Println("Expense added successfully!")
			}

		case 2:
			expenses, err := handlers.GetExpenses(db)
			if err != nil {
				log.Println("Error retrieving expenses:", err)
			} else {
				for _, e := range expenses {
					fmt.Printf("%d: %.2f %s (%s) - %s\n", e.Id, e.Amount, e.Category, e.Description, e.Date)
				}
			}

		case 3:
			var month string
			fmt.Print("Enter month (YYYY-MM): ")
			fmt.Scan(&month)

			if err := reports.GenerateMonthlyReport(db, month); err != nil {
				log.Println("Error generating report:", err)
			}

		case 4:
			fmt.Println("Goodbye!")
			return

		default:
			fmt.Println("Invalid option")
		}
	}
}
