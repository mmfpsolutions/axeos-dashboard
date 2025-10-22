package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/scottwalter/axeos-dashboard/internal/database"
	"github.com/scottwalter/axeos-dashboard/internal/services"
)

// collectAxeOSMetrics collects metrics from all configured AxeOS miners
func (m *Manager) collectAxeOSMetrics(ctx context.Context) error {
	cfg, err := m.cfgManager.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	for _, instance := range cfg.AxeosInstances {
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
	infoEndpoint := cfg.AxeosAPI["instanceInfo"]
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
	// Create RPC client to read rpcConfig.json
	configDir := m.cfgManager.GetConfigDir()
	rpcClient := services.NewRPCClient(configDir)

	// Try to load RPC config - if it fails, just return (no error)
	if err := rpcClient.LoadConfig(); err != nil {
		// Config file doesn't exist or can't be loaded - that's okay
		return nil
	}

	// Get list of configured nodes
	nodes := rpcClient.GetConfiguredNodes()
	if len(nodes) == 0 {
		return nil
	}

	// Collect metrics from each node
	for _, nodeID := range nodes {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := m.collectSingleNodeMetric(rpcClient, nodeID); err != nil {
				m.log.Error("Failed to collect node metrics from %s: %v", nodeID, err)
				continue
			}
		}
	}

	return nil
}

// collectSingleNodeMetric collects metrics from a single crypto node
func (m *Manager) collectSingleNodeMetric(rpcClient *services.RPCClient, nodeID string) error {
	metric := &database.NodeMetric{
		Timestamp: time.Now(),
		NodeID:    nodeID,
		NodeName:  nodeID,
	}

	// Get blockchain info (block height, difficulty)
	blockchainInfo, err := rpcClient.CallRPC(nodeID, "getblockchaininfo", []interface{}{})
	if err != nil {
		return fmt.Errorf("failed to get blockchain info: %w", err)
	}

	if blockchainInfo != nil {
		if infoMap, ok := blockchainInfo.(map[string]interface{}); ok {
			if blocks, ok := infoMap["blocks"].(float64); ok {
				metric.BlockHeight = int(blocks)
			}
			if diff, ok := infoMap["difficulty"].(float64); ok {
				metric.Difficulty = diff
			}
		}
	}

	// Get network info (connections)
	networkInfo, err := rpcClient.CallRPC(nodeID, "getnetworkinfo", []interface{}{})
	if err != nil {
		return fmt.Errorf("failed to get network info: %w", err)
	}

	if networkInfo != nil {
		if infoMap, ok := networkInfo.(map[string]interface{}); ok {
			if connections, ok := infoMap["connections"].(float64); ok {
				metric.Connections = int(connections)
			}
		}
	}

	// Insert into database
	if err := m.dbManager.InsertNodeMetric(metric); err != nil {
		return fmt.Errorf("failed to insert node metric: %w", err)
	}

	m.log.Info("Collected node metrics from %s", nodeID)
	return nil
}
