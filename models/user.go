package models

import (
	"database/sql"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type User struct {
	Id       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password,omitempty"`
	Role     string `json:"role"` // e.g., "admin", "user"
	UID      string `json:"uid"`
	Name     string `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	PasswordHash string `json:"password_hash,omitempty"`
	ProfilePicture sql.NullString `json:"profile_picture"`
	DefaultCurrency string `json:"default_currency"`
	TwoFactorEnabled bool `json:"two_factor_enabled"`
	UseDarkTheme     bool `json:"use_dark_theme"`
	DateFormat       string `json:"date_format"`
	Initials         string `json:"initials"`
}

func GetUserRole(db *sql.DB, userId int) (string, error) {
	var role string
	err := db.QueryRow("SELECT role FROM users WHERE id = ?", userId).Scan(&role)
	return role, err
}

func RegisterUser(db *sql.DB, username, email, password, role string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = db.Exec("INSERT INTO users (username, email, password, role) VALUES (?, ?, ?, ?)", username, email, hashedPassword, role)
	return err
}

func LoginUser(db *sql.DB, email, password string) (User, error) {
	var user User
	err := db.QueryRow("SELECT id, username, email, password, role FROM users WHERE email = ?", email).Scan(&user.Id, &user.Username, &user.Email, &user.Password, &user.Role)
	if err != nil {
		if err == sql.ErrNoRows {
			return User{}, errors.New("user not found")
		}
		return user, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return user, errors.New("invalid password")
	}

	return user, nil
}

func CreateUsersTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS users (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	email TEXT NOT NULL UNIQUE,
	password_hash TEXT NOT NULL,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	profile_picture TEXT,
	default_currency TEXT DEFAULT 'USD',
	date_format TEXT DEFAULT 'YYYY-MM-DD',
	use_dark_theme BOOLEAN DEFAULT false,
	two_factor_enabled BOOLEAN DEFAULT false,
	)`

	_, err := db.Exec(query)
	return err
}

func CreateUserSettingsTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS user_settings (
	user_id INTEGER PRIMARY KEY,
	email_notifications BOOLEAN DEFAULT true,
	use_dark_theme BOOLEAN DEFAULT false,
	language TEXT DEFAULT 'en',
	FOREIGN KEY (user_id) REFERENCES users(id)
	)`
	_, err := db.Exec(query)
	return err
}
