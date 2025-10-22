package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/scottwalter/axeos-dashboard/internal/config"
	"github.com/scottwalter/axeos-dashboard/internal/middleware"
)

// safeToFixed safely formats a number to 2 decimal places
func safeToFixed(value float64) string {
	return fmt.Sprintf("%.2f", value)
}

// HandleDashboard serves the dashboard HTML page
func HandleDashboard(cfgManager *config.Manager, publicDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg := cfgManager.GetConfig() // Get fresh config for hot reload
		dashboardHTMLPath := filepath.Join(publicDir, "html", "dashboard.html")

		htmlContent, err := os.ReadFile(dashboardHTMLPath)
		if err != nil {
			fmt.Printf("Error reading dashboard.html: %v\n", err)
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "<h1>Error</h1><p>An internal server error occurred while preparing the dashboard.</p><p>Details: %s</p>", err.Error())
			return
		}

		html := string(htmlContent)

		// Replace placeholders
		title := "AxeOS Dashboard"
		version := "1.0"
		if cfg != nil {
			title = cfg.Title // Use title from config
			version = safeToFixed(cfg.AxeosDashboardVersion)
		}

		currentYear := fmt.Sprintf("%d", time.Now().Year())
		timestamp := time.Now().Format("2006-01-02 15:04:05")

		html = strings.ReplaceAll(html, "<!-- TITLE -->", title)
		html = strings.ReplaceAll(html, "<!-- TIMESTAMP -->", timestamp)
		html = strings.ReplaceAll(html, "<!-- CURRENT_YEAR -->", currentYear)
		html = strings.ReplaceAll(html, "<!-- VERSION -->", version)

		// Handle config outdated warning
		if cfg != nil && cfg.ConfigurationOutdated {
			html = strings.ReplaceAll(html, "<!-- CONFIG VERSION -->", "<B><font color=\"red\">CONFIG FILE OUTDATED!</font></B>")
		} else {
			html = strings.ReplaceAll(html, "<!-- CONFIG VERSION -->", "")
		}

		// Handle login info
		loginInfo := ""
		if cfg != nil && !cfg.DisableAuthentication {
			user := middleware.GetUserFromContext(r)
			if user != nil {
				loginInfo = fmt.Sprintf("<p>Username: %s</p>", user.Username)
			}
		}
		html = strings.ReplaceAll(html, "<!-- LOGIN INFO -->", loginInfo)

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(html))
	}
}

// HandleLoginPage serves the login HTML page
func HandleLoginPage(cfgManager *config.Manager, publicDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg := cfgManager.GetConfig() // Get fresh config for hot reload
		loginHTMLPath := filepath.Join(publicDir, "html", "login.html")

		htmlContent, err := os.ReadFile(loginHTMLPath)
		if err != nil {
			fmt.Printf("Error serving login page: %v\n", err)
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal Server Error"))
			return
		}

		html := string(htmlContent)

		// Replace placeholders
		title := "AxeOS Dashboard"
		version := "1.0"
		currentYear := fmt.Sprintf("%d", time.Now().Year())

		if cfg != nil {
			version = safeToFixed(cfg.AxeosDashboardVersion)
		}

		html = strings.ReplaceAll(html, "<!-- TITLE -->", title)
		html = strings.ReplaceAll(html, "<!-- VERSION -->", version)
		html = strings.ReplaceAll(html, "<!-- CURRENT_YEAR -->", currentYear)

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(html))
	}
}
