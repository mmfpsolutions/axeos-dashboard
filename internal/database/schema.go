package database

const (
	// Schema for AxeOS miner metrics
	createAxeOSMetricsTable = `
		CREATE TABLE IF NOT EXISTS axeos_metrics (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp DATETIME NOT NULL,
			instance_id TEXT NOT NULL,
			instance_name TEXT NOT NULL,
			hashrate REAL,
			temperature REAL,
			power REAL,
			fan_speed INTEGER,
			best_diff TEXT,
			shares_accepted INTEGER,
			shares_rejected INTEGER,
			frequency INTEGER,
			voltage REAL,
			core_voltage REAL
		);
	`

	createAxeOSMetricsIndexes = `
		CREATE INDEX IF NOT EXISTS idx_axeos_timestamp ON axeos_metrics(timestamp);
		CREATE INDEX IF NOT EXISTS idx_axeos_instance ON axeos_metrics(instance_id);
	`

	// Schema for Mining Core pool metrics
	createPoolMetricsTable = `
		CREATE TABLE IF NOT EXISTS pool_metrics (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp DATETIME NOT NULL,
			pool_id TEXT NOT NULL,
			pool_name TEXT NOT NULL,
			pool_hashrate REAL,
			pool_workers INTEGER,
			network_hashrate REAL,
			network_difficulty REAL,
			last_block_time DATETIME,
			blocks_found INTEGER
		);
	`

	createPoolMetricsIndexes = `
		CREATE INDEX IF NOT EXISTS idx_pool_timestamp ON pool_metrics(timestamp);
		CREATE INDEX IF NOT EXISTS idx_pool_id ON pool_metrics(pool_id);
	`

	// Schema for crypto node metrics
	createNodeMetricsTable = `
		CREATE TABLE IF NOT EXISTS node_metrics (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp DATETIME NOT NULL,
			node_id TEXT NOT NULL,
			node_name TEXT NOT NULL,
			block_height INTEGER,
			connections INTEGER,
			difficulty REAL,
			network_hashrate REAL
		);
	`

	createNodeMetricsIndexes = `
		CREATE INDEX IF NOT EXISTS idx_node_timestamp ON node_metrics(timestamp);
		CREATE INDEX IF NOT EXISTS idx_node_id ON node_metrics(node_id);
	`
)

// initializeSchema creates all necessary tables and indexes
func (m *Manager) initializeSchema() error {
	statements := []string{
		createAxeOSMetricsTable,
		createAxeOSMetricsIndexes,
		createPoolMetricsTable,
		createPoolMetricsIndexes,
		createNodeMetricsTable,
		createNodeMetricsIndexes,
	}

	for _, stmt := range statements {
		if _, err := m.db.Exec(stmt); err != nil {
			return err
		}
	}

	return nil
}
