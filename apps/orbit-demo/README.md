# orbit-demo

orbit-demo is Orbit's sample workload: a small Go HTTP service used to
exercise the platform. It is also the reference implementation of the Golden
Path workload contract
([ADR-0002](../../docs/architecture/decisions/0002-developer-contract.md),
[ADR-0011](../../docs/architecture/decisions/0011-go-for-platform-services.md)):
it listens on a configured port, exposes health, readiness and metrics
endpoints, writes JSON logs to stdout, shuts down gracefully and runs as a
non-root user in a minimal image.

## Endpoints

All responses are JSON (`Content-Type: application/json`), except `/metrics`.

| Request | Response |
|---|---|
| `GET /` | 200 `{"service":"orbit-demo","message":…,"version":…,"hostname":…}` |
| `GET /healthz` | 200 `{"status":"ok"}` while the process runs (liveness; checks nothing else) |
| `GET /readyz` | 200 `{"status":"ready"}`; 503 `{"status":"shutting down"}` once shutdown starts |
| `GET /metrics` | Prometheus text exposition |
| Unknown path | 404 `{"error":"not found"}` |
| Wrong method on a known path | 405 `{"error":"method not allowed"}`, `Allow: GET, HEAD` |

orbit-demo has no dependencies, so it is ready as soon as its listener
accepts requests.

### Metrics

- Go runtime (`go_*`) and process (`process_*`) metrics.
- `http_requests_total{method,route,code}` and
  `http_request_duration_seconds{method,route}` (histogram, default buckets).
- `orbit_demo_build_info{version}`, always 1.

`route` is the registered route (`/`, `/healthz`, `/readyz`, `/metrics`),
never the raw request path; anything else is `route="unmatched"`. Unknown
methods are counted as `method="OTHER"`. Both keep label cardinality
bounded.

## Configuration

Environment variables, validated at startup. An invalid value logs one JSON
error line and the process exits with code 1.

| Variable | Default | Meaning |
|---|---|---|
| `PORT` | `8080` | Listen port (1–65535) |
| `LOG_LEVEL` | `info` | `debug`, `info`, `warn` or `error` |
| `MESSAGE` | `Hello from Orbit` | Returned by `/`, so a configuration change made through Git is visible |
| `SHUTDOWN_DRAIN` | `5s` | Time between readiness turning 503 and the listener closing (Go duration, ≥ 0) |
| `SHUTDOWN_TIMEOUT` | `10s` | Maximum time for in-flight requests to finish (Go duration, > 0) |

## Logs

One JSON object per line on stdout, with `time`, `level` and `msg`:

- a startup line with the version and the effective configuration;
- one `request` line per request at `info`, with `method`, `path`, `route`,
  `status` and `duration_ms`. Requests to `/healthz`, `/readyz` and
  `/metrics` are logged at `debug` only, so probes and scrapes don't flood
  the logs;
- the shutdown steps.

```json
{"time":"…","level":"INFO","msg":"request","method":"GET","path":"/","route":"/","status":200,"duration_ms":0.393}
```

## Graceful shutdown

On SIGTERM or SIGINT:

1. `/readyz` returns 503 and a `shutting down` line is logged. Keep-alive
   connections are closed after their current request.
2. The server keeps serving for `SHUTDOWN_DRAIN`, so Kubernetes can remove
   the pod from its endpoints before the listener closes.
3. The listener closes and in-flight requests get up to `SHUTDOWN_TIMEOUT`
   to finish.
4. Exit code 0 on a clean shutdown, 1 if the timeout is hit.

The defaults (5 s + 10 s) stay under Kubernetes' default 30 s termination
grace period.

## Image

- Multi-stage build: `golang:1.27.1` builds a static binary
  (`CGO_ENABLED=0`, `-trimpath`, `-s -w`); the runtime image is
  `gcr.io/distroless/static-debian13:nonroot`. Both are pinned by digest.
- The final image holds only the binary on top of distroless: no shell, no
  package manager.
- Runs as user `65532:65532` (numeric, so Kubernetes can enforce
  `runAsNonRoot`), exposes 8080, and the binary is PID 1 (exec-form
  `ENTRYPOINT`).
- The version is injected at build time from the `VERSION` build argument
  (default `dev`) and is also the `org.opencontainers.image.revision` label.
  Images are tagged with the commit SHA, never `latest`; the `local` tag below
  is for local use only and is never pushed.

## Build, run and test locally

Requires Go 1.27 and Docker. Commands run from the repository root.

```sh
# Tests
cd apps/orbit-demo
gofmt -l .
go vet ./...
go test -race ./...
cd ../..

# Image
docker build --build-arg VERSION=$(git rev-parse --short HEAD) \
  -t orbit-demo:local apps/orbit-demo

# Run with the same restrictions the platform applies
docker run --rm -p 8080:8080 --read-only --cap-drop=ALL \
  --security-opt no-new-privileges orbit-demo:local

curl -s localhost:8080/
curl -s localhost:8080/healthz
curl -s localhost:8080/readyz
curl -s localhost:8080/metrics

# Overrides
docker run --rm -p 9090:9090 -e PORT=9090 -e MESSAGE="Hello from dev" \
  orbit-demo:local
```

Without Docker: `go run .` from `apps/orbit-demo/` (version `dev`).
