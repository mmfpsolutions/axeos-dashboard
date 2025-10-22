package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/scottwalter/axeos-dashboard/internal/database"
)

// collectAxeOSMetrics collects metrics from all configured AxeOS miners
func (m *Manager) collectAxeOSMetrics(ctx context.Context) error {
	cfg, err := m.cfgManager.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	for _, instance := range cfg.BitaxeInstances {
		for name, baseURL := range instance {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				if err := m.collectSingleAxeOSMetric(name, baseURL); err != nil {
					m.log.Error("Failed to collect AxeOS metrics from %s: %v", name, err)
					// Continue with other instances even if one fails
					continue
				}
			}
		}
	}

	return nil
}

// collectSingleAxeOSMetric collects metrics from a single AxeOS miner
func (m *Manager) collectSingleAxeOSMetric(instanceName, baseURL string) error {
	cfg, err := m.cfgManager.LoadConfig()
	if err != nil {
		return err
	}

	// Fetch system info
	infoEndpoint := cfg.BitaxeAPI["instanceInfo"]
	if infoEndpoint == "" {
		infoEndpoint = "/api/system/info" // Default endpoint
	}
	infoURL := baseURL + infoEndpoint
	resp, err := http.Get(infoURL)
	if err != nil {
		return fmt.Errorf("failed to fetch info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Extract metrics and save to database
	metric := &database.AxeOSMetric{
		Timestamp:    time.Now(),
		InstanceID:   instanceName,
		InstanceName: instanceName,
	}

	// Parse fields (with safe type assertions and default values)
	if hashrate, ok := data["hashRate"].(float64); ok {
		metric.Hashrate = hashrate
	}
	if temp, ok := data["temp"].(float64); ok {
		metric.Temperature = temp
	}
	if power, ok := data["power"].(float64); ok {
		metric.Power = power
	}
	if fanSpeed, ok := data["fanSpeed"].(float64); ok {
		metric.FanSpeed = int(fanSpeed)
	}
	if bestDiff, ok := data["bestDiff"].(string); ok {
		metric.BestDiff = bestDiff
	}
	if sharesAccepted, ok := data["sharesAccepted"].(float64); ok {
		metric.SharesAccepted = int(sharesAccepted)
	}
	if sharesRejected, ok := data["sharesRejected"].(float64); ok {
		metric.SharesRejected = int(sharesRejected)
	}
	if freq, ok := data["frequency"].(float64); ok {
		metric.Frequency = int(freq)
	}
	if voltage, ok := data["voltage"].(float64); ok {
		metric.Voltage = voltage
	}
	if coreVoltage, ok := data["coreVoltage"].(float64); ok {
		metric.CoreVoltage = coreVoltage
	}

	// Insert into database
	if err := m.dbManager.InsertAxeOSMetric(metric); err != nil {
		return fmt.Errorf("failed to insert metric: %w", err)
	}

	m.log.Info("Collected AxeOS metrics from %s", instanceName)
	return nil
}

// collectPoolMetrics collects metrics from all configured Mining Core pools
func (m *Manager) collectPoolMetrics(ctx context.Context) error {
	cfg, err := m.cfgManager.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	for _, poolMap := range cfg.MiningCoreURL {
		for poolName, poolURL := range poolMap {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				if err := m.collectSinglePoolMetric(poolName, poolURL); err != nil {
					m.log.Error("Failed to collect pool metrics from %s: %v", poolName, err)
					continue
				}
			}
		}
	}

	return nil
}

// collectSinglePoolMetric collects metrics from a single Mining Core pool
func (m *Manager) collectSinglePoolMetric(poolName, poolURL string) error {
	// Fetch pool stats (adjust endpoint based on Mining Core API)
	statsURL := poolURL + "/api/pools"
	resp, err := http.Get(statsURL)
	if err != nil {
		return fmt.Errorf("failed to fetch pool stats: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Extract and save pool metrics
	metric := &database.PoolMetric{
		Timestamp: time.Now(),
		PoolID:    poolName,
		PoolName:  poolName,
	}

	// Parse fields (adjust based on actual Mining Core API response structure)
	if hashrate, ok := data["poolHashrate"].(float64); ok {
		metric.PoolHashrate = hashrate
	}
	if workers, ok := data["poolWorkers"].(float64); ok {
		metric.PoolWorkers = int(workers)
	}
	if netHashrate, ok := data["networkHashrate"].(float64); ok {
		metric.NetworkHashrate = netHashrate
	}
	if netDiff, ok := data["networkDifficulty"].(float64); ok {
		metric.NetworkDifficulty = netDiff
	}
	if blocks, ok := data["totalBlocks"].(float64); ok {
		metric.BlocksFound = int(blocks)
	}

	// Insert into database
	if err := m.dbManager.InsertPoolMetric(metric); err != nil {
		return fmt.Errorf("failed to insert pool metric: %w", err)
	}

	m.log.Info("Collected pool metrics from %s", poolName)
	return nil
}

// collectNodeMetrics collects metrics from all configured crypto nodes
func (m *Manager) collectNodeMetrics(ctx context.Context) error {
	cfg, err := m.cfgManager.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Check if RPC config exists
	if cfg.RPCConfig == nil {
		return nil
	}

	for nodeName, nodeConfig := range cfg.RPCConfig {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := m.collectSingleNodeMetric(nodeName, nodeConfig); err != nil {
				m.log.Error("Failed to collect node metrics from %s: %v", nodeName, err)
				continue
			}
		}
	}

	return nil
}

// collectSingleNodeMetric collects metrics from a single crypto node
func (m *Manager) collectSingleNodeMetric(nodeName string, nodeConfig map[string]interface{}) error {
	// This is a placeholder - you'll need to implement RPC calls based on your existing services/rpc.go
	// For now, we'll create a basic structure

	metric := &database.NodeMetric{
		Timestamp: time.Now(),
		NodeID:    nodeName,
		NodeName:  nodeName,
	}

	// TODO: Implement actual RPC calls to fetch node metrics
	// This should integrate with your existing internal/services/rpc.go functionality
	// For example:
	// - getblockchaininfo for block height, difficulty
	// - getnetworkinfo for connections, version
	// - getmempoolinfo for mempool stats

	// Insert into database
	if err := m.dbManager.InsertNodeMetric(metric); err != nil {
		return fmt.Errorf("failed to insert node metric: %w", err)
	}

	m.log.Info("Collected node metrics from %s", nodeName)
	return nil
}
