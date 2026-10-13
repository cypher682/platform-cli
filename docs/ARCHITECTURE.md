# Architecture

Single Go binary. `cmd/` holds thin cobra commands; all logic lives in `internal/` packages that are testable without a terminal.

| Package | Responsibility |
|---|---|
| `internal/config` | Load `platform.yaml` (viper). Sentinel `ErrNoServices`. |
| `internal/health` | Concurrent HTTP checks: one goroutine per service, `WaitGroup`, per-check `context` timeout. |
| `internal/docker` | `docker compose` wrapper (shell-out) and Docker SDK client for log streaming. |
| `internal/logs` | `Source` interface + `FanIn`: N goroutines send into one channel; regex matcher. |
| `internal/migrate` | `Migrator` interface with Prisma, Alembic and golang-migrate backends. |
| `internal/scaffold` | `go:embed` templates rendered with `[[ ]]` delimiters; go-git initial commit. |

## Decisions

- **Compose via subprocess, logs via SDK.** `docker compose` is a CLI plugin with no stable SDK; streaming logs is where the SDK pays off (demultiplexing stdout/stderr, context cancellation).
- **Interfaces at the seams that vary.** `logs.Source` lets fan-in be tested with a fake; `migrate.Migrator` lets `db` commands ignore which tool is configured.
- **Custom template delimiters.** Helm chart stubs use `{{ }}`; the scaffolder uses `[[ ]]` so neither syntax needs escaping.
- **Honest unsupported operations.** Prisma has no down-migrations; `db rollback` returns `ErrUnsupported` with the manual recovery path rather than faking it.
- **Exit codes.** `health check` exits non-zero when any service is unhealthy, so it works in scripts and CI.

## Known limits

- `env up` waits on HTTP health endpoints only; services without one are reported as down.
- `logs tail` resolves containers by the compose service label, then by container name.
- `deploy local` assumes minikube, helm and kubectl are on PATH and each service directory contains a `Dockerfile` and a `helm/` chart (the scaffolder generates both).
