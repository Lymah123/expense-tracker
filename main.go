package main

import (
	"database/sql"
	"encoding/json"
	"expense-tracker/apihandlers"
	"expense-tracker/firebase"
	"expense-tracker/handlers"
	"expense-tracker/middleware"
	"expense-tracker/models"
	"expense-tracker/reports"
	"expense-tracker/storage"
	"expense-tracker/utils"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "modernc.org/sqlite"
	"github.com/pquerna/otp/totp"
)

var (
	db               *sql.DB
	templates        map[string]*template.Template
	connectionString string
)

func init() {
	// Database path configuration
	dbPath := filepath.Join("data", "expenses.db")
	fmt.Println("Database path:", dbPath)
	connectionString = dbPath

	// Create data directory if it doesn't exist
	if err := os.MkdirAll("data", 0755); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}

	templates = make(map[string]*template.Template)
	templateFiles := []string{"dashboard", "home", "add", "view", "report", "currencies", "receipts", "footer", "login", "register"}
	for _, tmpl := range templateFiles {
		t, err := template.ParseFiles("templates/"+tmpl+".html", "templates/footer.html")
		if err != nil {
			log.Fatalf("Failed to parse template %s: %v", tmpl, err)
		}
		templates[tmpl] = t
	}
}

func main() {

	// Initialize database connection
	var err error
	db, err = sql.Open("sqlite", connectionString)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	defer db.Close()

	// // Initialize database tables after connection is established
	// if err := handlers.CreateReceiptsTable(db); err != nil {
	// 	log.Fatalf("Failed to create receipts table: %v", err)
	// }

	// Test the database connection
	if err = db.Ping(); err != nil {
		log.Fatalf("Error connecting to the database: %v", err)
	}

	// Load environment variables from .env file
	fmt.Println("Attempting to load .env file...")
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Err loading .env file: %v", err)
	}
	fmt.Println(".env file loaded successfully!")

	// Test the database connection
	if err = db.Ping(); err != nil {
		log.Fatalf("Error connecting to the database: %v", err)
	}
	rows, err := db.Query("SELECT 1")
	if err != nil {
		log.Fatalf("Cannot execute simple query: %v", err)
	}
	rows.Close()
	fmt.Println("Database connection verified with test query ")

	// Print environment variables status
	apiKey := os.Getenv("FIREBASE_API_KEY")
	serviceAccountKey := os.Getenv("FIREBASE_SERVICE_ACCOUNT_KEY")

	if apiKey != "" {
		fmt.Println("FIREBASE_API_KEY loaded successfully")
	}
	if serviceAccountKey != "" {
		fmt.Println("FIREBASE_SERVICE_ACCOUNT_KEY loaded successfully")
	}

	// Create necessary tables
	 if err := models.CreateCurrenciesTable(db); err != nil {
		log.Fatalf("Failed to create currencies table: %v", err)
	}
	if err := createBudgetTable(db); err != nil {
		log.Fatalf("Failed to create budget table: %v", err)
	}
	if err := models.CreateExpensesTable(db); err != nil {
		log.Fatalf("Failed to create expenses table: %v", err)
	}
	if err := models.CreateRecurringExpensesTable(db); err != nil {
		log.Fatalf("Failed to create recurring expenses table: %v", err)
	}
	if err := handlers.CreateReceiptsTable(db); err != nil {
		log.Fatalf("Failed to create receipts table: %v", err)
	}

	go utils.UpdateExchangeRates()
	go utils.ProcessRecurringExpenses(db)

	// Initialize Firebase
	err = firebase.InitFirebase()
	if err != nil {
		log.Fatalf("Error initializing Firebase: %v", err)
	}

	// Initialize the auth middleware
	if err := middleware.InitAuthMiddleware(); err != nil {
		log.Fatalf("Failed to initialize auth middleware: %v", err)
	}

	// Call the backup function periodically
	go func() {
		for {
			time.Sleep(24 * time.Hour)
			err := storage.BackupDB("data/expenses.db")
			if err != nil {
				log.Println("Error creating backup:", err)
			}
		}
	}()

	r := mux.NewRouter()

	// Public routes
	r.HandleFunc("/api/login", middleware.LoginHandler).Methods("POST")
	r.HandleFunc("/api/register", middleware.RegisterHandler).Methods("POST")
	r.HandleFunc("/login", loginHandler).Methods("GET")
	r.HandleFunc("/register", registerHandler).Methods("GET")

	// Redirect root to login
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	})

	// Protected routes
	protected := r.PathPrefix("/api").Subrouter()
	protected.Use(middleware.AuthMiddleware)
	protected.HandleFunc("/expenses", apihandlers.GetExpensesAPI(db)).Methods("GET")
	// Add other protected routes here

	// Other routes
	r.HandleFunc("/dashboard", dashboardHandler)
	r.HandleFunc("/home", homeHandler)
	r.HandleFunc("/view", viewExpensesHandler)
	r.HandleFunc("/report", generateReportHandler)
	r.HandleFunc("/receipts", ReceiptsHandler)
	r.HandleFunc("/upload-receipts", storage.UploadReceipt)
	r.HandleFunc("/view-receipts", handlers.ViewReceipt)
	r.HandleFunc("/setup-2fa", setup2FAHandler)
	r.HandleFunc("/verify-2fa", verify2FAHandler)
	r.HandleFunc("/admin", requireRole("admin", adminHandler))
	// r.HandleFunc("/currencies", currenciesHandler)
	r.HandleFunc("/report/weekly", generateWeeklyReportHandler)
	r.HandleFunc("/report/yearly", generateYearlyReportHandler)

	r.HandleFunc("/add", addExpensePageHandler).Methods("GET")
	r.HandleFunc("/add", addExpenseHandler).Methods("POST")
	// File Upload Router
	r.HandleFunc("/upload", apihandlers.UploadFileHandler).Methods("POST")

	// Register API routes
	r.HandleFunc("/api/overview", apihandlers.OverviewHandler(db))
	r.HandleFunc("/api/recent-transactions", apihandlers.RecentTransactionsHandler(db))
	r.HandleFunc("/api/expense-distribution", apihandlers.ExpenseDistributionHandler(db))
	r.HandleFunc("/api/budget-tracking", budgetTrackingHandler)
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Currency conversion API route
	convertedAmount, _ := utils.ConvertCurrency(100, "USD", "EUR")
	fmt.Println("Converted amount:", convertedAmount)

	// Search and Bulk expenses handlers
	r.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		handlers.SearchExpensesHandler(db, w, r)
	})
	r.HandleFunc("/bulk-add", handlers.BulkAddExpensesHandler(db))

	// Recurring expense routes
	r.HandleFunc("/recurring", handlers.ListRecurringExpensesHandler(db))
	r.HandleFunc("/recurring/add", handlers.AddRecurringExpenseHandler(db))
	r.HandleFunc("/recurring{id}/toggle", handlers.ToggleRecurringExpenseHandler(db))
	r.HandleFunc("/recurring{id}/delete", handlers.DeleteRecurringExpenseHandler(db))
	r.HandleFunc("/recurring/{id}", handlers.GetRecurringExpenseHandler(db))

	// Currency routes
	r.HandleFunc("/currencies", handlers.ListCurrenciesHandler(db))
	r.HandleFunc("/currencies/add", handlers.AddCurrencyHandler(db))
	r.HandleFunc("/currencies/update", handlers.UpdateCurrencyHandler(db))
	r.HandleFunc("/currencies/delete", handlers.DeleteCurrencyHandler(db))
	r.HandleFunc("/api/currencies", handlers.GetCurrencyHandler(db))

	fmt.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func budgetTrackingHandler(w http.ResponseWriter, r *http.Request) {
	budgets, err := handlers.GetCategoryBudgets(db)
	if err != nil {
		http.Error(w, "Error retrieving budgets", http.StatusInternalServerError)
		return
	}

	var budgetTracking []struct {
		Category     string  `json:"category"`
		BudgetAmount float64 `json:"budgetAmount"`
		ActualAmount float64 `json:"actualAmount"`
	}

	for _, budget := range budgets {
		var actualAmount float64
		err := db.QueryRow("SELECT COALESCE(SUM(amount), 0) FROM expenses WHERE category = ?", budget.Category).Scan(&actualAmount)
		if err != nil {
			http.Error(w, "Error calculating actual amount", http.StatusInternalServerError)
			return
		}

		budgetTracking = append(budgetTracking, struct {
			Category     string  `json:"category"`
			BudgetAmount float64 `json:"budgetAmount"`
			ActualAmount float64 `json:"actualAmount"`
		}{
			Category:     budget.Category,
			BudgetAmount: budget.BudgetAmount,
			ActualAmount: actualAmount,
		})
	}
	json.NewEncoder(w).Encode(budgetTracking)
}

func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "dashboard", nil)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "home", nil)
}

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

func createBudgetTable(db *sql.DB) error {
	query := `
  CREATE TABLE IF NOT EXISTS budgets (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    category TEXT NOT NULL UNIQUE,
    amount REAL NOT NULL
    );`

	_, err := db.Exec(query)
	if err != nil {
		return err
	}

	return nil
}

func addExpenseHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		exp := models.Expense{
			Amount:       parseFloat(r.FormValue("amount")),
			Category:     r.FormValue("category"),
			Description:  r.FormValue("description"),
			Date:         r.FormValue("date"),
			CurrencyCode: r.FormValue("currency"),
		}

		if err := handlers.AddExpenses(db, exp); err != nil {
			http.Error(w, "Error adding expense", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
}

func viewExpensesHandler(w http.ResponseWriter, r *http.Request) {
	expenses, err := handlers.GetExpenses(db)
	if err != nil {
		http.Error(w, "Error retrieving expenses", http.StatusInternalServerError)
		return
	}
	renderTemplate(w, "view", expenses)
}

func ReceiptsHandler(w http.ResponseWriter, r *http.Request) {
	receipts, err := handlers.GetReceipts(db)
	if err != nil {
		log.Println("Error retrieving receipts:", err)
		http.Error(w, "Error retrieving receipts", http.StatusInternalServerError)
		return
	}

	// Data structure
	data := struct {
		Receipts []models.Receipt
	}{
		Receipts: receipts,
	}
	renderTemplate(w, "receipts", data)
}

func generateReportHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		reportType := r.FormValue("report_type")
		startDate := r.FormValue("start_date")
		endDate := r.FormValue("end_date")
		exportFormat := r.FormValue("format")

		log.Printf("Generating report for %s Report:\nStart Date: %s\nEnd Date: %s\nExport Format: %s\n", reportType, startDate, endDate, exportFormat)

		// Generate the report based on the form date
		err := reports.GenerateReport(db, reportType, startDate, endDate, exportFormat)
		if err != nil {
			log.Println("Eror generating repport:", err)
			http.Error(w, "Error generating report", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	}
	renderTemplate(w, "report", nil)
}

func generateWeeklyReportHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		startDate := r.FormValue("start_date")
		endDate := r.FormValue("end_date")
		exportFormat := r.FormValue("format")

		log.Printf("Getting weekly report from %s to %s in %s format", startDate, endDate, exportFormat)

		if err := reports.GenerateWeeklyReport(db, startDate, endDate, exportFormat); err != nil {
			log.Println("Error generating report:", err)
			http.Error(w, "Error generating report", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}
	renderTemplate(w, "report", nil)
}

func generateYearlyReportHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		startDate := r.FormValue("start_date")
		endDate := r.FormValue("end_date")
		exportFormat := r.FormValue("format")

		log.Printf("Generating yearly report from %s to %s in %s format", startDate, endDate, exportFormat)

		if err := reports.GenerateYearlyReport(db, startDate, endDate, exportFormat); err != nil {
			log.Println("Error generating report:", err)
			http.Error(w, "Error generating report", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}
	renderTemplate(w, "report", nil)
}

func addExpensePageHandler(w http.ResponseWriter, r *http.Request) {
	// Get currencies for the dropdown
	currencies, err := models.GetCurrencies(db)
	if err != nil {
		log.Printf("Currency error: %v", err)
		http.Error(w, "Error retrieving currencies", http.StatusInternalServerError); return
	}
	renderTemplate(w, "add", currencies)
}

// 2FA setup and verification endpoints
func setup2FAHandler(w http.ResponseWriter, r *http.Request) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "ExpenseTracker",
		AccountName: "user@example.com",
	})
	if err != nil {
		http.Error(w, "Error generating 2FA key", http.StatusInternalServerError)
	}
	// save key.Secrete() to the user's account in the database
	http.Redirect(w, r, "/verify-2fa?secret="+key.Secret(), http.StatusSeeOther)
}

func verify2FAHandler(w http.ResponseWriter, r *http.Request) {
	secret := r.URL.Query().Get("secret")
	code := r.FormValue("code")
	if totp.Validate(code, secret) {
		// Mark the user as verified in the database
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	} else {
		http.Error(w, "Invalid 2FA code", http.StatusUnauthorized)
	}
}

func parseFloat(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return f
}

// Admin handler
func adminHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "admin", nil)
}

// Middleware for role-based access control
func requireRole(role string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userId := 1 // Get user ID from session or content
		userRole, err := models.GetUserRole(db, userId)
		if err != nil || userRole != role {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		next(w, r)
	}
}

type ReportData struct {
	ReportType   string
	StartDate    string
	EndDate      string
	ExportFormat string
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "login", nil)
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "register", nil)
}
