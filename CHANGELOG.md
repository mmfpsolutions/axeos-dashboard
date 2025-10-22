# Changelog

All notable changes to the AxeOS Dashboard project will be documented in this file.

## [Unreleased]

### Added
- **Data Collection System** - Automated historical metrics collection
  - SQLite database for metrics storage (modernc.org/sqlite - pure Go, no CGO)
  - Singleton pattern for database and scheduler managers
  - Non-blocking goroutine-based collection tasks using `time.Ticker`
  - Three metric tables: `axeos_metrics`, `pool_metrics`, `node_metrics`
  - Configurable collection intervals and data retention periods
  - Volume mount at `/app/data` for metrics persistence

- **Centralized Logging System**
  - Standard log format: `[timestamp] [client_ip/system] [module] action`
  - Module-based logging: main, config, database, scheduler, middleware, service, auth
  - Automatic client IP extraction for HTTP requests
  - Methods: `Info()`, `Error()`, `Warn()`, `Fatal()`, `Debug()`, plus `*WithRequest()` variants
  - Easy filtering by timestamp, IP, module, or action

### Configuration
- New `data_collection_enabled` field (default: `false`)
- New `collection_interval_seconds` field (default: `300` = 5 minutes)
- New `data_retention_days` field (default: `30` days)
- New `rpc_config` field for cryptocurrency node RPC credentials

### Technical Improvements
- Pure Go implementation (no CGO dependencies)
- Cross-platform compatibility maintained (Alpine/Debian, AMD64/ARM64)
- Thread-safe singleton patterns for database and scheduler
- Graceful shutdown handling for data collection tasks
- SQLite WAL mode for better concurrent performance

### Documentation
- Added comprehensive logging section to README.md
- Updated data collection documentation with SQLite details
- Enhanced CLAUDE.md with logging and data collection architecture
- Added project structure updates reflecting new packages

## [Previous Versions]

See git history for previous changes before formal changelog was introduced.
