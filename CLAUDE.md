# AxeOS Dashboard - Claude AI Context

This document contains important context and preferences for working on the AxeOS Dashboard project.

## Project Overview

**AxeOS Dashboard** is a complete rewrite of the Bitaxe Dashboard, migrated from Node.js to Go. It monitors and manages:
- AxeOS miners
- MiningCore stratum proxies
- DigiByte Core nodes

The application presents real-time mining data through a web dashboard interface.

## Core Technology Decisions

### Backend Architecture
- **Language**: Pure Go with minimal third-party dependencies
- **Philosophy**: Use Go standard library wherever possible
- **Runtime**: Runs in Docker containers
- **Build System**: GitHub Actions for CI/CD
- **Architecture**: Single Page Application (SPA) with custom router

### Frontend Architecture
- **Language**: Vanilla JavaScript (no frameworks)
- **Markup**: HTML
- **Styling**: CSS in separate files (no inline styles)
- **Build Process**: All JavaScript and CSS are minified during Docker build

### Configuration Management
- **Format**: JSON files (preferred over YAML)
- **Externalization**: Configurations mounted via Docker `-v` flag for persistence
- **Location**: `/config` directory (git-ignored due to sensitive data)

#### Configuration Files
1. **config.json** - Main application settings
2. **access.json** - User credentials (SHA256 hashed passwords)
3. **jsonWebTokenKey.json** - JWT secret and expiration settings
4. **rpcConfig.json** - Cryptocurrency node RPC credentials

### Authentication
- **Method**: JWT (JSON Web Tokens)
- **Dependencies**: `access.json` and `jsonWebTokenKey.json`
- **Optional**: Can be disabled via configuration
- **Cookie Management**: HTTP-only cookies with SameSite=Strict

## Build and Deployment

### Target Architectures
- **AMD64** (x86_64) - Linux, macOS
- **ARM64** (aarch64) - Linux Ubuntu, Raspberry Pi

### Docker Configuration
- **Local Development**: Use `Dockerfile` for build and test
- **Production**: Multi-stage builds
  1. Builder stage: Compiles Go binary, minifies CSS/JS with Node.js tools
  2. Final stage: Alpine-based minimal image

### GitHub Actions
- Automated builds for both AMD64 and ARM64
- See `.github/workflows/` for CI/CD pipelines

## Important File Locations

### Not Pushed to GitHub
```
/config/*
/data/*
```
These files contain sensitive data (credentials, secrets) and metrics data, and are git-ignored.

### Project Structure
```
axeos-dashboard/
├── cmd/server/           # Main application entry point
├── internal/
│   ├── auth/             # JWT authentication
│   ├── config/           # Configuration management (singleton)
│   ├── database/         # SQLite database management (singleton)
│   ├── handlers/         # HTTP request handlers
│   ├── logger/           # Centralized logging system
│   ├── middleware/       # Auth & logging middleware
│   ├── router/           # HTTP routing
│   ├── scheduler/        # Data collection scheduler (time.Ticker tasks)
│   └── services/         # Business logic (crypto nodes, RPC)
├── public/               # Static assets (HTML, CSS, JS)
│   ├── html/
│   ├── css/
│   └── js/
├── config/               # Config files (volume mount point)
├── data/                 # Metrics database (volume mount point)
└── Dockerfile            # Multi-stage Docker build
```

## Development Preferences

### Code Style
- Use Go standard library patterns
- Minimize third-party dependencies
- Keep CSS in separate files
- No inline styles in HTML
- JavaScript and CSS must be minified in production builds

### Configuration Preferences
- JSON over YAML for all config files
- External configuration for Docker containers
- Volume mounts for config persistence: `-v /path/to/config:/app/config`
- Volume mounts for data persistence: `-v /path/to/data:/app/data`

### Docker Workflow
1. Build locally using `Dockerfile`
2. Test in container before pushing
3. Use volume mounts to persist configurations across container replacements

## Key Features

### Hot Configuration Reload
- Update settings without server restart
- Changes made through UI are immediately reflected
- Mutex-based thread-safe config management

### Bootstrap Mode
- Automatically starts in bootstrap mode if config files don't exist
- Switches to normal mode when configuration is complete
- No manual intervention required

### API Design
- RESTful endpoints
- JSON request/response bodies
- JWT authentication on protected routes
- Supports device control, statistics, and configuration management

### Data Collection Architecture
- **Database**: SQLite embedded analytical database for metrics storage (modernc.org/sqlite - pure Go, no CGO)
- **Scheduler**: Singleton pattern using Go's `time.Ticker` for scheduled tasks
- **Non-blocking**: Each collection task runs in its own goroutine
- **Separation of Concerns**: `/internal/database` for persistence, `/internal/scheduler` for task management
- **Configurable**: Collection intervals, retention periods, enable/disable per config
- **Tables**: Three main tables for miners, pools, and nodes
  - `axeos_metrics` - Miner device metrics (hashrate, temp, power, shares, etc.)
  - `pool_metrics` - Mining pool statistics (hashrate, workers, blocks, etc.)
  - `node_metrics` - Cryptocurrency node data (block height, connections, mempool, etc.)

### Logging Architecture
- **Centralized Logger**: Single logging utility in `/internal/logger` package
- **Standard Format**: `[timestamp] [client_ip/system] [module] action`
- **Modules**: main, config, database, scheduler, middleware, service, auth
- **HTTP Request Logging**: Automatic client IP extraction from headers (X-Forwarded-For, X-Real-IP)
- **Methods**: `Info()`, `Error()`, `Warn()`, `Fatal()`, `Debug()`, plus `*WithRequest()` variants
- **Timestamp Format**: YYYY-MM-DD HH:MM:SS (easy to parse and filter)

## Migration from Bitaxe Dashboard

This project maintains 100% compatibility with the Node.js Bitaxe Dashboard:
- Same API endpoints
- Same configuration file formats
- Same authentication flow (SHA256 + JWT)
- Drop-in replacement - just copy config files and run

## Performance Improvements over Node.js Version

| Metric | Node.js | Go | Improvement |
|--------|---------|----|----|
| Memory | 50-80 MB | 15-25 MB | 60-70% less |
| Binary Size | ~150 MB | ~15 MB | 90% smaller |
| Startup Time | 2-3 sec | 50-100 ms | 95% faster |
| Request Latency | 5-10 ms | 1-2 ms | 80% faster |
| Concurrent Requests | ~500/sec | ~5000/sec | 10x improvement |

## Security Practices

- JWT authentication with HTTP-only cookies
- SameSite=Strict for CSRF protection
- SHA256 password hashing
- No debug symbols in production builds
- Cache control headers to prevent stale data
- Thread-safe configuration management

## Important Notes for AI Assistant

1. **Prefer Go Standard Library**: Always check if functionality exists in stdlib before suggesting third-party packages
2. **Respect JSON Format**: Don't suggest converting configs to YAML
3. **Maintain Docker Compatibility**: Any changes must work in containerized environment
4. **Config Externalization**: Remember that configs are mounted, not baked into image
5. **Multi-Architecture**: Consider both AMD64 and ARM64 when making build-related changes
6. **Minification Required**: Frontend assets must be minified for production
7. **CSS Separation**: Never suggest inline styles - always use separate CSS files
8. **No Config in Git**: Never commit or suggest committing `/config` or `/data` files
9. **Data Persistence**: Remember that metrics database is mounted at `/data` and survives container replacements
10. **SQLite Usage**: Use modernc.org/sqlite (pure Go, no CGO) for analytical queries on historical metrics data
11. **Logging Standards**: All new modules must use the centralized logger from `/internal/logger` with standard format
12. **Pure Go Only**: Avoid CGO dependencies to maintain cross-platform compatibility (Alpine/Debian, AMD64/ARM64)

## Reference Documentation

See [README.md](README.md) for complete API documentation, deployment instructions, and troubleshooting guides.
