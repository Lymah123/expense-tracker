package main

import (
    "database/sql"
    handlers "expense-tracker/handlers"
    "expense-tracker/models"
    "expense-tracker/reports"
    "expense-tracker/storage"
		"expense-tracker/apihandlers"
    "fmt"
    "html/template"
    "log"
    "net/http"
    "strconv"
)

func landingHandler(w http.ResponseWriter, r *http.Request) {
    renderTemplate(w, "landing", nil)
}

func dashboardHandler(w http.ResponseWriter, r *http.Request) {
    renderTemplate(w, "dashboard", nil)
}

var (
    db        *sql.DB
    templates map[string]*template.Template
)

func init() {
    templates = make(map[string]*template.Template)
    templateFiles := []string{"landing", "dashboard", "home", "add", "view", "report"}
    for _, tmpl := range templateFiles {
        t, err := template.ParseFiles("templates/" + tmpl + ".html")
        if err != nil {
            log.Fatalf("Failed to parse template %s: %v", tmpl, err)
        }
        templates[tmpl] = t
    }
}

func main() {
    var err error
    db, err = storage.InitDB("data/expenses.db")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

		// // Create the budget table if it doesn't exist
		// err = createBudgetTable(db)
		// if err != nil {
		// 	log.Fatal(err)
		// }

    // Create a new router instance
    mux := http.NewServeMux()

    // Register routes in the correct order
    mux.HandleFunc("/", landingHandler)
    mux.HandleFunc("/dashboard", dashboardHandler)
    mux.HandleFunc("/home", homeHandler)
    mux.HandleFunc("/add", addExpenseHandler)
    mux.HandleFunc("/view", viewExpensesHandler)
    mux.HandleFunc("/report", generateReportHandler)

		// Register API routes
		mux.HandleFunc("/api/overview", apihandlers.OverviewHandler(db))
		mux.HandleFunc("/api/recent-transactions", apihandlers.RecentTransactionsHandler(db))
		mux.HandleFunc("/api/expense-distribution", apihandlers.ExpenseDistributionHandler(db))

    fmt.Println("Server started on :8080")
    log.Fatal(http.ListenAndServe(":8080", mux))
}

// func createBudgetTable(db *sql.DB) error {
// 	// Create the budget table if it doesn't exist
// 	budgetTable := `
// 	CREATE TABLE IF NOT EXISTS budget (
// 	id INTEGER PRIMARY KEY AUTOINCREMENT,
// 	budget float
// 	);`
// 	_, err := db.Exec(budgetTable)
// 	if err != nil {
// 		return err
// }
// // Insert a default budget value if the table is empty
// _, err = db.Exec("INSERT INTO budget (budget) SELECT 5000 WHERE NOT EXISTS (SELECT 1 FROM budget);")
// if err != nil {
//     return err
// }
// return nil
// }

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
    t, ok := templates[tmpl]
    if !ok {
        http.Error(w, "Template not found", http.StatusInternalServerError)
        return
    }
    err := t.Execute(w, data)
    if err != nil {
        log.Printf("Template execution error: %v", err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
    }
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
    renderTemplate(w, "home", nil)
}

func addExpenseHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method == http.MethodPost {
        exp := models.Expense{
            Amount:      parseFloat(r.FormValue("amount")),
            Category:    r.FormValue("category"),
            Description: r.FormValue("description"),
            Date:        r.FormValue("date"),
        }

        if err := handlers.AddExpenses(db, exp); err != nil {
            http.Error(w, "Error adding expense", http.StatusInternalServerError)
            return
        }
        http.Redirect(w, r, "/", http.StatusSeeOther)
        return
    }
    renderTemplate(w, "add", nil)
}

func viewExpensesHandler(w http.ResponseWriter, r *http.Request) {
    expenses, err := handlers.GetExpenses(db)
    if err != nil {
        http.Error(w, "Error retrieving expenses", http.StatusInternalServerError)
        return
    }
    renderTemplate(w, "view", expenses)
}

func generateReportHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method == http.MethodPost {
        month := r.FormValue("month")
        if err := reports.GenerateMonthlyReport(db, month); err != nil {
            http.Error(w, "Error generating report", http.StatusInternalServerError)
            return
        }
        http.Redirect(w, r, "/", http.StatusSeeOther)
        return
    }
    renderTemplate(w, "report", nil)
}

func parseFloat(s string) float64 {
    f, err := strconv.ParseFloat(s, 64)
    if err != nil {
        return 0
    }
    return f
}
