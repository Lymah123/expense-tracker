package handlers

import (
	"database/sql"
	"expense-tracker/models"
	"expense-tracker/utils"
	"log"
	"net/http"
	"strings"
)

func ProfileHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userId := getUserIdFromSession(r)
		if userId == 0 {
			http.Redirect(w, r, "login", http.StatusSeeOther)
			return
		}

		var user models.User
		err := db.QueryRow(`SELECT id, name, email, created_at, profile_picture, default_currency, date_format, use_dark_theme, two_factor_enabled FROM users WHERE id = ?`, userId).Scan(
			&user.Id, &user.Name, &user.Email, &user.CreatedAt, &user.ProfilePicture, &user.DefaultCurrency, &user.DateFormat, &user.UseDarkTheme, &user.TwoFactorEnabled,
		)

		profilePicUrl := ""
		if user.ProfilePicture.Valid {
			profilePicUrl = user.ProfilePicture.String
		}

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
					user.Initials += string(nameParts[len(nameParts)-1][0])
   				}
			}
		}

		// Get available currencies for the user
		currencies, err := models.GetCurrencies(db)
		if err != nil {
			log.Printf("Error retrieving currencies: %v", err)
		}

		data := struct {
			User 		 models.User
			ProfilePicUrl string
			Currencies []models.Currency
			Message    string
		}{
			User:			 user,
			ProfilePicUrl: profilePicUrl,
			Currencies: currencies,
			Message:    r.URL.Query().Get("message"),
		}

		utils.RenderTemplate(w, "profile", data)
	}
}

// UpdatePersonalInfoHandler handles updating user personal information
func UpdatePersonalInfoHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		_, err := db.Exec(`UPDATE users SET name = ?, email = ?, profile_picture = ? WHERE id = ?`, name, email, profilePicture, userId,
	)

	if err != nil {
		log.Printf("Error updating profile: %v", err)
		http.Error(w, "Error updating profile", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/profile?message=Personal+information+updated+successfully", http.StatusSeeOther)
	}
}

// UpdatePreferencesHandler handles updating user preferences
func UpdatePreferencesHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		_, err := db.Exec(
			"UPDATE users SET default_currency = ?, date_format = ?, use_dark_theme = ? WHERE id = ?",
			defaultCurrency, dateFormat, darkTheme, userId,
		)

		if err != nil {
			log.Printf("Error updating preferences: %v", err)
			http.Error(w, "Error updating preferences", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/profile?message=Preferences+updated+succcessfully", http.StatusSeeOther)
	}
}

// SettingsHandler handles displaying the settings page
func SettingsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userId := getUserIdFromSession(r)
		if userId == 0 {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		var settings struct {
			EmailNotification   bool
			DarkMode         bool
			Language        string
		}

		// Get user settings from the database
		err := db.QueryRow(`SELECT email_notifications, use_dark_mode, language FROM user_settings WHERE user_id = ?`, userId).Scan(
			&settings.EmailNotification, &settings.DarkMode, &settings.Language,
		)

		if err != nil && err != sql.ErrNoRows {
			log.Printf("Error retrieving user settings: %v", err)
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

		utils.RenderTemplate(w, "settings", data)
	}
}

// UpdateSettingsHandler handles updating user settings
func UpdateSettingsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		userId := getUserIdFromSession(r)
		if userId == 0 {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		emailNotifcations := r.FormValue("email_notifications") == "on"
		darkMode := r.FormValue("dark_mode") == "on"
		language := r.FormValue("language")

		// Check if settings exist first
		var count int
		err := db.QueryRow(`SELECT COUNT(*) FROM user_settings WHERE user_id = ?`, userId).Scan(&count)
		if err != nil {
			log.Printf("Error checking settings: %v", err)
			http.Error(w, "Error updating settings", http.StatusInternalServerError)
			return
		}

		var query string
		if count > 0 {
			// Update existing settings
			query = "UPDATE user_settings SET email_notifications = ?, use_dark_theme = ?, language = ? WHERE user_id = ?"
		} else {
			// Insert new settings
			query = "INSERT INTO user_settings (email_notifications, use_dark_theme, language, user_id) VALUES (?, ?, ?, ?)"
		}

		_, err = db.Exec(query, emailNotifcations, darkMode, language, userId)
		if err != nil {
			log.Printf("Error saving settings: %v", err)
			http.Error(w, "Error updating settings", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/settings?message=Settings+updated+successfully", http.StatusSeeOther)
	}
}

// Helper function to get user Id from session
func getUserIdFromSession( _ *http.Request) int {
	// For now, returns a placeholder value
    // This should be replaced with your actual session handling code
    return 1
}
