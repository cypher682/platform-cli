# platform-cli

[![CI](https://github.com/cypher682/platform-cli/actions/workflows/ci.yml/badge.svg)](https://github.com/cypher682/platform-cli/actions/workflows/ci.yml)

A Go CLI for local developer-platform tasks: environment lifecycle, health checks, multiplexed log tailing, database migrations, service scaffolding and local Kubernetes deploys. Single static binary.

## Commands

```
platform-cli
├── env        up | down | status | reset
├── health     check | watch
├── logs       tail | search
├── db         migrate | rollback | seed | status
├── scaffold   service | route
├── deploy     local | status
└── version
```

| Command | Behaviour |
|---|---|
| `env up` | Checks the Docker daemon, runs `docker compose up -d`, waits for every health endpoint (`--timeout`), prints a service table. |
| `env reset` | `down -v`, then `up`. |
| `health check` | One concurrent pass; exits non-zero if any service is not healthy. |
| `health watch` | Repeats every `--interval`; terminal bell and a log line on status change; Ctrl+C exits cleanly. |
| `logs tail [svc...]` | Multiplexes container logs via the Docker SDK, colour-coded per service. `--since`, `--filter` (regex), `--tail`, `-f`. |
| `logs search <regex> [svc...]` | Searches recent logs (`--since`, default 1h) and prints a match count. |
| `db migrate\|rollback\|seed\|status` | Runs the configured tool (`prisma`, `alembic`, `golang-migrate`) inside `database.service`. Prisma has no rollback; the command says so. |
| `scaffold service` | Prompts (or flags) for language/name/port; generates app, Dockerfile, compose snippet, Helm chart stub, then `git init` + initial commit. Languages: `fastapi`, `express`. |
| `scaffold route <name>` | Adds a router file to an existing FastAPI or Express service and prints how to register it. |
| `deploy local --service X` | `docker build` → `minikube image load` → `helm upgrade --install`. |
| `deploy status` | `kubectl get deployments` across one or more `--context`s. |

## Configuration

Copy `platform.yaml.example` to `platform.yaml` (or pass `--config`).

```yaml
project: nexuspay
services:
  - name: api-gateway
    port: 3000
    health: http://localhost:3000/health
    compose: services/api-gateway
database:
  type: postgres
  migration_tool: prisma
  container: postgres
  service: auth-service
```

## Install

```bash
# from a release (Linux/macOS)
curl -fsSL https://raw.githubusercontent.com/cypher682/platform-cli/main/install.sh | sh

# from source
go install github.com/cypher682/platform-cli@latest
```

Windows: download the `.zip` from the Releases page.

## Develop

```bash
go mod tidy
go test ./...                         # unit tests
go test -tags=integration ./tests/... # needs a Docker daemon
go build -o platform-cli .
```

Releases are built by goreleaser on `v*` tags (Linux/macOS/Windows, amd64/arm64).

## Design

See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md). Evidence checklist: [docs/evidence/README.md](docs/evidence/README.md).
