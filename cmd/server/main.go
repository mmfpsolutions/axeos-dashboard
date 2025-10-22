package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/scottwalter/axeos-dashboard/internal/auth"
	"github.com/scottwalter/axeos-dashboard/internal/config"
	"github.com/scottwalter/axeos-dashboard/internal/database"
	"github.com/scottwalter/axeos-dashboard/internal/logger"
	"github.com/scottwalter/axeos-dashboard/internal/router"
	"github.com/scottwalter/axeos-dashboard/internal/scheduler"
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
	log := logger.New(logger.ModuleMain)

	// Check if we're in bootstrap mode and config files now exist
	if h.isBootstrapMode {
		if config.CheckConfigFilesExist(h.configDir) {
			log.Info("Configuration files detected. Switching to normal mode...")

			// Initialize JWT service
			if err := auth.InitJWTService(h.configDir); err != nil {
				log.Error("Error initializing JWT service: %v", err)
				http.Error(w, "Failed to initialize authentication", http.StatusInternalServerError)
				return
			}

			// Load configuration
			h.cfgManager = config.GetManager(h.configDir)
			cfg, err := h.cfgManager.LoadConfig()
			if err != nil {
				log.Error("Error loading configuration: %v", err)
				http.Error(w, "Failed to load configuration", http.StatusInternalServerError)
				return
			}

			// Setup normal router
			h.normalHandler = router.SetupRouter(h.cfgManager, cfg, h.configDir, h.publicDir)
			h.isBootstrapMode = false

			log.Info("Successfully switched to normal mode!")
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
	log := logger.New(logger.ModuleMain)
	if err := run(); err != nil {
		log.Fatal("FAILED TO START SERVER: %v", err)
	}
}

func run() error {
	log := logger.New(logger.ModuleMain)

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
	dataDir := filepath.Join(baseDir, "data")

	log.Info("Base directory: %s", baseDir)
	log.Info("Config directory: %s", configDir)
	log.Info("Public directory: %s", publicDir)
	log.Info("Data directory: %s", dataDir)

	// Check if configuration files exist
	configFilesExist := config.CheckConfigFilesExist(configDir)
	log.Info("Config files exist: %v", configFilesExist)

	var cfg *config.Config
	var isBootstrapMode bool
	var cfgManager *config.Manager
	var dbManager *database.Manager
	var schedManager *scheduler.Manager

	if !configFilesExist {
		log.Info("Configuration files missing. Starting in bootstrap mode...")
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

		// Initialize database if data collection is enabled
		if cfg.DataCollectionEnabled {
			dbManager = database.GetManager(dataDir)
			if err := dbManager.Initialize(); err != nil {
				return fmt.Errorf("failed to initialize database: %w", err)
			}
			defer dbManager.Close()

			// Initialize scheduler
			schedManager = scheduler.GetManager(dbManager, cfgManager)
			if err := schedManager.Start(); err != nil {
				return fmt.Errorf("failed to start scheduler: %w", err)
			}
			defer schedManager.Stop()

			log.Info("Data collection enabled and scheduler started")
		} else {
			log.Info("Data collection disabled")
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
		configDir:        configDir,
		publicDir:        publicDir,
		isBootstrapMode:  isBootstrapMode,
		cfgManager:       cfgManager,
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

	log.Info("Server running on http://localhost:%d", port)
	log.Info("Server started at: %s", time.Now().Format(time.RFC3339))
	log.Info("Config directory: %s", configDir)
	log.Info("Public directory: %s", publicDir)

	// Setup graceful shutdown
	serverErr := make(chan error, 1)
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
		log.Info("Shutdown signal received, gracefully shutting down...")
	case err := <-serverErr:
		return fmt.Errorf("server error: %w", err)
	}

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	log.Info("Server stopped gracefully")
	return nil
}
