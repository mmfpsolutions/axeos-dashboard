package database

import (
	"database/sql"
	"fmt"
)

// InsertAxeOSMetric inserts a single AxeOS metric into the database
func (m *Manager) InsertAxeOSMetric(metric *AxeOSMetric) error {
	query := `
		INSERT INTO axeos_metrics (
			timestamp, instance_id, instance_name, hashrate, temperature, power,
			fan_speed, best_diff, shares_accepted, shares_rejected,
			frequency, voltage, core_voltage
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := m.db.Exec(query,
		metric.Timestamp,
		metric.InstanceID,
		metric.InstanceName,
		metric.Hashrate,
		metric.Temperature,
		metric.Power,
		metric.FanSpeed,
		metric.BestDiff,
		metric.SharesAccepted,
		metric.SharesRejected,
		metric.Frequency,
		metric.Voltage,
		metric.CoreVoltage,
	)

	if err != nil {
		return fmt.Errorf("failed to insert AxeOS metric: %w", err)
	}

	return nil
}

// InsertPoolMetric inserts a single pool metric into the database
func (m *Manager) InsertPoolMetric(metric *PoolMetric) error {
	query := `
		INSERT INTO pool_metrics (
			timestamp, pool_id, pool_name, pool_hashrate, pool_workers,
			network_hashrate, network_difficulty, last_block_time, blocks_found
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	var lastBlockTime interface{}
	if metric.LastBlockTime != nil {
		lastBlockTime = *metric.LastBlockTime
	}

	_, err := m.db.Exec(query,
		metric.Timestamp,
		metric.PoolID,
		metric.PoolName,
		metric.PoolHashrate,
		metric.PoolWorkers,
		metric.NetworkHashrate,
		metric.NetworkDifficulty,
		lastBlockTime,
		metric.BlocksFound,
	)

	if err != nil {
		return fmt.Errorf("failed to insert pool metric: %w", err)
	}

	return nil
}

// InsertNodeMetric inserts a single node metric into the database
func (m *Manager) InsertNodeMetric(metric *NodeMetric) error {
	query := `
		INSERT INTO node_metrics (
			timestamp, node_id, node_name, block_height, connections,
			difficulty, network_hashrate
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	_, err := m.db.Exec(query,
		metric.Timestamp,
		metric.NodeID,
		metric.NodeName,
		metric.BlockHeight,
		metric.Connections,
		metric.Difficulty,
		metric.NetworkHashrate,
	)

	if err != nil {
		return fmt.Errorf("failed to insert node metric: %w", err)
	}

	return nil
}

// GetAxeOSMetrics retrieves AxeOS metrics for a specific instance within a time range
func (m *Manager) GetAxeOSMetrics(instanceID string, startTime, endTime string, limit int) ([]*AxeOSMetric, error) {
	query := `
		SELECT timestamp, instance_id, instance_name, hashrate, temperature, power,
		       fan_speed, best_diff, shares_accepted, shares_rejected,
		       frequency, voltage, core_voltage
		FROM axeos_metrics
		WHERE instance_id = ? AND timestamp BETWEEN ? AND ?
		ORDER BY timestamp DESC
		LIMIT ?
	`

	rows, err := m.db.Query(query, instanceID, startTime, endTime, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query AxeOS metrics: %w", err)
	}
	defer rows.Close()

	return scanAxeOSMetrics(rows)
}

// GetPoolMetrics retrieves pool metrics for a specific pool within a time range
func (m *Manager) GetPoolMetrics(poolID string, startTime, endTime string, limit int) ([]*PoolMetric, error) {
	query := `
		SELECT timestamp, pool_id, pool_name, pool_hashrate, pool_workers,
		       network_hashrate, network_difficulty, last_block_time, blocks_found
		FROM pool_metrics
		WHERE pool_id = ? AND timestamp BETWEEN ? AND ?
		ORDER BY timestamp DESC
		LIMIT ?
	`

	rows, err := m.db.Query(query, poolID, startTime, endTime, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query pool metrics: %w", err)
	}
	defer rows.Close()

	return scanPoolMetrics(rows)
}

// GetNodeMetrics retrieves node metrics for a specific node within a time range
func (m *Manager) GetNodeMetrics(nodeID string, startTime, endTime string, limit int) ([]*NodeMetric, error) {
	query := `
		SELECT timestamp, node_id, node_name, block_height, connections,
		       difficulty, network_hashrate
		FROM node_metrics
		WHERE node_id = ? AND timestamp BETWEEN ? AND ?
		ORDER BY timestamp DESC
		LIMIT ?
	`

	rows, err := m.db.Query(query, nodeID, startTime, endTime, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query node metrics: %w", err)
	}
	defer rows.Close()

	return scanNodeMetrics(rows)
}

// Helper functions to scan rows into structs

func scanAxeOSMetrics(rows *sql.Rows) ([]*AxeOSMetric, error) {
	var metrics []*AxeOSMetric

	for rows.Next() {
		metric := &AxeOSMetric{}
		err := rows.Scan(
			&metric.Timestamp,
			&metric.InstanceID,
			&metric.InstanceName,
			&metric.Hashrate,
			&metric.Temperature,
			&metric.Power,
			&metric.FanSpeed,
			&metric.BestDiff,
			&metric.SharesAccepted,
			&metric.SharesRejected,
			&metric.Frequency,
			&metric.Voltage,
			&metric.CoreVoltage,
		)
		if err != nil {
			return nil, err
		}
		metrics = append(metrics, metric)
	}

	return metrics, rows.Err()
}

func scanPoolMetrics(rows *sql.Rows) ([]*PoolMetric, error) {
	var metrics []*PoolMetric

	for rows.Next() {
		metric := &PoolMetric{}
		var lastBlockTime sql.NullTime

		err := rows.Scan(
			&metric.Timestamp,
			&metric.PoolID,
			&metric.PoolName,
			&metric.PoolHashrate,
			&metric.PoolWorkers,
			&metric.NetworkHashrate,
			&metric.NetworkDifficulty,
			&lastBlockTime,
			&metric.BlocksFound,
		)
		if err != nil {
			return nil, err
		}

		if lastBlockTime.Valid {
			metric.LastBlockTime = &lastBlockTime.Time
		}

		metrics = append(metrics, metric)
	}

	return metrics, rows.Err()
}

func scanNodeMetrics(rows *sql.Rows) ([]*NodeMetric, error) {
	var metrics []*NodeMetric

	for rows.Next() {
		metric := &NodeMetric{}
		err := rows.Scan(
			&metric.Timestamp,
			&metric.NodeID,
			&metric.NodeName,
			&metric.BlockHeight,
			&metric.Connections,
			&metric.Difficulty,
			&metric.NetworkHashrate,
		)
		if err != nil {
			return nil, err
		}
		metrics = append(metrics, metric)
	}

	return metrics, rows.Err()
}

// CleanupOldMetrics deletes metrics older than the specified retention period (in days)
func (m *Manager) CleanupOldMetrics(retentionDays int) error {
	queries := []string{
		fmt.Sprintf("DELETE FROM axeos_metrics WHERE timestamp < NOW() - INTERVAL '%d days'", retentionDays),
		fmt.Sprintf("DELETE FROM pool_metrics WHERE timestamp < NOW() - INTERVAL '%d days'", retentionDays),
		fmt.Sprintf("DELETE FROM node_metrics WHERE timestamp < NOW() - INTERVAL '%d days'", retentionDays),
	}

	for _, query := range queries {
		if _, err := m.db.Exec(query); err != nil {
			return fmt.Errorf("failed to cleanup old metrics: %w", err)
		}
	}

	return nil
}
