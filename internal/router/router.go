package router

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/scottwalter/axeos-dashboard/internal/config"
	"github.com/scottwalter/axeos-dashboard/internal/handlers"
	"github.com/scottwalter/axeos-dashboard/internal/middleware"
	"github.com/scottwalter/axeos-dashboard/internal/services"
)

// SetupRouter configures all routes for the application
func SetupRouter(cfgManager *config.Manager, cfg *config.Config, configDir, publicDir string) http.Handler {
	mux := http.NewServeMux()

	cryptoNodeSvc := services.NewCryptoNodeService(configDir)

	// Static assets - no authentication required
	publicPath := "/public/"
	mux.Handle(publicPath, http.StripPrefix(publicPath,
		middleware.LoggingMiddleware(
			http.FileServer(http.Dir(publicDir)),
		),
	))

	// Login page - no authentication required
	mux.Handle("/login",
		middleware.LoggingMiddleware(
			http.HandlerFunc(handlers.HandleLoginPage(cfgManager, publicDir)),
		),
	)

	// Login API endpoint - no authentication required
	mux.Handle("/api/login",
		middleware.LoggingMiddleware(
			handlers.HandleLogin(configDir),
		),
	)

	// Logout API endpoint - no authentication required
	mux.Handle("/api/logout",
		middleware.LoggingMiddleware(
			http.HandlerFunc(handlers.HandleLogout),
		),
	)

	// Dashboard page - authentication required
	dashboardHandler := middleware.AuthMiddleware(cfgManager, true)(
		http.HandlerFunc(handlers.HandleDashboard(cfgManager, publicDir)),
	)
	mux.Handle("/", middleware.LoggingMiddleware(dashboardHandler))
	mux.Handle("/index.html", middleware.LoggingMiddleware(dashboardHandler))

	// API endpoints - authentication required
	apiAuthMiddleware := middleware.AuthMiddleware(cfgManager, true)

	// Systems info
	mux.Handle("/api/systems/info",
		middleware.LoggingMiddleware(
			apiAuthMiddleware(handlers.HandleSystemsInfo(cfgManager, cryptoNodeSvc)),
		),
	)

	// Instance info
	mux.Handle("/api/instance/info",
		middleware.LoggingMiddleware(
			apiAuthMiddleware(handlers.HandleInstanceInfo(cfgManager)),
		),
	)

	// Instance restart
	mux.Handle("/api/instance/service/restart",
		middleware.LoggingMiddleware(
			apiAuthMiddleware(handlers.HandleInstanceRestart(cfgManager)),
		),
	)

	// Instance settings
	mux.Handle("/api/instance/service/settings",
		middleware.LoggingMiddleware(
			apiAuthMiddleware(handlers.HandleInstanceSettings(cfgManager)),
		),
	)

	// Configuration endpoint
	mux.Handle("/api/configuration",
		middleware.LoggingMiddleware(
			apiAuthMiddleware(handlers.HandleConfiguration(cfgManager, cfg)),
		),
	)

	// Statistics endpoint
	mux.Handle("/api/statistics",
		middleware.LoggingMiddleware(
			apiAuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handlers.HandleStatistics(w, r, cfgManager)
			})),
		),
	)

	// Migration status endpoint
	mux.Handle("/api/migration/status",
		middleware.LoggingMiddleware(
			apiAuthMiddleware(http.HandlerFunc(handlers.HandleMigrationStatus)),
		),
	)

	// Migration clear endpoint
	mux.Handle("/api/migration/clear",
		middleware.LoggingMiddleware(
			apiAuthMiddleware(http.HandlerFunc(handlers.HandleMigrationClear)),
		),
	)

	return mux
}

// ServeStaticAsset serves a static file with proper MIME type
func ServeStaticAsset(publicDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Security check: prevent directory traversal
		cleanPath := filepath.Clean(r.URL.Path)
		if strings.Contains(cleanPath, "..") {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("Forbidden"))
			return
		}

		// Get file extension and set Content-Type
		ext := strings.ToLower(filepath.Ext(cleanPath))
		contentType := getContentType(ext)

		w.Header().Set("Content-Type", contentType)
		http.ServeFile(w, r, filepath.Join(publicDir, cleanPath))
	}
}

func getContentType(ext string) string {
	mimeTypes := map[string]string{
		".css":  "text/css",
		".js":   "application/javascript",
		".png":  "image/png",
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".gif":  "image/gif",
		".svg":  "image/svg+xml",
		".ico":  "image/x-icon",
		".webp": "image/webp",
		".html": "text/html",
	}

	if ct, ok := mimeTypes[ext]; ok {
		return ct
	}
	return "application/octet-stream"
}
