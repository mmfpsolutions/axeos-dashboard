package router

import (
	"net/http"

	"github.com/scottwalter/axeos-dashboard/internal/handlers"
)

// SetupBootstrapRouter sets up routes for bootstrap mode (first-time setup)
func SetupBootstrapRouter(configDir, publicDir string) http.Handler {
	mux := http.NewServeMux()

	// Serve static files (CSS, JS, images, fonts)
	fileServer := http.FileServer(http.Dir(publicDir))
	mux.Handle("/public/", http.StripPrefix("/public/", fileServer))

	// Bootstrap page (GET)
	mux.HandleFunc("/", handlers.HandleBootstrapPage)

	// Bootstrap form submission (POST)
	mux.HandleFunc("/bootstrap", handlers.HandleBootstrapSubmit(configDir))

	return mux
}
