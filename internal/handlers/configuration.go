package handlers

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/scottwalter/axeos-dashboard/internal/config"
)

// HandleConfiguration handles GET and PATCH /api/configuration
func HandleConfiguration(cfgManager *config.Manager, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check if configurations are disabled
		if cfg.DisableConfigurations {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]string{"message": "Configurations are disabled by configuration."})
			return
		}

		switch r.Method {
		case http.MethodGet:
			handleGetConfiguration(w, r, cfgManager)
		case http.MethodPatch:
			handleUpdateConfiguration(w, r, cfgManager)
		default:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(map[string]string{
				"status":  "error",
				"message": "Method " + r.Method + " not allowed",
			})
		}
	}
}

func handleGetConfiguration(w http.ResponseWriter, r *http.Request, cfgManager *config.Manager) {
	currentConfig := cfgManager.GetConfig()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data":   currentConfig,
	})
}

func handleUpdateConfiguration(w http.ResponseWriter, r *http.Request, cfgManager *config.Manager) {
	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": "Failed to read request body",
		})
		return
	}
	defer r.Body.Close()

	if len(body) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": "Request body is empty. Please provide configuration settings to update.",
		})
		return
	}

	// Parse JSON
	var updates map[string]interface{}
	if err := json.Unmarshal(body, &updates); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": "Invalid JSON in request body: " + err.Error(),
		})
		return
	}

	// Update configuration
	if err := cfgManager.UpdateConfig(updates); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	// Get updated config
	updatedConfig := cfgManager.GetConfig()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Configuration updated successfully! Changes have been applied immediately.",
		"data":    updatedConfig,
	})
}
