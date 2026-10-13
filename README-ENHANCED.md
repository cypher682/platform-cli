# platform-cli

A Go CLI for local developer-platform tasks: environment lifecycle, health checks, multiplexed log tailing, database migrations, service scaffolding and local Kubernetes deploys. Single static binary.

## Quick Start

### Windows
1. Download `platform-cli.exe` from releases (or build from source)
2. Add to PATH or run directly: `platform-cli.exe --help`

### Linux/macOS
```bash
curl -fsSL https://raw.githubusercontent.com/cypher682/platform-cli/main/install.sh | sh
platform-cli --help
```

## Commands

```
platform-cli
├── env        up | down | status | reset
├── health     check | watch
├── logs       tail | search
├── db         migrate | rollback | seed | status
├── scaffold   service | route
├── deploy     local | status
├── completion bash|zsh|fish|powershell
└── version
```

## Global Flags
- `-c, --config` - config file path (default: platform.yaml, auto-discovers)
- `-f, --format text|json` - output format
- `-q, --quiet` - suppress non-essential output
- `--no-color` - disable colors
- `-v, --verbose` - verbose logging

## Usage Examples

```bash
# Start environment with health checks
platform-cli env up --timeout 120s

# Watch health continuously
platform-cli health watch --interval 5s

# Tail logs from multiple services
platform-cli logs tail api auth --follow --since 10m

# Search logs
platform-cli logs search "error" api --since 1h

# Database operations
platform-cli db migrate
platform-cli db status

# Scaffold new service
platform-cli scaffold service --name payment --language fastapi --port 8000

# Deploy locally
platform-cli deploy local --service api --wait
```

## Configuration

Copy `platform.yaml.example` to `platform.yaml`:
```yaml
project: nexuspay
services:
  - name: api-gateway
    port: 3000
    health: http://localhost:3000/health
  - name: dummy
    health_check:
      type: none  # skip health checks
database:
  type: postgres
  migration_tool: prisma
  container: postgres
  service: auth-service
```

Supports health probe types: `http` (default), `tcp`, `none`. Config auto-discovers in CWD and parent dirs; `PLATFORM_CONFIG` env var also works.
