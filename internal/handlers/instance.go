package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/scottwalter/axeos-dashboard/internal/config"
	"github.com/scottwalter/axeos-dashboard/internal/services"
)

// HandleInstanceInfo handles GET /api/instance/info?instanceId=X
func HandleInstanceInfo(cfgManager *config.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg := cfgManager.GetConfig() // Get fresh config for hot reload
		if r.Method != http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(map[string]string{
				"error":   "Method Not Allowed",
				"message": "This endpoint only accepts GET requests.",
			})
			return
		}

		instanceID := r.URL.Query().Get("instanceId")
		if instanceID == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error":   "Bad Request",
				"message": "Missing \"instanceId\" query parameter.",
			})
			return
		}

		// Find the instance
		var instanceURL string
		for _, instance := range cfg.BitaxeInstances {
			if url, ok := instance[instanceID]; ok {
				instanceURL = url
				break
			}
		}

		if instanceURL == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{
				"error":   "Not Found",
				"message": fmt.Sprintf("Bitaxe instance \"%s\" not found in configuration.", instanceID),
			})
			return
		}

		// Get API path
		apiPath := services.GetAPIPath(cfg, "instanceInfo")
		infoURL := instanceURL + apiPath

		// Fetch data from the Bitaxe device
		resp, err := http.Get(infoURL)
		if err != nil {
			fmt.Printf("Error fetching from Bitaxe instance: %v\n", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"error":   "Internal Server Error",
				"message": err.Error(),
			})
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			errorText, _ := io.ReadAll(resp.Body)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(resp.StatusCode)
			json.NewEncoder(w).Encode(map[string]string{
				"error":   "Failed to fetch data from Bitaxe instance",
				"message": fmt.Sprintf("HTTP error! Status: %d, Body: %s", resp.StatusCode, string(errorText)),
			})
			return
		}

		// Forward the response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		io.Copy(w, resp.Body)
	}
}

// HandleInstanceRestart handles POST /api/instance/service/restart?instanceId=X
func HandleInstanceRestart(cfgManager *config.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg := cfgManager.GetConfig() // Get fresh config for hot reload
		if cfg.DisableSettings {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]string{"message": "Settings are disabled by configuration."})
			return
		}

		if r.Method != http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(map[string]string{"message": "Method Not Allowed"})
			return
		}

		instanceID := r.URL.Query().Get("instanceId")
		if instanceID == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"message": "Missing instanceId parameter"})
			return
		}

		// Find the instance
		var instanceURL string
		for _, instance := range cfg.BitaxeInstances {
			if url, ok := instance[instanceID]; ok {
				instanceURL = url
				break
			}
		}

		if instanceURL == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{
				"message": fmt.Sprintf("Bitaxe instance \"%s\" not found in configuration.", instanceID),
			})
			return
		}

		// Get API path and make request
		apiPath := services.GetAPIPath(cfg, "instanceRestart")
		restartURL := instanceURL + apiPath

		resp, err := http.Post(restartURL, "application/json", nil)
		if err != nil {
			fmt.Printf("Failed to restart Bitaxe: %v\n", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"message": "Internal Server Error", "error": err.Error()})
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			errorText, _ := io.ReadAll(resp.Body)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(resp.StatusCode)
			json.NewEncoder(w).Encode(map[string]string{
				"message": fmt.Sprintf("HTTP error! Status: %d, Body: %s", resp.StatusCode, string(errorText)),
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "success",
			"message": fmt.Sprintf("Restart initiated for %s", instanceID),
		})
	}
}

// HandleInstanceSettings handles PATCH /api/instance/service/settings?instanceId=X
func HandleInstanceSettings(cfgManager *config.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg := cfgManager.GetConfig() // Get fresh config for hot reload
		if cfg.DisableSettings {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]string{"message": "Settings are disabled by configuration."})
			return
		}

		if r.Method != http.MethodPatch {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(map[string]string{"message": "Method Not Allowed"})
			return
		}

		instanceID := r.URL.Query().Get("instanceId")
		if instanceID == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"message": "Missing instanceId parameter"})
			return
		}

		// Find the instance
		var instanceURL string
		for _, instance := range cfg.BitaxeInstances {
			if url, ok := instance[instanceID]; ok {
				instanceURL = url
				break
			}
		}

		if instanceURL == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{
				"message": fmt.Sprintf("Bitaxe instance \"%s\" not found in configuration.", instanceID),
			})
			return
		}

		// Read request body
		body, err := io.ReadAll(r.Body)
		if err != nil || len(body) == 0 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"message": "Request body cannot be empty."})
			return
		}
		defer r.Body.Close()

		// Validate JSON
		var testJSON map[string]interface{}
		if err := json.Unmarshal(body, &testJSON); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"message": "Invalid JSON in request body"})
			return
		}

		// Get API path and make request
		apiPath := services.GetAPIPath(cfg, "instanceSettings")
		settingsURL := instanceURL + apiPath

		req, err := http.NewRequest(http.MethodPatch, settingsURL, bytes.NewBuffer(body))
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"message": "Internal Server Error", "error": err.Error()})
			return
		}
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("Failed to update settings: %v\n", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"message": "Internal Server Error", "error": err.Error()})
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			errorText, _ := io.ReadAll(resp.Body)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(resp.StatusCode)
			json.NewEncoder(w).Encode(map[string]string{
				"message": fmt.Sprintf("HTTP error! Status: %d, Body: %s", resp.StatusCode, string(errorText)),
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "success",
			"message": fmt.Sprintf("Settings updated for %s", instanceID),
		})
	}
}
