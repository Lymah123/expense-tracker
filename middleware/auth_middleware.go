package middleware

import (
    "context"
    "expense-tracker/firebase"
    "net/http"
    "strings"
		"fmt"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		if firebase.AuthClient == nil {
			fmt.Println("Firebase AuthClient is nil!")
			http.Error(w, "Firebase authentication is unavailable", http.StatusInternalServerError)
			return
		}

		token, err := firebase.AuthClient.VerifyIDToken(context.Background(), tokenString)
		if err != nil {
			http.Error(w, `{"error": "Invalid token"}`, http.StatusUnauthorized)
			return
		}

		fmt.Println("Authenticated user:", token.UID)
		next.ServeHTTP(w, r)
	})
}
