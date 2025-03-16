package models

import (
	"database/sql"
	"errors"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Id       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password,omitempty"`
	Role     string `json:"role"` // e.g., "admin", "user"
	UID      string `json:"uid"`
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
