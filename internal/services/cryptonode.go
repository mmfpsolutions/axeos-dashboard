package services

import (
	"fmt"
	"log"
	"sync"

	"github.com/scottwalter/axeos-dashboard/internal/config"
)

// CryptoNodeService handles crypto node interactions
type CryptoNodeService struct {
	configDir string
	rpcClient *RPCClient
}

// NodeData represents the aggregated data for a single crypto node
type NodeData struct {
	ID             string      `json:"id"`
	NodeID         string      `json:"nodeId"`
	NodeType       string      `json:"nodeType"`
	NodeAlgo       string      `json:"nodeAlgo,omitempty"`
	Status         string      `json:"status"`
	Message        string      `json:"message,omitempty"`
	BlockchainInfo interface{} `json:"blockchainInfo,omitempty"`
	NetworkTotals  interface{} `json:"networkTotals,omitempty"`
	Balance        interface{} `json:"balance,omitempty"`
	NetworkInfo    interface{} `json:"networkInfo,omitempty"`
	DisplayFields  interface{} `json:"displayFields,omitempty"`
}

// NodeConfig represents a node configuration from config.json
type NodeConfig struct {
	NodeType   string `json:"NodeType"`
	NodeName   string `json:"NodeName"`
	NodeID     string `json:"NodeId"`
	NodeAlgo   string `json:"NodeAlgo"`
}

// NewCryptoNodeService creates a new crypto node service
func NewCryptoNodeService(configDir string) *CryptoNodeService {
	return &CryptoNodeService{
		configDir: configDir,
		rpcClient: NewRPCClient(configDir),
	}
}

// getBlockchainInfo fetches blockchain info from a crypto node
func (c *CryptoNodeService) getBlockchainInfo(nodeID string) (interface{}, error) {
	result, err := c.rpcClient.CallRPC(nodeID, "getblockchaininfo", []interface{}{})
	if err != nil {
		return nil, fmt.Errorf("error fetching blockchain info for %s: %w", nodeID, err)
	}
	return result, nil
}

// getNetworkTotals fetches network totals from a crypto node
func (c *CryptoNodeService) getNetworkTotals(nodeID string) (interface{}, error) {
	result, err := c.rpcClient.CallRPC(nodeID, "getnettotals", []interface{}{})
	if err != nil {
		return nil, fmt.Errorf("error fetching network totals for %s: %w", nodeID, err)
	}
	return result, nil
}

// getBalance fetches wallet balance from a crypto node
func (c *CryptoNodeService) getBalance(nodeID string) (interface{}, error) {
	result, err := c.rpcClient.CallRPC(nodeID, "getbalance", []interface{}{})
	if err != nil {
		return nil, fmt.Errorf("error fetching balance for %s: %w", nodeID, err)
	}
	return result, nil
}

// getNetworkInfo fetches network info from a crypto node
func (c *CryptoNodeService) getNetworkInfo(nodeID string) (interface{}, error) {
	result, err := c.rpcClient.CallRPC(nodeID, "getnetworkinfo", []interface{}{})
	if err != nil {
		return nil, fmt.Errorf("error fetching network info for %s: %w", nodeID, err)
	}
	return result, nil
}

// fetchCryptoNodeData aggregates all crypto node data for a single node
func (c *CryptoNodeService) fetchCryptoNodeData(nodeConfig NodeConfig, displayFields interface{}) NodeData {
	nodeID := nodeConfig.NodeID

	// Fetch all data concurrently using goroutines
	var wg sync.WaitGroup
	var blockchainInfo, networkTotals, balance, networkInfo interface{}
	var bcErr, ntErr, balErr, niErr error

	wg.Add(4)

	go func() {
		defer wg.Done()
		blockchainInfo, bcErr = c.getBlockchainInfo(nodeID)
	}()

	go func() {
		defer wg.Done()
		networkTotals, ntErr = c.getNetworkTotals(nodeID)
	}()

	go func() {
		defer wg.Done()
		balance, balErr = c.getBalance(nodeID)
	}()

	go func() {
		defer wg.Done()
		networkInfo, niErr = c.getNetworkInfo(nodeID)
	}()

	wg.Wait()

	// Check if any errors occurred
	if bcErr != nil || ntErr != nil || balErr != nil || niErr != nil {
		errMsg := ""
		if bcErr != nil {
			errMsg += bcErr.Error() + "; "
		}
		if ntErr != nil {
			errMsg += ntErr.Error() + "; "
		}
		if balErr != nil {
			errMsg += balErr.Error() + "; "
		}
		if niErr != nil {
			errMsg += niErr.Error()
		}

		log.Printf("Failed to fetch data for node %s: %s", nodeID, errMsg)

		// Return error object for this node
		nodeName := nodeConfig.NodeName
		if nodeName == "" {
			nodeName = nodeID
		}

		return NodeData{
			ID:       nodeName,
			NodeID:   nodeID,
			NodeType: nodeConfig.NodeType,
			Status:   "Error",
			Message:  errMsg,
		}
	}

	// Combine all data into a single object
	nodeName := nodeConfig.NodeName
	if nodeName == "" {
		nodeName = nodeID
	}

	return NodeData{
		ID:             nodeName,
		NodeID:         nodeID,
		NodeType:       nodeConfig.NodeType,
		NodeAlgo:       nodeConfig.NodeAlgo,
		Status:         "online",
		BlockchainInfo: blockchainInfo,
		NetworkTotals:  networkTotals,
		Balance:        balance,
		NetworkInfo:    networkInfo,
		DisplayFields:  displayFields,
	}
}

// FetchAllCryptoNodes fetches data from all configured crypto nodes
func (c *CryptoNodeService) FetchAllCryptoNodes(cfg *config.Config) (interface{}, error) {
	// Check if crypto nodes are enabled
	if !cfg.CryptNodesEnabled {
		return []interface{}{}, nil
	}

	// Parse the cryptoNodes configuration structure
	cryptoNodes, ok := cfg.CryptoNodes.([]interface{})
	if !ok || len(cryptoNodes) == 0 {
		return []interface{}{}, nil
	}

	// Find the Nodes and NodeDisplayFields in the cryptoNodes array
	var nodes []NodeConfig
	var displayFields interface{}

	for _, item := range cryptoNodes {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		// Check for Nodes array
		if nodesRaw, exists := itemMap["Nodes"]; exists {
			if nodesArray, ok := nodesRaw.([]interface{}); ok {
				for _, nodeRaw := range nodesArray {
					if nodeMap, ok := nodeRaw.(map[string]interface{}); ok {
						node := NodeConfig{}
						if nt, ok := nodeMap["NodeType"].(string); ok {
							node.NodeType = nt
						}
						if nn, ok := nodeMap["NodeName"].(string); ok {
							node.NodeName = nn
						}
						if nid, ok := nodeMap["NodeId"].(string); ok {
							node.NodeID = nid
						}
						if na, ok := nodeMap["NodeAlgo"].(string); ok {
							node.NodeAlgo = na
						}
						nodes = append(nodes, node)
					}
				}
			}
		}

		// Check for NodeDisplayFields
		if ndf, exists := itemMap["NodeDisplayFields"]; exists {
			displayFields = ndf
		}
	}

	// If nodes array is empty, return empty array
	if len(nodes) == 0 {
		return []interface{}{}, nil
	}

	// Fetch data for all nodes concurrently
	var wg sync.WaitGroup
	nodeDataChan := make(chan NodeData, len(nodes))

	for _, nodeConfig := range nodes {
		wg.Add(1)
		go func(nc NodeConfig) {
			defer wg.Done()
			nodeData := c.fetchCryptoNodeData(nc, displayFields)
			nodeDataChan <- nodeData
		}(nodeConfig)
	}

	wg.Wait()
	close(nodeDataChan)

	// Collect all node data
	var result []interface{}
	for nodeData := range nodeDataChan {
		result = append(result, nodeData)
	}

	return result, nil
}
