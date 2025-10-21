package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/scottwalter/axeos-dashboard/internal/auth"
	"github.com/scottwalter/axeos-dashboard/internal/config"
	"github.com/scottwalter/axeos-dashboard/internal/router"
)

const (
	DefaultWebServerPort = 3000
)

// dynamicHandler wraps the bootstrap and normal handlers,
// allowing hot-reload from bootstrap mode to normal mode
type dynamicHandler struct {
	configDir        string
	publicDir        string
	isBootstrapMode  bool
	cfgManager       *config.Manager
	bootstrapHandler http.Handler
	normalHandler    http.Handler
}

// ServeHTTP implements http.Handler interface
func (h *dynamicHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Check if we're in bootstrap mode and config files now exist
	if h.isBootstrapMode {
		if config.CheckConfigFilesExist(h.configDir) {
			fmt.Println("Configuration files detected. Switching to normal mode...")

			// Initialize JWT service
			if err := auth.InitJWTService(h.configDir); err != nil {
				fmt.Printf("Error initializing JWT service: %v\n", err)
				http.Error(w, "Failed to initialize authentication", http.StatusInternalServerError)
				return
			}

			// Load configuration
			h.cfgManager = config.GetManager(h.configDir)
			cfg, err := h.cfgManager.LoadConfig()
			if err != nil {
				fmt.Printf("Error loading configuration: %v\n", err)
				http.Error(w, "Failed to load configuration", http.StatusInternalServerError)
				return
			}

			// Setup normal router
			h.normalHandler = router.SetupRouter(h.cfgManager, cfg, h.configDir, h.publicDir)
			h.isBootstrapMode = false

			fmt.Println("Successfully switched to normal mode!")
		}
	}

	// Route to appropriate handler
	if h.isBootstrapMode {
		h.bootstrapHandler.ServeHTTP(w, r)
	} else {
		h.normalHandler.ServeHTTP(w, r)
	}
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("FAILED TO START SERVER: %v", err)
	}
}

func run() error {
	// Determine paths
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	baseDir := filepath.Dir(execPath)

	// For development, use current directory
	if _, err := os.Stat("./config"); err == nil {
		baseDir, _ = os.Getwd()
	}

	configDir := filepath.Join(baseDir, "config")
	publicDir := filepath.Join(baseDir, "public")

	fmt.Printf("Base directory: %s\n", baseDir)
	fmt.Printf("Config directory: %s\n", configDir)
	fmt.Printf("Public directory: %s\n", publicDir)

	// Check if configuration files exist
	configFilesExist := config.CheckConfigFilesExist(configDir)
	fmt.Printf("Config files exist: %v\n", configFilesExist)

	var cfg *config.Config
	var isBootstrapMode bool
	var cfgManager *config.Manager

	if !configFilesExist {
		fmt.Println("Configuration files missing. Starting in bootstrap mode...")
		isBootstrapMode = true
		cfg = &config.Config{
			WebServerPort: DefaultWebServerPort,
		}
	} else {
		// Initialize JWT service
		if err := auth.InitJWTService(configDir); err != nil {
			return fmt.Errorf("failed to initialize JWT service: %w", err)
		}

		// Load configuration
		cfgManager = config.GetManager(configDir)
		cfg, err = cfgManager.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}
	}

	// Determine port
	port := DefaultWebServerPort
	if portEnv := os.Getenv("PORT"); portEnv != "" {
		if p, err := strconv.Atoi(portEnv); err == nil {
			port = p
		}
	} else if cfg.WebServerPort != 0 {
		port = cfg.WebServerPort
	}

	// Create dynamic handler that can switch from bootstrap to normal mode
	handler := &dynamicHandler{
		configDir:       configDir,
		publicDir:       publicDir,
		isBootstrapMode: isBootstrapMode,
		cfgManager:      cfgManager,
		bootstrapHandler: router.SetupBootstrapRouter(configDir, publicDir),
	}

	// Initialize normal handler if not in bootstrap mode
	if !isBootstrapMode {
		handler.normalHandler = router.SetupRouter(cfgManager, cfg, configDir, publicDir)
	}

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	fmt.Printf("Server running on http://localhost:%d\n", port)
	fmt.Printf("Server started at: %s\n", time.Now().Format(time.RFC3339))
	fmt.Printf("Config directory: %s\n", configDir)
	fmt.Printf("Public directory: %s\n", publicDir)

	// Start server
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}
