package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/scottwalter/axeos-dashboard/internal/config"
	"github.com/scottwalter/axeos-dashboard/internal/database"
	"github.com/scottwalter/axeos-dashboard/internal/logger"
)

var (
	instance *Manager
	once     sync.Once
)

// Manager handles scheduled data collection tasks
type Manager struct {
	dbManager  *database.Manager
	cfgManager *config.Manager
	tasks      []*Task
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	mu         sync.RWMutex
	log        *logger.Logger
}

// Task represents a scheduled collection task
type Task struct {
	Name     string
	Interval time.Duration
	Ticker   *time.Ticker
	Fn       func(context.Context) error
}

// GetManager returns the singleton scheduler manager instance
func GetManager(dbManager *database.Manager, cfgManager *config.Manager) *Manager {
	once.Do(func() {
		instance = &Manager{
			dbManager:  dbManager,
			cfgManager: cfgManager,
			tasks:      make([]*Task, 0),
			log:        logger.New(logger.ModuleScheduler),
		}
	})
	return instance
}

// Start begins all scheduled collection tasks
func (m *Manager) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.cancel != nil {
		return fmt.Errorf("scheduler already running")
	}

	m.ctx, m.cancel = context.WithCancel(context.Background())

	// Get current configuration
	cfg, err := m.cfgManager.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Register collection tasks based on configuration
	m.registerTasks(cfg)

	// Start all tasks
	for _, task := range m.tasks {
		m.wg.Add(1)
		go m.runTask(task)
	}

	m.log.Info("Scheduler started with %d tasks", len(m.tasks))
	return nil
}

// Stop gracefully stops all scheduled tasks
func (m *Manager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.cancel == nil {
		return
	}

	m.log.Info("Stopping scheduler...")
	m.cancel()

	// Stop all tickers
	for _, task := range m.tasks {
		if task.Ticker != nil {
			task.Ticker.Stop()
		}
	}

	// Wait for all tasks to complete
	m.wg.Wait()

	m.cancel = nil
	m.tasks = make([]*Task, 0)

	m.log.Info("Scheduler stopped")
}

// registerTasks creates collection tasks based on configuration
func (m *Manager) registerTasks(cfg *config.Config) {
	// Default collection interval (5 minutes if not specified)
	defaultInterval := 5 * time.Minute

	// Get collection interval from config (if it exists)
	collectionInterval := defaultInterval
	if cfg.CollectionIntervalSeconds > 0 {
		collectionInterval = time.Duration(cfg.CollectionIntervalSeconds) * time.Second
	}

	// Register AxeOS miner collection task
	if len(cfg.BitaxeInstances) > 0 {
		m.tasks = append(m.tasks, &Task{
			Name:     "AxeOS Miners Collection",
			Interval: collectionInterval,
			Fn:       m.collectAxeOSMetrics,
		})
	}

	// Register Mining Core pool collection task
	if cfg.MiningCoreEnabled && len(cfg.MiningCoreURL) > 0 {
		m.tasks = append(m.tasks, &Task{
			Name:     "Mining Core Pools Collection",
			Interval: collectionInterval,
			Fn:       m.collectPoolMetrics,
		})
	}

	// Register crypto node collection task
	if cfg.CryptNodesEnabled {
		m.tasks = append(m.tasks, &Task{
			Name:     "Crypto Nodes Collection",
			Interval: collectionInterval,
			Fn:       m.collectNodeMetrics,
		})
	}
}

// runTask runs a single scheduled task in a goroutine
func (m *Manager) runTask(task *Task) {
	defer m.wg.Done()

	// Create ticker for this task
	task.Ticker = time.NewTicker(task.Interval)
	defer task.Ticker.Stop()

	m.log.Info("Started task: %s (interval: %v)", task.Name, task.Interval)

	// Run immediately on start
	if err := task.Fn(m.ctx); err != nil {
		m.log.Error("Error in task %s: %v", task.Name, err)
	}

	// Then run on ticker
	for {
		select {
		case <-m.ctx.Done():
			m.log.Info("Stopped task: %s", task.Name)
			return
		case <-task.Ticker.C:
			if err := task.Fn(m.ctx); err != nil {
				m.log.Error("Error in task %s: %v", task.Name, err)
			}
		}
	}
}

// IsRunning returns whether the scheduler is currently running
func (m *Manager) IsRunning() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.cancel != nil
}
