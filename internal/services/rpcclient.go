package services

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/scottwalter/axeos-dashboard/internal/logger"
)

// RPCConfig represents the rpcConfig.json structure
type RPCConfig struct {
	CryptoNodes []RPCNodeConfig `json:"cryptoNodes"`
}

// RPCNodeConfig represents a single node's RPC configuration
type RPCNodeConfig struct {
	NodeID         string `json:"NodeId"`
	NodeRPCAddress string `json:"NodeRPCAddress"`
	NodeRPCPort    int    `json:"NodeRPCPort"`
	NodeRPAuth     string `json:"NodeRPAuth"`
}

// RPCClient handles JSON-RPC calls to cryptocurrency nodes
type RPCClient struct {
	configDir string
	rpcConfig *RPCConfig
	mu        sync.RWMutex
	client    *http.Client
	log       *logger.Logger
}

// RPCRequest represents a JSON-RPC request
type RPCRequest struct {
	JSONRpc string        `json:"jsonrpc"`
	ID      string        `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

// RPCResponse represents a JSON-RPC response
type RPCResponse struct {
	Result interface{}          `json:"result"`
	Error  *RPCError            `json:"error"`
	ID     string               `json:"id"`
}

// RPCError represents a JSON-RPC error
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// NewRPCClient creates a new RPC client
func NewRPCClient(configDir string) *RPCClient {
	return &RPCClient{
		configDir: configDir,
		client: &http.Client{
			Timeout: 30 * 1000000000, // 30 seconds in nanoseconds
		},
		log: logger.New(logger.ModuleService),
	}
}

// loadRPCConfig loads the RPC configuration from rpcConfig.json
func (r *RPCClient) loadRPCConfig() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	configPath := filepath.Join(r.configDir, "rpcConfig.json")

	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read rpcConfig.json: %w", err)
	}

	var config RPCConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse rpcConfig.json: %w", err)
	}

	r.rpcConfig = &config
	return nil
}

// getRPCConnectionDetails gets RPC connection details for a specific node ID
func (r *RPCClient) getRPCConnectionDetails(nodeID string) (*RPCNodeConfig, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.rpcConfig == nil {
		return nil, fmt.Errorf("RPC config not loaded")
	}

	for _, node := range r.rpcConfig.CryptoNodes {
		if node.NodeID == nodeID {
			return &node, nil
		}
	}

	return nil, fmt.Errorf("node ID '%s' not found in rpcConfig.json", nodeID)
}

// CallRPC makes a JSON-RPC call to a cryptocurrency node
func (r *RPCClient) CallRPC(nodeID, method string, params []interface{}) (interface{}, error) {
	// Ensure config is loaded
	if r.rpcConfig == nil {
		if err := r.loadRPCConfig(); err != nil {
			return nil, err
		}
	}

	// Get connection details
	nodeConfig, err := r.getRPCConnectionDetails(nodeID)
	if err != nil {
		return nil, err
	}

	// Create RPC request
	rpcReq := RPCRequest{
		JSONRpc: "2.0",
		ID:      "axeos-dashboard",
		Method:  method,
		Params:  params,
	}

	reqBody, err := json.Marshal(rpcReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal RPC request: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("http://%s:%d", nodeConfig.NodeRPCAddress, nodeConfig.NodeRPCPort)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	authEncoded := base64.StdEncoding.EncodeToString([]byte(nodeConfig.NodeRPAuth))
	req.Header.Set("Authorization", "Basic "+authEncoded)

	r.log.Info("Sending RPC request to %s:%d - Method: %s",
		nodeConfig.NodeRPCAddress, nodeConfig.NodeRPCPort, method)

	// Send request
	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("RPC request error: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for empty response (often indicates auth failure)
	if len(body) == 0 {
		return nil, fmt.Errorf("empty response from RPC server. Check RPC credentials (rpcauth) and rpcallowip in node config. Status: %d", resp.StatusCode)
	}

	// Parse response
	var rpcResp RPCResponse
	if err := json.Unmarshal(body, &rpcResp); err != nil {
		return nil, fmt.Errorf("failed to parse RPC response: %w - %s", err, string(body))
	}

	// Check for RPC error
	if rpcResp.Error != nil {
		return nil, fmt.Errorf("RPC error %d: %s", rpcResp.Error.Code, rpcResp.Error.Message)
	}

	return rpcResp.Result, nil
}
