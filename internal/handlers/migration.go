package handlers

import (
	"encoding/json"
	"net/http"
)

// MigrationStatusResponse represents the response for migration status
type MigrationStatusResponse struct {
	Success bool                   `json:"success"`
	Data    map[string]interface{} `json:"data,omitempty"`
	Message string                 `json:"message,omitempty"`
}

// HandleMigrationStatus handles GET /api/migration/status
// Returns empty migration status (Go version doesn't need migration)
func HandleMigrationStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response := MigrationStatusResponse{
			Success: false,
			Message: "Method not allowed",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Go version doesn't have migration status - return empty/no migration needed
	response := MigrationStatusResponse{
		Success: true,
		Data: map[string]interface{}{
			"migrated": false,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// HandleMigrationClear handles POST /api/migration/clear
// No-op for Go version but returns success for compatibility
func HandleMigrationClear(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response := MigrationStatusResponse{
			Success: false,
			Message: "Method not allowed",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(response)
		return
	}

	// No-op for Go version - just return success
	response := MigrationStatusResponse{
		Success: true,
		Message: "Migration status cleared",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
