package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/scottwalter/axeos-dashboard/internal/logger"
)

// Config represents the application configuration
type Config struct {
	WebServerPort            int                      `json:"web_server_port"`
	AxeosDashboardVersion    float64                  `json:"axeos_dashboard_version"`
	Title                    string                   `json:"title"`
	AxeosInstances           []map[string]string      `json:"axeos_instances"`
	DisplayFields            interface{}              `json:"display_fields"` // Can be []string or complex nested structure
	MiningCoreEnabled        bool                     `json:"mining_core_enabled"`
	MiningCoreURL            []map[string]string      `json:"mining_core_url"`
	MiningCoreDisplayFields  interface{}              `json:"mining_core_display_fields"` // Can be []string or complex nested structure
	CryptNodesEnabled        bool                     `json:"cryptNodesEnabled"`
	CryptoNodes              interface{}              `json:"cryptoNodes"` // Crypto node configuration
	DisableAuthentication    bool                     `json:"disable_authentication"`
	DisableSettings          bool                     `json:"disable_settings"`
	DisableConfigurations    bool                     `json:"disable_configurations"`
	CookieMaxAge             int                      `json:"cookie_max_age"`
	ConfigurationOutdated    bool                     `json:"configuration_outdated"`
	AxeosAPI                 map[string]string        `json:"axeos_api"`

	// Data collection settings
	DataCollectionEnabled    bool `json:"data_collection_enabled"`
	CollectionIntervalSeconds int  `json:"collection_interval_seconds"`
	DataRetentionDays        int  `json:"data_retention_days"`

	// NOTE: RPC credentials are stored in a separate rpcConfig.json file
	// and should NEVER be exposed through the API or stored in config.json

	mu sync.RWMutex
}

// Manager handles configuration loading and hot-reloading
type Manager struct {
	config     *Config
	configPath string
	mu         sync.RWMutex
	log        *logger.Logger
}

var (
	instance *Manager
	once     sync.Once
)

// GetManager returns the singleton configuration manager instance
func GetManager(configDir string) *Manager {
	once.Do(func() {
		instance = &Manager{
			configPath: filepath.Join(configDir, "config.json"),
			log:        logger.New(logger.ModuleConfig),
		}
	})
	return instance
}

// LoadConfig loads the configuration from file
func (m *Manager) LoadConfig() (*Config, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.log.Info("Loading configuration from: %s", m.configPath)

	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}

	// Set default values if not present
	config.ConfigurationOutdated = false

	// Apply defaults for missing fields
	if config.CookieMaxAge == 0 {
		config.CookieMaxAge = 3600 // 1 hour default
	}

	// Apply defaults for data collection
	if config.CollectionIntervalSeconds == 0 {
		config.CollectionIntervalSeconds = 300 // 5 minutes default
	}
	if config.DataRetentionDays == 0 {
		config.DataRetentionDays = 30 // 30 days default
	}

	m.config = &config
	m.log.Info("Configuration loaded successfully")

	return &config, nil
}

// ReloadConfig reloads the configuration from file
func (m *Manager) ReloadConfig() (*Config, error) {
	m.log.Info("Reloading configuration...")
	return m.LoadConfig()
}

// GetConfig returns the current configuration
func (m *Manager) GetConfig() *Config {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config
}

// GetConfigDir returns the configuration directory path
func (m *Manager) GetConfigDir() string {
	return filepath.Dir(m.configPath)
}

// UpdateConfig updates the configuration file with new values
func (m *Manager) UpdateConfig(updates map[string]interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Read current config
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return fmt.Errorf("error reading config file: %w", err)
	}

	var currentConfig map[string]interface{}
	if err := json.Unmarshal(data, &currentConfig); err != nil {
		return fmt.Errorf("error parsing config file: %w", err)
	}

	// Apply updates
	for key, value := range updates {
		currentConfig[key] = value
	}

	// Write back to file
	updatedData, err := json.MarshalIndent(currentConfig, "", "    ")
	if err != nil {
		return fmt.Errorf("error marshaling config: %w", err)
	}

	if err := os.WriteFile(m.configPath, updatedData, 0644); err != nil {
		return fmt.Errorf("error writing config file: %w", err)
	}

	// Reload config into memory (unlock first to avoid deadlock)
	m.mu.Unlock()
	_, err = m.LoadConfig()
	m.mu.Lock() // Re-lock before defer unlocks
	return err
}

// CheckConfigFilesExist checks if all required configuration files exist
func CheckConfigFilesExist(configDir string) bool {
	requiredFiles := []string{"config.json", "access.json", "jsonWebTokenKey.json"}

	for _, file := range requiredFiles {
		path := filepath.Join(configDir, file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return false
		}
	}

	return true
}
