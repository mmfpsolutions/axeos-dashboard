package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/scottwalter/axeos-dashboard/internal/config"
	"github.com/scottwalter/axeos-dashboard/internal/services"
)

// MinerData represents data from a single miner instance
type MinerData struct {
	ID       string                 `json:"id"`
	Hostname string                 `json:"hostname,omitempty"`
	Status   string                 `json:"status,omitempty"`
	Message  string                 `json:"message,omitempty"`
	Data     map[string]interface{} `json:",inline"`
}

// MiningCoreInstanceData represents mining core instance data
type MiningCoreInstanceData struct {
	InstanceName string                   `json:"instanceName"`
	Status       string                   `json:"status"`
	Message      string                   `json:"message,omitempty"`
	Pools        []map[string]interface{} `json:"pools"`
}

// SystemsInfoResponse represents the aggregated response
type SystemsInfoResponse struct {
	MinerData                []map[string]interface{}  `json:"minerData"`
	DisplayFields            interface{}               `json:"displayFields"` // Can be []string or complex nested structure
	MiningCoreData           []MiningCoreInstanceData  `json:"miningCoreData"`
	MiningCoreDisplayFields  interface{}               `json:"miningCoreDisplayFields"` // Can be []string or complex nested structure
	CryptoNodeData           interface{}               `json:"cryptoNodeData"`
	DisableSettings          bool                      `json:"disable_settings"`
	DisableConfigurations    bool                      `json:"disable_configurations"`
	DisableAuthentication    bool                      `json:"disable_authentication"`
	MiningCoreEnabled        bool                      `json:"mining_core_enabled"`
}

// HandleSystemsInfo handles GET /api/systems/info
func HandleSystemsInfo(cfgManager *config.Manager, cryptoNodeSvc *services.CryptoNodeService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg := cfgManager.GetConfig() // Get fresh config for hot reload
		apiPath := services.GetAPIPath(cfg, "instanceInfo")
		allMinerData := []map[string]interface{}{}

		// Fetch data from all AxeOS instances concurrently
		var wg sync.WaitGroup
		minerChan := make(chan map[string]interface{}, len(cfg.AxeosInstances))

		for _, instance := range cfg.AxeosInstances {
			for instanceName, instanceURL := range instance {
				wg.Add(1)
				go func(name, url string) {
					defer wg.Done()

					resp, err := http.Get(url + apiPath)
					if err != nil {
						fmt.Printf("Network or JSON parsing error for %s (%s): %v\n", name, url, err)
						minerChan <- map[string]interface{}{
							"id":       name,
							"hostname": name,
							"status":   "Error",
							"message":  err.Error(),
						}
						return
					}
					defer resp.Body.Close()

					if resp.StatusCode != http.StatusOK {
						fmt.Printf("Error fetching data from %s: %d %s\n", url, resp.StatusCode, resp.Status)
						minerChan <- map[string]interface{}{
							"id":       name,
							"hostname": name,
							"status":   "Error",
							"message":  fmt.Sprintf("%d %s", resp.StatusCode, resp.Status),
						}
						return
					}

					var data map[string]interface{}
					if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
						fmt.Printf("JSON parsing error for %s: %v\n", name, err)
						minerChan <- map[string]interface{}{
							"id":       name,
							"hostname": name,
							"status":   "Error",
							"message":  err.Error(),
						}
						return
					}

					data["id"] = name
					minerChan <- data
				}(instanceName, instanceURL)
			}
		}

		// Wait for all miner fetches to complete
		go func() {
			wg.Wait()
			close(minerChan)
		}()

		// Collect miner data
		for data := range minerChan {
			allMinerData = append(allMinerData, data)
		}

		// Prepare response
		response := SystemsInfoResponse{
			MinerData:               allMinerData,
			DisplayFields:           cfg.DisplayFields,
			MiningCoreData:          []MiningCoreInstanceData{},
			MiningCoreDisplayFields: cfg.MiningCoreDisplayFields,
			CryptoNodeData:          nil,
			DisableSettings:         cfg.DisableSettings,
			DisableConfigurations:   cfg.DisableConfigurations,
			DisableAuthentication:   cfg.DisableAuthentication,
			MiningCoreEnabled:       cfg.MiningCoreEnabled,
		}

		// Fetch mining core data if enabled
		if cfg.MiningCoreEnabled && len(cfg.MiningCoreURL) > 0 {
			miningCoreAPIPath := services.GetAPIPath(cfg, "pools")
			var mcWg sync.WaitGroup
			mcChan := make(chan MiningCoreInstanceData, len(cfg.MiningCoreURL))

			for _, instance := range cfg.MiningCoreURL {
				for instanceName, instanceURL := range instance {
					mcWg.Add(1)
					go func(name, url string) {
						defer mcWg.Done()

						resp, err := http.Get(url + miningCoreAPIPath)
						if err != nil {
							fmt.Printf("Network error for mining core %s (%s): %v\n", name, url, err)
							mcChan <- MiningCoreInstanceData{
								InstanceName: name,
								Status:       "Error",
								Message:      err.Error(),
								Pools:        []map[string]interface{}{},
							}
							return
						}
						defer resp.Body.Close()

						if resp.StatusCode != http.StatusOK {
							fmt.Printf("Error fetching mining core data from %s: %d %s\n", url, resp.StatusCode, resp.Status)
							mcChan <- MiningCoreInstanceData{
								InstanceName: name,
								Status:       "Error",
								Message:      fmt.Sprintf("%d %s", resp.StatusCode, resp.Status),
								Pools:        []map[string]interface{}{},
							}
							return
						}

						var mcData map[string]interface{}
						if err := json.NewDecoder(resp.Body).Decode(&mcData); err != nil {
							fmt.Printf("JSON parsing error for mining core %s: %v\n", name, err)
							mcChan <- MiningCoreInstanceData{
								InstanceName: name,
								Status:       "Error",
								Message:      err.Error(),
								Pools:        []map[string]interface{}{},
							}
							return
						}

						pools := []map[string]interface{}{}
						if poolsData, ok := mcData["pools"].([]interface{}); ok {
							for _, pool := range poolsData {
								if poolMap, ok := pool.(map[string]interface{}); ok {
									pools = append(pools, poolMap)
								}
							}
						}

						mcChan <- MiningCoreInstanceData{
							InstanceName: name,
							Status:       "OK",
							Pools:        pools,
						}
					}(instanceName, instanceURL)
				}
			}

			go func() {
				mcWg.Wait()
				close(mcChan)
			}()

			for data := range mcChan {
				response.MiningCoreData = append(response.MiningCoreData, data)
			}
		}

		// Fetch crypto node data if enabled
		if cfg.CryptNodesEnabled && cryptoNodeSvc != nil {
			cryptoNodeData, err := cryptoNodeSvc.FetchAllCryptoNodes(cfg)
			if err != nil {
				fmt.Printf("Error fetching crypto node data: %v\n", err)
				response.CryptoNodeData = []interface{}{}
			} else {
				response.CryptoNodeData = cryptoNodeData
			}
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		w.WriteHeader(http.StatusOK)
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		encoder.Encode(response)
	}
}
