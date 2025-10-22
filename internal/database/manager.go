package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/scottwalter/axeos-dashboard/internal/logger"
	_ "modernc.org/sqlite"
)

var (
	instance *Manager
	once     sync.Once
)

// Manager handles SQLite database connections and operations
type Manager struct {
	db       *sql.DB
	dataPath string
	mu       sync.RWMutex
	log      *logger.Logger
}

// GetManager returns the singleton database manager instance
func GetManager(dataPath string) *Manager {
	once.Do(func() {
		instance = &Manager{
			dataPath: dataPath,
			log:      logger.New(logger.ModuleDatabase),
		}
	})
	return instance
}

// Initialize sets up the SQLite database connection and creates tables
func (m *Manager) Initialize() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Ensure data directory exists
	if err := os.MkdirAll(m.dataPath, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	// Database file path
	dbFile := filepath.Join(m.dataPath, "metrics.db")

	// Open SQLite connection
	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		return fmt.Errorf("failed to open SQLite: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping SQLite: %w", err)
	}

	// Set SQLite pragmas for better performance
	_, err = db.Exec(`
		PRAGMA journal_mode=WAL;
		PRAGMA synchronous=NORMAL;
		PRAGMA cache_size=-64000;
		PRAGMA busy_timeout=5000;
	`)
	if err != nil {
		return fmt.Errorf("failed to set SQLite pragmas: %w", err)
	}

	m.db = db

	// Initialize schema
	if err := m.initializeSchema(); err != nil {
		return fmt.Errorf("failed to initialize schema: %w", err)
	}

	m.log.Info("SQLite initialized successfully at: %s", dbFile)
	return nil
}

// Close closes the database connection
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.db != nil {
		return m.db.Close()
	}
	return nil
}

// DB returns the database connection (for queries)
func (m *Manager) DB() *sql.DB {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.db
}
