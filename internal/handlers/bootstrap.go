package handlers

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// BitaxeInstance represents a single Bitaxe device
type BitaxeInstance struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// MiningCoreInstance represents a single Mining Core instance
type MiningCoreInstance struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// BootstrapRequest represents the bootstrap form submission
type BootstrapRequest struct {
	// Basic Settings
	Title string `json:"title"`
	Port  string `json:"port"` // Port comes as string from frontend

	// Authentication Settings
	EnableAuth      string `json:"enableAuth"` // Comes as "true"/"false" string
	Username        string `json:"username"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirmPassword"`
	JWTKey          string `json:"jwtKey"`
	JWTExpiry       string `json:"jwtExpiry"`

	// Bitaxe Devices
	BitaxeInstances []BitaxeInstance `json:"bitaxeInstances"`

	// Mining Core Settings
	EnableMiningCore    string               `json:"enableMiningCore"` // Comes as "true"/"false" string
	MiningCoreInstances []MiningCoreInstance `json:"miningCoreInstances"`

	// Crypto Node Settings
	EnableCryptoNode   string `json:"enableCryptoNode"` // Comes as "true"/"false" string
	CryptoNodeType     string `json:"cryptoNodeType"`
	CryptoNodeName     string `json:"cryptoNodeName"`
	CryptoNodeAlgo     string `json:"cryptoNodeAlgo"`
	CryptoNodeId       string `json:"cryptoNodeId"`
	CryptoNodeRpcIp    string `json:"cryptoNodeRpcIp"`
	CryptoNodeRpcPort  string `json:"cryptoNodeRpcPort"` // Port comes as string
	CryptoNodeRpcAuth  string `json:"cryptoNodeRpcAuth"`
}

// HandleBootstrapPage serves the bootstrap HTML page
func HandleBootstrapPage(w http.ResponseWriter, r *http.Request) {
	// Determine publicDir from the current working directory
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting current working directory: %v\n", err)
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "<h1>Error</h1><p>Failed to determine public directory</p>")
		return
	}

	publicDir := filepath.Join(cwd, "public")
	bootstrapHTMLPath := filepath.Join(publicDir, "html", "bootstrap.html")

	htmlContent, err := os.ReadFile(bootstrapHTMLPath)
	if err != nil {
		fmt.Printf("Error reading bootstrap.html: %v\n", err)
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "<h1>Error</h1><p>Failed to load bootstrap page: %v</p>", err)
		return
	}

	html := string(htmlContent)

	// Replace placeholders
	title := "AxeOS Dashboard"
	version := "1.0"
	currentYear := fmt.Sprintf("%d", time.Now().Year())

	html = strings.ReplaceAll(html, "<!-- TITLE -->", title)
	html = strings.ReplaceAll(html, "<!-- VERSION -->", version)
	html = strings.ReplaceAll(html, "<!-- CURRENT_YEAR -->", currentYear)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

// HandleBootstrapSubmit processes the bootstrap form submission
func HandleBootstrapSubmit(configDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(map[string]string{"message": "Method Not Allowed"})
			return
		}

		// Read request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"message": "Invalid request body"})
			return
		}
		defer r.Body.Close()

		// Parse bootstrap request
		var req BootstrapRequest
		if err := json.Unmarshal(body, &req); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"message": "Invalid request format"})
			return
		}

		// Validate basic settings
		if req.Title == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"message": "Title is required"})
			return
		}

		// Validate authentication settings if enabled
		enableAuth := req.EnableAuth == "true"
		if enableAuth {
			if req.Username == "" || req.Password == "" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{"message": "Username and password are required when authentication is enabled"})
				return
			}
			if len(req.JWTKey) != 32 {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{"message": "JWT key must be 32 characters"})
				return
			}
		}

		// Validate at least one Bitaxe device
		if len(req.BitaxeInstances) == 0 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"message": "At least one Bitaxe device is required"})
			return
		}

		// Create config directory if it doesn't exist
		if err := os.MkdirAll(configDir, 0755); err != nil {
			fmt.Printf("Error creating config directory: %v\n", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"message": "Failed to create config directory"})
			return
		}

		// Create config.json
		cfg := createConfig(req)
		if err := saveConfigJSON(configDir, cfg); err != nil {
			fmt.Printf("Error saving config.json: %v\n", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"message": "Failed to save configuration"})
			return
		}

		// Create access.json if authentication is enabled
		if enableAuth {
			if err := saveAccessJSON(configDir, req.Username, req.Password); err != nil {
				fmt.Printf("Error saving access.json: %v\n", err)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{"message": "Failed to save access credentials"})
				return
			}

			// Create jsonWebTokenKey.json
			if err := saveJWTKeyJSON(configDir, req.JWTKey); err != nil {
				fmt.Printf("Error saving jsonWebTokenKey.json: %v\n", err)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{"message": "Failed to save JWT key"})
				return
			}
		} else {
			// Create empty access.json and jsonWebTokenKey.json files for non-auth mode
			if err := saveAccessJSON(configDir, "", ""); err != nil {
				fmt.Printf("Error saving empty access.json: %v\n", err)
			}
			// Generate a random JWT key even if auth is disabled (for potential future use)
			randomKey := generateRandomKey(32)
			if err := saveJWTKeyJSON(configDir, randomKey); err != nil {
				fmt.Printf("Error saving jsonWebTokenKey.json: %v\n", err)
			}
		}

		// Create rpcConfig.json if crypto node is enabled
		enableCryptoNode := req.EnableCryptoNode == "true"
		if enableCryptoNode {
			if err := saveRPCConfigJSON(configDir, req); err != nil {
				fmt.Printf("Error saving rpcConfig.json: %v\n", err)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{"message": "Failed to save RPC configuration"})
				return
			}
		}

		// Success - configuration files created
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Configuration created successfully! Redirecting to dashboard...",
		})
	}
}

// createConfig creates a Config struct from the bootstrap request
func createConfig(req BootstrapRequest) map[string]interface{} {
	// Parse port string to int
	port := 3000 // default
	if portInt, err := parsePort(req.Port); err == nil {
		port = portInt
	}

	enableAuth := req.EnableAuth == "true"
	enableMiningCore := req.EnableMiningCore == "true"
	enableCryptoNode := req.EnableCryptoNode == "true"

	// Create config as map to preserve exact JSON structure
	cfg := map[string]interface{}{
		"bitaxe_dashboard_version": 3.0,
		"disable_authentication":   !enableAuth,
		"cookie_max_age":           3600,
		"disable_settings":         false,
		"disable_configurations":   false,
		"web_server_port":          port,
		"title":                    req.Title,
		"bitaxe_instances":         []map[string]string{},
		"display_fields":           getDefaultDisplayFields(),
		"mining_core_enabled":      enableMiningCore,
		"mining_core_url":          []map[string]string{},
		"mining_core_display_fields": getDefaultMiningCoreDisplayFields(),
		"cryptNodesEnabled":        enableCryptoNode,
		"cryptoNodes":              nil,
		"configuration_outdated":   false,
		"bitaxe_api":               nil,
	}

	// Add Bitaxe devices
	bitaxeInstances := []map[string]string{}
	for _, device := range req.BitaxeInstances {
		if device.Name != "" && device.URL != "" {
			bitaxeInstances = append(bitaxeInstances, map[string]string{
				device.Name: device.URL,
			})
		}
	}
	cfg["bitaxe_instances"] = bitaxeInstances

	// Add Mining Core instances if enabled
	if enableMiningCore && req.MiningCoreInstances != nil {
		miningCoreURLs := []map[string]string{}
		for _, mc := range req.MiningCoreInstances {
			if mc.Name != "" && mc.URL != "" {
				miningCoreURLs = append(miningCoreURLs, map[string]string{
					mc.Name: mc.URL,
				})
			}
		}
		cfg["mining_core_url"] = miningCoreURLs
	}

	// Set JWT expiry (convert to seconds)
	if enableAuth && req.JWTExpiry != "" {
		cfg["cookie_max_age"] = parseJWTExpiry(req.JWTExpiry)
	}

	// Add crypto nodes if enabled
	if enableCryptoNode {
		cfg["cryptoNodes"] = []map[string]interface{}{
			{
				"NodeType":  req.CryptoNodeType,
				"NodeName":  req.CryptoNodeName,
				"NodeId":    req.CryptoNodeId,
				"NodeAlgo":  req.CryptoNodeAlgo,
				"NodeDisplayFields": getCryptoNodeDisplayFields(req.CryptoNodeType),
			},
		}
	}

	return cfg
}

// parsePort safely parses port string to int
func parsePort(portStr string) (int, error) {
	if portStr == "" {
		return 3000, nil
	}
	var port int
	_, err := fmt.Sscanf(portStr, "%d", &port)
	if err != nil {
		return 0, err
	}
	if port < 1024 || port > 65523 {
		return 0, fmt.Errorf("port out of range")
	}
	return port, nil
}

// parseJWTExpiry converts expiry string (1h, 8h, 24h, 7d) to seconds
func parseJWTExpiry(expiry string) int {
	switch expiry {
	case "1h":
		return 3600
	case "8h":
		return 28800
	case "24h":
		return 86400
	case "7d":
		return 604800
	default:
		return 3600 // Default 1 hour
	}
}

// saveConfigJSON saves the config to config.json
func saveConfigJSON(configDir string, cfg map[string]interface{}) error {
	configPath := filepath.Join(configDir, "config.json")
	data, err := json.MarshalIndent(cfg, "", "    ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, data, 0644)
}

// saveAccessJSON saves the access credentials to access.json
func saveAccessJSON(configDir, username, password string) error {
	accessPath := filepath.Join(configDir, "access.json")

	// If no username/password, create empty object
	if username == "" || password == "" {
		emptyData := []byte("{}")
		return os.WriteFile(accessPath, emptyData, 0644)
	}

	// Hash the password with SHA256
	hasher := sha256.New()
	hasher.Write([]byte(password))
	hashedPassword := hex.EncodeToString(hasher.Sum(nil))

	// Create access data
	accessData := map[string]string{
		username: hashedPassword,
	}

	data, err := json.MarshalIndent(accessData, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(accessPath, data, 0644)
}

// saveJWTKeyJSON saves the JWT key to jsonWebTokenKey.json
func saveJWTKeyJSON(configDir, jwtKey string) error {
	jwtKeyPath := filepath.Join(configDir, "jsonWebTokenKey.json")
	jwtData := map[string]string{
		"jsonWebTokenKey": jwtKey,
		"expiresIn":       "1h",
	}
	data, err := json.MarshalIndent(jwtData, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(jwtKeyPath, data, 0644)
}

// saveRPCConfigJSON saves the RPC configuration to rpcConfig.json
func saveRPCConfigJSON(configDir string, req BootstrapRequest) error {
	rpcConfigPath := filepath.Join(configDir, "rpcConfig.json")

	// Parse RPC port string to int
	rpcPort := 8332 // default
	if req.CryptoNodeRpcPort != "" {
		if port, err := parsePort(req.CryptoNodeRpcPort); err == nil {
			rpcPort = port
		}
	}

	rpcConfig := map[string]interface{}{
		"cryptoNodes": []map[string]interface{}{
			{
				"type":    req.CryptoNodeType,
				"name":    req.CryptoNodeName,
				"algo":    req.CryptoNodeAlgo,
				"id":      req.CryptoNodeId,
				"rpcIp":   req.CryptoNodeRpcIp,
				"rpcPort": rpcPort,
				"rpcAuth": req.CryptoNodeRpcAuth,
			},
		},
	}

	data, err := json.MarshalIndent(rpcConfig, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(rpcConfigPath, data, 0644)
}

// generateRandomKey generates a random hex string of specified length
func generateRandomKey(length int) string {
	bytes := make([]byte, length/2)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based key if random fails
		timestamp := time.Now().UnixNano()
		return fmt.Sprintf("%032x", timestamp)[:length]
	}
	return hex.EncodeToString(bytes)
}

// getDefaultDisplayFields returns the default display fields for Bitaxe instances
func getDefaultDisplayFields() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"Mining Metrics": []map[string]string{
				{"hashRate": "Hashrate"},
				{"expectedHashrate": "Expect Hashrate"},
				{"bestDiff": "Best Difficulty"},
				{"bestSessionDiff": "Best Session Difficulty"},
				{"poolDifficulty": "Pool Difficulty"},
				{"sharesAccepted": "Shares Accepted"},
				{"sharesRejected": "Shares Rejected"},
				{"sharesRejectedReasons": "Shares Rejected Reasons"},
				{"responseTime": "Response Time"},
			},
		},
		{
			"General Information": []map[string]string{
				{"hostname": "Hostname"},
				{"power": "Power"},
				{"voltage": "Voltage"},
				{"coreVoltageActual": "ASIC Voltage"},
				{"frequency": "Frequency"},
				{"temp": "ASIC Temp"},
				{"vrTemp": "VR Temp"},
				{"fanspeed": "Fan Speed"},
				{"minFanSpeed": "Min Fan Speed"},
				{"fanrpm": "Fan RPM"},
				{"temptarget": "Target Temp"},
				{"overheat_mode": "Over Heat Mode"},
				{"uptimeSeconds": "Uptime"},
				{"coreVoltage": "Core Voltage"},
				{"current": "Current"},
				{"wifiRSSI": "Wifi RSSI"},
				{"stratumURL": "Stratum URL"},
				{"stratumUser": "Stratum User"},
				{"stratumPort": "Stratum Port"},
				{"isUsingFallbackStratum": "Using Fallback Stratum"},
				{"axeOSVersion": "AxeOS Version"},
				{"idfVersion": "IDF Version"},
				{"boardVersion": "Board Version"},
				{"ASICModel": "ASIC Chip"},
			},
		},
	}
}

// getDefaultMiningCoreDisplayFields returns the default display fields for Mining Core instances
func getDefaultMiningCoreDisplayFields() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"Network Status": []map[string]string{
				{"networkHashrate": "Network Hashrate"},
				{"networkDifficulty": "Network Difficulty"},
				{"lastNetworkBlockTime": "Last Block Time"},
				{"blockHeight": "Block Height"},
				{"connectedPeers": "Connected Peers"},
				{"nodeVersion": "Node Version"},
			},
		},
		{
			"Miner(s) Status": []map[string]string{
				{"connectedMiners": "Connected Miners"},
				{"poolHashrate": "Pool Hashrate"},
			},
		},
		{
			"Rewards Status": []map[string]string{
				{"totalPaid": "Total Paid"},
				{"totalBlocks": "Total Blocks"},
				{"totalConfirmedBlocks": "Total Confirmed Blocks"},
				{"totalPendingBlocks": "Totoal Pending Blocks"},
				{"lastPoolBlockTime": "Last Pool Block Time"},
				{"blockReward": "Block Reward"},
			},
		},
	}
}

// getCryptoNodeDisplayFields returns the default display fields for crypto nodes
func getCryptoNodeDisplayFields(nodeType string) []map[string]interface{} {
	// Return generic crypto node display fields (works for dgb, btc, etc.)
	return []map[string]interface{}{
		{
			"Block Chain Info": []map[string]string{
				{"chain": "Chain"},
				{"blocks": "Blocks"},
				{"headers": "Headers"},
				{"size_on_disk": "Size on Disk"},
				{"mediantime": "Median Time"},
				{"pruned": "Pruned"},
				{"verificationprogress": "Verification"},
				{"initialblockdownload": "Initializing"},
				{"warnings": "Warnings"},
				{"difficulties/sha256d": "Difficulty"},
			},
		},
		{
			"Network Info": []map[string]string{
				{"version": "Version"},
				{"subversion": "Subversion"},
				{"protocolversion": "Protocol"},
				{"networkactive": "Active"},
				{"warnings": "Warnings"},
				{"connections": "Connections"},
				{"connections_in": "In"},
				{"connections_out": "Out"},
			},
		},
		{
			"Network Totals": []map[string]string{
				{"target": "Target"},
				{"totalbytesrecv": "Received"},
				{"totalbytessent": "Sent"},
				{"bytes_left_in_cycle": "Bytes Left"},
				{"timemillis": "Updated"},
				{"target_reached": "Target Reached"},
				{"serve_historical_blocks": "Historicals"},
				{"timeframe": "Cycle Time"},
				{"time_left_in_cycle": "Time Left"},
			},
		},
		{
			"Wallet Info": []map[string]string{
				{"balance": "Balance"},
			},
		},
	}
}
