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
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/pquerna/otp/totp"
	_ "modernc.org/sqlite"
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

	// Initialize templates map
	templates = make(map[string]*template.Template)

	// Fix: Combined declaration and assignment
	defaultTemplates, err := template.ParseFiles("templates/base.html", "templates/defaults.html", "templates/footer.html")
	if err != nil {
		log.Fatal("Failed to parse default templates:", err)
	}

	templateFiles := []string{"dashboard", "home", "add", "view", "report", "currencies", "receipts", "login", "register", "faq", "contact", "privacy", "terms", "profile", "settings", "setup_2fa", "verify_2fa", "2fa_already_enabled" }

	for _, tmpl := range templateFiles {

		t, err := defaultTemplates.Clone()
		if err != nil {
			log.Fatalf("Failed to clone template for %s: %v", tmpl, err)
		}

		// Reuse err variable
		t, err = t.ParseFiles("templates/" + tmpl + ".html")
		if err != nil {
			log.Fatalf("Failed to parse template %s: %v", tmpl, err)
		}
		templates[tmpl] = t
	}
}

func main() {
	utils.LoadTemplates()

	// Initialize database connection
	var err error
	db, err = sql.Open("sqlite", connectionString)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	defer db.Close()

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
	if err := createUserTables(db); err != nil {
		log.Fatalf("Failed to crate user tables: %v", err)
	}
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

	r.HandleFunc("/debug/templates", func(w http.ResponseWriter, r *http.Request){
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintln(w, "Templates loaded in main.go:")
		for name := range templates {
			fmt.Fprintf(w, "- %s\n", name)
		}

		fmt.Fprintln(w, "\nTemplates loaded by utils.LoadTemplates():")
		for _, name := range utils.GetTemplateNames() {
			fmt.Fprintf(w, "- %s\n", name)
		}
	})

	// Handlers for footer navigations
	r.HandleFunc("/faq", func(w http.ResponseWriter, r *http.Request) {
		renderTemplate(w, "faq", nil)
	})
	r.HandleFunc("/contact", func(w http.ResponseWriter, r *http.Request) {
		renderTemplate(w, "contact", nil)
	})
	r.HandleFunc("/privacy", func(w http.ResponseWriter, r *http.Request) {
		renderTemplate(w, "privacy", nil)
	})
	r.HandleFunc("/terms", func(w http.ResponseWriter, r *http.Request) {
		renderTemplate(w, "terms", nil)
	})

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

	// Other routes
	r.HandleFunc("/profile", profileHandler)
	r.HandleFunc("/settings", settingsHandler)
	r.HandleFunc("/logout", logoutHandler)
	r.HandleFunc("/profile/update-personal", updatePersonalInfoHandler)
	r.HandleFunc("/profile/update-preferences", updatePreferencesHandler)
	r.HandleFunc("/settings/update", updateSettingsHandler)
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

	r.HandleFunc("/recurring", handlers.ListRecurringExpensesHandler(db))
	r.HandleFunc("/recurring/add", handlers.AddRecurringExpenseHandler(db))
	r.HandleFunc("/recurring/{id}/toggle", handlers.ToggleRecurringExpenseHandler(db))
	r.HandleFunc("/recurring/{id}/delete", handlers.DeleteRecurringExpenseHandler(db))
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

	templateData := struct {
		PageName string
		Data     interface{}
	}{
		PageName: tmpl,
		Data:     data,
	}

	err := t.ExecuteTemplate(w, "base", templateData)
	if err != nil {
		log.Printf("Template execution error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	templateData.PageName = tmpl
	templateData.Data = data
}

func createUserTables(db *sql.DB) error {
    usersQuery := `
    CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL,
        email TEXT NOT NULL UNIQUE,
        password TEXT NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        profile_picture TEXT,
        default_currency TEXT DEFAULT 'USD',
        date_format TEXT DEFAULT 'YYYY-MM-DD',
        use_dark_theme BOOLEAN DEFAULT FALSE,
        two_factor_enabled BOOLEAN DEFAULT FALSE,
        two_factor_secret TEXT,
        role TEXT DEFAULT 'user'
    );`

    if _, err := db.Exec(usersQuery); err != nil {
        return err
    }

    settingsQuery := `
    CREATE TABLE IF NOT EXISTS user_settings (
        user_id INTEGER PRIMARY KEY,
        email_notifications BOOLEAN DEFAULT FALSE,
        use_dark_theme BOOLEAN DEFAULT FALSE,
        language TEXT DEFAULT 'en',
        FOREIGN KEY (user_id) REFERENCES users(id)
    );`

    _, err := db.Exec(settingsQuery)
		if err != nil {
			return err
		}

		var count int
		if err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count); err != nil {
			return err
		}

		if count == 0 {
			_, err := db.Exec("INSERT INTO users (name, email, password, role) VALUES (?, ?, ?, ?)", "Admin User", "@dmin@example.com", "password123", "admin")
			if err != nil {
				return err
			}

			_, err = db.Exec("INSERT INTO user_settings (user_id, email_notifications, use_dark_theme, language) VALUES (?, ?, ?, ?)", 1, false, false, "en")
			if err != nil {
				return err
			}
		}

		return nil
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
		http.Error(w, "Error retrieving currencies", http.StatusInternalServerError)
		return
	}
	renderTemplate(w, "add", currencies)
}

func profileHandler(w http.ResponseWriter, r *http.Request) {
	userId := getUserIdFromSession(r)
	if userId == 0 {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	var user models.User
	err := db.QueryRow(`SELECT id, name, email, created_at, profile_picture, default_currency, date_format, use_dark_theme, two_factor_enabled FROM users WHERE id = ?`, userId).Scan(&user.Id, &user.Name, &user.Email, &user.CreatedAt, &user.ProfilePicture, &user.DefaultCurrency, &user.DateFormat, &user.UseDarkTheme, &user.TwoFactorEnabled,)

	if err != nil {
		log.Printf("Error retrieving user: %v", err)
		http.Error(w, "Error retrieving user profile", http.StatusInternalServerError)
		return
	}

	// Get user's initials for profile picture placeholder
	if user.Name != "" {
		nameParts := strings.Fields(user.Name)
		if len(nameParts) > 0 {
			user.Initials = string(nameParts[0][0])
			if len(nameParts) > 1 {
				user.Initials += string(nameParts[len(nameParts) -1][0])
			}
		}
	}

	// Get available currencies
	currencies, err := models.GetCurrencies(db)
	if err != nil {
		log.Printf("Error retrieving currencies: %v", err)
	}

	data := struct {
		User models.User
		Currencies []models.Currency
		Message string
	}{
		User: user,
		Currencies: currencies,
		Message: r.URL.Query().Get("message"),
	}

	renderTemplate(w, "profile", data)
}

func getUserIdFromSession(_ *http.Request) int {
	// This function should retrieve the user Id from the session
	//For now, we'll just return a placeholder value
	return 1
}

func settingsHandler(w http.ResponseWriter, r *http.Request) {
	userId := getUserIdFromSession(r)
	if userId == 0 {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	var settings struct {
		EmailNotifications bool
		DarkMode bool
		Language string
	}

	// Get user settings from the database
	err := db.QueryRow(`SELECT email_notifications, use_dark_theme, language FROM user_settings WHERE user_id = ?`, userId).Scan(&settings.EmailNotifications, &settings.DarkMode, &settings.Language,
	)

	if err != nil && err != sql.ErrNoRows {
		log.Printf("Error retrieving settings: %v", err)
		http.Error(w, "Error retrieving user settings", http.StatusInternalServerError)
		return
	}

	data := struct {
		Settings interface{}
		Message string
	}{
	Settings: settings,
	Message: r.URL.Query().Get("message"),
	}

	renderTemplate(w, "settings", data)
}


func setup2FAHandler(w http.ResponseWriter, r *http.Request) {
	userId := getUserIdFromSession(r)
	if userId == 0 {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Check if user already has 2FA
	var has2FA bool
	err := db.QueryRow("SELECT two_factor_enabled FROM users WHERE id = ?", userId).Scan(&has2FA)
	if err != nil {
		log.Printf("Error checking 2FA status: %v", err)
		http.Error(w, "Error setting up 2FA", http.StatusInternalServerError)
		return
	}

	if has2FA {
		// User already has 2FA
		renderTemplate(w, "2fa_already_enabled", nil)
		return
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer: 		"ExpenseTracker",
		AccountName: getUserEmail(userId),
	})
	if err != nil {
		http.Error(w, "Error generating 2FA key", http.StatusInternalServerError)
		return
	}

	storeSecretInSession(w, r, key.Secret())

	data := struct {
		Secret string
		QRCode string
	}{
		Secret: key.Secret(),
		QRCode: key.URL(),
	}

	renderTemplate(w, "setup_2fa", data)
}

// Helper to get user email
func getUserEmail(userId int) string {
	var email string
	err := db.QueryRow("SELECT email FROM users WHERE id = ?", userId).Scan(&email)
	if err != nil {
		return "user@example.com"
	}
	return email
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:  "session_token",
		Value: "",
		Path:  "/",
		MaxAge: -1,
		HttpOnly: true,
	})

	// Clear any Firebase auth tokens (if using Firebase)
	// You might need additional cleanup based on the auth mechanism

	// Redirect to login page
	http.Redirect(w, r, "/login?message=Successfully+loggged+out", http.StatusSeeOther)
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
	data := map[string]interface{}{
		"PageName": "login",
	}
	renderTemplate(w, "login", data)
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"PageName": "register",
	}
	renderTemplate(w, "register", data)
}

func RecurringExpenseHandler(w http.ResponseWriter, r *http.Request) {
	recurringExpenses, err := models.GetRecurringExpenses(db)
	if err != nil {
		log.Printf("Error retrieving recurring expenses: %v", err)
		http.Error(w, "Error retrieving recurring expenses", http.StatusInternalServerError)
		return
	}

	data := struct {
		RecurringExpenses []models.RecurringExpense
	}{
		RecurringExpenses: recurringExpenses,
	}

	renderTemplate(w, "recurring", data)
}

// Parse float helper function
func parseFloat(s string) float64 {
	value, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0.0
	}
	return value
}

func storeSecretInSession(w http.ResponseWriter, _*http.Request, secret string) {
	http.SetCookie(w, &http.Cookie{
		Name:  "totp_secret",
		Value: secret,
		Path:	"/",
		HttpOnly: true,
		MaxAge:  3600,
	})
}

func verify2FAHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		renderTemplate(w, "verify_2fa", nil)
		return
	}

	code := r.FormValue("code")
	secret := r.URL.Query().Get("secret")

	valid := totp.Validate(code, secret)
	if !valid {
		renderTemplate(w, "verify_2fa", map[string]interface{}{
			"Error": "Invalid verification code",
			"secret": secret,
		})
		return
	}

	// Code is valid, enable 2FA for user
	userId := getUserIdFromSession(r)
	_, err := db.Exec("UPDATE users SET two_factor_enabled = 1, two_factor_secret = ? WHERE id = ?", secret, userId)
if err != nil {
	http.Error(w, "Error enabling 2FA", http.StatusInternalServerError)
	return
}

http.Redirect(w, r, "/profile?message=Two-factor+authentication+enabled", http.StatusSeeOther)
}

func updatePersonalInfoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userId := getUserIdFromSession(r)
	if userId == 0 {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	name := r.FormValue("name")
	email := r.FormValue("email")
	profilePicture := r.FormValue("profile_picture")

	_, err := db.Exec("UPDATE users SET name = ?, email = ?, profile_picture = ? WHERE id = ?", name, email, profilePicture, userId)
	if err != nil {
		http.Error(w, "Error updating profile", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/profile?message=Profile+updated+successfully", http.StatusSeeOther)
}

// UpdatePreferencesHandler
func updatePreferencesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userId := getUserIdFromSession(r)
	if userId == 0 {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	defaultCurrency := r.FormValue("default_currency")
	dateFormat := r.FormValue("date_format")
	darkTheme := r.FormValue("dark_theme") == "on"

	_, err := db.Exec("UPDATE users SET default_currency = ?, date_format = ?, use_dark_theme = ? WHERE id = ?", defaultCurrency, dateFormat, darkTheme, userId)
	if err != nil {
		http.Error(w, "Error updating preferences", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/profile?message=Preferences+updated+successfully", http.StatusSeeOther)
}

func updateSettingsHandler (w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userId := getUserIdFromSession(r)
	if userId == 0 {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	emailNotifications := r.FormValue("email_notifications") == "on"
	darkMode := r.FormValue("dark_mode") == "on"
	language := r.FormValue("language")

	// Check if settings exist
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM user_settings WHERE user_id = ?", userId).Scan(&count)
	if err != nil {
		http.Error(w, "Error checking user settings", http.StatusInternalServerError)
		return
	}

	var query string
	if count > 0 {
		query = "UPDATE user_settings SET email_notifications = ?, use_dark_theme = ?, language = ? WHERE user_id = ?"
	} else {
		query = "INSERT INTO user_settings (email_notifications, use_dark_theme, language, user_id) VALUES (?, ?, ?, ?)"
	}

	_, err = db.Exec(query, emailNotifications, darkMode, language, userId)
	if err != nil {
		http.Error(w, "Error updating user settings", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/settings?message=Settings+updated+successfully", http.StatusSeeOther)
}
