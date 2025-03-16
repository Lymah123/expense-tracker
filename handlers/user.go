package handlers

import (
	"context"
	"encoding/json"
	"expense-tracker/firebase"
	"expense-tracker/models"
	"firebase.google.com/go/v4/auth"
	"net/http"
)

func RegisterUser(w http.ResponseWriter, r *http.Request) {
	var user models.User

	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if firebase.AuthClient == nil {
		http.Error(w, "Firebase not initialized", http.StatusInternalServerError)
		return
	}

	params := (&auth.UserToCreate{}).
		Email(user.Email).
		Password(user.Password)

		createdUser, err := firebase.AuthClient.CreateUser(context.Background(), params)
		if err != nil {
			http.Error(w, "Error creating user:"+err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(map[string]string{"uid": createdUser.UID})

}

func LoginUser(w http.ResponseWriter, r *http.Request) {
	var user models.User

	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if firebase.AuthClient == nil {
		http.Error(w, "Firebase not initialized", http.StatusInternalServerError)
		return
	}

// Fetch user from Firebase
firebaseUser, err := firebase.AuthClient.GetUserByEmail(context.Background(), user.Email)
if err != nil {
	http.Error(w, "Invalid credentials", http.StatusUnauthorized)
	return
}

// Check if email is verified
if !firebaseUser.EmailVerified {
	http.Error(w, "Email not verified", http.StatusUnauthorized)
	return
}

// Generate a custom token for authentication
token, err := firebase.AuthClient.CustomToken(context.Background(), firebaseUser.UID)
if err != nil {
	http.Error(w, "Error generating token", http.StatusInternalServerError)
	return
}

json.NewEncoder(w).Encode(map[string]string{"token": token})
}
