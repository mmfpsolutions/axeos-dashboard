package services

import "github.com/scottwalter/axeos-dashboard/internal/config"

// GetAPIPath returns the API path for the given endpoint type
func GetAPIPath(cfg *config.Config, endpointType string) string {
	if cfg.AxeosAPI == nil {
		// Return default paths
		switch endpointType {
		case "instanceInfo":
			return "/api/system/info"
		case "instanceRestart":
			return "/api/system/restart"
		case "instanceSettings":
			return "/api/system"
		case "pools":
			return "/api/pools"
		default:
			return ""
		}
	}

	// Return configured path or default
	if path, ok := cfg.AxeosAPI[endpointType]; ok {
		return path
	}

	// Fallback to defaults
	switch endpointType {
	case "instanceInfo":
		return "/api/system/info"
	case "instanceRestart":
		return "/api/system/restart"
	case "instanceSettings":
		return "/api/system"
	case "pools":
		return "/api/pools"
	default:
		return ""
	}
}
