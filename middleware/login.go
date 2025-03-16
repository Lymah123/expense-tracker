package middleware

import (
    "bytes"
    "fmt"
    "context"
    "encoding/json"
    "expense-tracker/firebase"
    "firebase.google.com/go/v4/auth"
    "io"
    "net/http"
    "os"
    "sync"
)

var (
    firebaseAPIKey string
    initialized    bool
    initMutex      sync.Mutex
)

// InitAuthMiddleware initializes the authentication middleware
func InitAuthMiddleware() error {
    initMutex.Lock()
    defer initMutex.Unlock()

    if initialized {
        return nil
    }

    firebaseAPIKey = os.Getenv("FIREBASE_API_KEY")
    if firebaseAPIKey == "" {
        return fmt.Errorf("FIREBASE_API_KEY environment variable is missing")
    }

    initialized = true
    return nil
}

type FirebaseLoginResponse struct {
    IDToken string `json:"idToken"`
}

// LoginHandler: Authenticates a user using Firebase REST API
func LoginHandler(w http.ResponseWriter, r *http.Request) {
    // Check if middleware is initialized
    if !initialized {
        http.Error(w, `{"error": "Auth middleware not initialized"}`, http.StatusInternalServerError)
        return
    }

    var creds struct {
        Email    string `json:"email"`
        Password string `json:"password"`
    }

    if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
        http.Error(w, `{"error": "Invalid request payload"}`, http.StatusBadRequest)
        return
    }

    requestBody, err := json.Marshal(map[string]string{
        "email":             creds.Email,
        "password":          creds.Password,
        "returnSecureToken": "true",
    })
    if err != nil {
        http.Error(w, `{"error": "Failed to encode request"}`, http.StatusInternalServerError)
        return
    }

    firebaseURL := "https://identitytoolkit.googleapis.com/v1/accounts:signInWithPassword?key=" + firebaseAPIKey
    resp, err := http.Post(firebaseURL, "application/json", bytes.NewBuffer(requestBody))
    if err != nil {
        http.Error(w, `{"error": "Failed to reach Firebase"}`, http.StatusInternalServerError)
        return
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        http.Error(w, `{"error": "Invalid email or password"}`, http.StatusUnauthorized)
        return
    }

    var firebaseResp FirebaseLoginResponse
    if err := json.NewDecoder(resp.Body).Decode(&firebaseResp); err != nil {
        http.Error(w, `{"error": "Failed to parse Firebase response"}`, http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"token": firebaseResp.IDToken})
}

// RegisterHandler: Creates a user in Firebase and logs them in automatically
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	// Check if middleware is initialized
	if !initialized {
			http.Error(w, `{"error": "Auth middleware not initialized"}`, http.StatusInternalServerError)
			return
	}

	var creds struct {
			Email    string `json:"email"`
			Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
			http.Error(w, `{"error": "Invalid request payload"}`, http.StatusBadRequest)
			return
	}

	// Ensure Firebase AuthClient is initialized before using it
	if firebase.AuthClient == nil {
			http.Error(w, `{"error": "Firebase authentication client is not initialized"}`, http.StatusInternalServerError)
			return
	}

	user, err := firebase.AuthClient.CreateUser(context.Background(), (&auth.UserToCreate{}).
			Email(creds.Email).
			Password(creds.Password))

	if err != nil {
			// Log the specific Firebase error for debugging
			fmt.Printf("Firebase user creation error: %v\n", err)
			http.Error(w, `{"error": "Failed to create user"}`, http.StatusInternalServerError)
			return
	}

	// Auto-login after registration
	requestBody, err := json.Marshal(map[string]string{
			"email":             creds.Email,
			"password":          creds.Password,
			"returnSecureToken": "true",
	})
	if err != nil {
			http.Error(w, `{"error": "Failed to encode request"}`, http.StatusInternalServerError)
			return
	}

	firebaseURL := "https://identitytoolkit.googleapis.com/v1/accounts:signInWithPassword?key=" + firebaseAPIKey
	resp, err := http.Post(firebaseURL, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
			http.Error(w, `{"error": "Failed to reach Firebase"}`, http.StatusInternalServerError)
			return
	}
	defer resp.Body.Close()

	// Read the body once into a variable
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
			http.Error(w, `{"error": "Failed to read response body"}`, http.StatusInternalServerError)
			return
	}

	if resp.StatusCode != http.StatusOK {
			// Use the already read body
			fmt.Printf("Firebase login error: %s\n", string(bodyBytes))
			http.Error(w, fmt.Sprintf(`{"error": "Login failed with status %d"}`, resp.StatusCode), http.StatusInternalServerError)
			return
	}

	// Use the previously read body for decoding
	var firebaseResp FirebaseLoginResponse
	if err := json.Unmarshal(bodyBytes, &firebaseResp); err != nil {
			http.Error(w, `{"error": "Failed to parse Firebase response"}`, http.StatusInternalServerError)
			return
	}

	w.Header().Set("Content-Type", "application/json")
	// Only set the status code once, right before writing the response
	json.NewEncoder(w).Encode(map[string]string{
			"token": firebaseResp.IDToken,
			"uid":   user.UID,
	})
}
