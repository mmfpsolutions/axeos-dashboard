package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/scottwalter/axeos-dashboard/internal/config"
	"github.com/scottwalter/axeos-dashboard/internal/services"
)

// StatisticsResponse represents the response structure for statistics endpoint
type StatisticsResponse struct {
	Success     bool        `json:"success"`
	InstanceID  string      `json:"instanceId,omitempty"`
	InstanceURL string      `json:"instanceUrl,omitempty"`
	Data        interface{} `json:"data,omitempty"`
	Message     string      `json:"message,omitempty"`
}

// HandleStatistics handles GET /api/statistics?instanceId=X
// Proxies the request to the BitAxe instance's /api/system/statistics/dashboard endpoint
func HandleStatistics(w http.ResponseWriter, r *http.Request, cfgManager *config.Manager) {
	cfg := cfgManager.GetConfig() // Get fresh config for hot reload
	// Only allow GET requests
	if r.Method != http.MethodGet {
		response := StatisticsResponse{
			Success: false,
			Message: "Method not allowed",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Get instanceId from query parameters
	instanceID := r.URL.Query().Get("instanceId")
	if instanceID == "" {
		response := StatisticsResponse{
			Success: false,
			Message: "instanceId parameter is required",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Find the instance configuration
	var instanceURL string
	for _, instance := range cfg.BitaxeInstances {
		if url, ok := instance[instanceID]; ok {
			instanceURL = url
			break
		}
	}

	if instanceURL == "" {
		response := StatisticsResponse{
			Success: false,
			Message: fmt.Sprintf("Instance '%s' not found in configuration", instanceID),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Get the statistics endpoint path from API map
	statisticsPath := services.GetAPIPath(cfg, "statisticsDashboard")
	if statisticsPath == "" {
		statisticsPath = "/api/system/statistics/dashboard" // Default fallback
	}

	statisticsURL := fmt.Sprintf("%s%s", instanceURL, statisticsPath)

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Fetch statistics from the BitAxe instance
	resp, err := client.Get(statisticsURL)
	if err != nil {
		log.Printf("Failed to fetch statistics for %s: %v", instanceID, err)
		response := StatisticsResponse{
			Success:    false,
			Message:    fmt.Sprintf("Failed to fetch statistics from %s: %v", instanceID, err),
			InstanceID: instanceID,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("Statistics endpoint returned non-OK status %d: %s", resp.StatusCode, string(body))
		response := StatisticsResponse{
			Success:    false,
			Message:    fmt.Sprintf("HTTP %d: %s", resp.StatusCode, resp.Status),
			InstanceID: instanceID,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Read and parse the statistics data
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read statistics response: %v", err)
		response := StatisticsResponse{
			Success:    false,
			Message:    fmt.Sprintf("Failed to read statistics response: %v", err),
			InstanceID: instanceID,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Parse the statistics data
	var statisticsData interface{}
	if err := json.Unmarshal(body, &statisticsData); err != nil {
		log.Printf("Failed to parse statistics JSON: %v", err)
		response := StatisticsResponse{
			Success:    false,
			Message:    fmt.Sprintf("Failed to parse statistics data: %v", err),
			InstanceID: instanceID,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Create enriched response with metadata
	response := StatisticsResponse{
		Success:     true,
		InstanceID:  instanceID,
		InstanceURL: instanceURL,
		Data:        statisticsData,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
