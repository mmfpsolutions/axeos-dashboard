package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/scottwalter/axeos-dashboard/internal/auth"
	"github.com/scottwalter/axeos-dashboard/internal/config"
	"github.com/scottwalter/axeos-dashboard/internal/logger"
)

type contextKey string

const UserContextKey contextKey = "user"

// User represents authenticated user information
type User struct {
	Username string
}

// AuthMiddleware creates a middleware that checks JWT authentication
func AuthMiddleware(cfgManager *config.Manager, requireJWT bool) func(http.Handler) http.Handler {
	log := logger.New(logger.ModuleAuth)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cfg := cfgManager.GetConfig() // Get fresh config for hot reload
			// Skip JWT check if authentication is disabled or not required
			if !requireJWT || cfg.DisableAuthentication {
				next.ServeHTTP(w, r)
				return
			}

			// Get token from cookie
			cookie, err := r.Cookie("sessionToken")
			if err != nil {
				// No token found, redirect to login
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}

			// Verify token
			jwtService := auth.GetJWTService()
			claims, err := jwtService.VerifyToken(cookie.Value)
			if err != nil {
				// Token is invalid or expired, redirect to login and clear cookie
				log.WarnWithRequest(r, "JWT verification failed, redirecting to login: %v", err)
				http.SetCookie(w, &http.Cookie{
					Name:     "sessionToken",
					Value:    "",
					Path:     "/",
					HttpOnly: true,
					MaxAge:   -1,
				})
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}

			// Add user to context
			user := &User{Username: claims.Username}
			ctx := context.WithValue(r.Context(), UserContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserFromContext retrieves the user from the request context
func GetUserFromContext(r *http.Request) *User {
	user, ok := r.Context().Value(UserContextKey).(*User)
	if !ok {
		return nil
	}
	return user
}

// LoggingMiddleware logs each request
func LoggingMiddleware(next http.Handler) http.Handler {
	log := logger.New(logger.ModuleMiddleware)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip logging for health check endpoint to avoid log clutter
		if strings.Contains(r.URL.Path, "health.html") {
			next.ServeHTTP(w, r)
			return
		}

		// Log the request with client IP
		log.InfoWithRequest(r, "Request: %s %s", r.Method, r.URL.String())

		next.ServeHTTP(w, r)
	})
}
