package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"

	"github.com/scottwalter/axeos-dashboard/internal/auth"
	"github.com/scottwalter/axeos-dashboard/internal/config"
)

// LoginRequest represents the login request body
type LoginRequest struct {
	Username       string `json:"username"`
	HashedPassword string `json:"hashedPassword"`
}

// LoginResponse represents the login response
type LoginResponse struct {
	Message string `json:"message"`
}

// HandleLogin handles POST /api/login
func HandleLogin(configDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(map[string]string{"message": "Method Not Allowed"})
			return
		}

		// Read request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"message": "Invalid request body"})
			return
		}
		defer r.Body.Close()

		// Parse login request
		var loginReq LoginRequest
		if err := json.Unmarshal(body, &loginReq); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"message": "Invalid request body"})
			return
		}

		// Load access credentials
		accessData, err := auth.LoadAccessCredentials(configDir)
		if err != nil {
			fmt.Printf("Error reading access.json: %v\n", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"message": "Server configuration error."})
			return
		}

		// Verify credentials
		hashedPassword, exists := accessData[loginReq.Username]
		if !exists || hashedPassword != loginReq.HashedPassword {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"message": "Invalid username or password"})
			return
		}

		// Create JWT token
		jwtService := auth.GetJWTService()
		token, err := jwtService.CreateToken(loginReq.Username)
		if err != nil {
			fmt.Printf("Error creating JWT: %v\n", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"message": "Internal Server Error"})
			return
		}

		// Get cookie max age from config
		cfgManager := config.GetManager(filepath.Dir(configDir))
		cfg := cfgManager.GetConfig()
		maxAge := cfg.CookieMaxAge
		if maxAge == 0 {
			maxAge = 3600 // Default 1 hour
		}

		// Set JWT in HTTP-only cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "sessionToken",
			Value:    token,
			Path:     "/",
			HttpOnly: true,
			MaxAge:   maxAge,
			SameSite: http.SameSiteStrictMode,
		})

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(LoginResponse{Message: "Login successful"})
	}
}

// HandleLogout handles ANY /api/logout
func HandleLogout(w http.ResponseWriter, r *http.Request) {
	// Clear session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "sessionToken",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Logout successful"})
}
