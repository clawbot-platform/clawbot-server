# Security

## Phase 0 security posture

Phase 0 is local-development oriented, but the defaults still follow least-privilege and explicit-configuration principles:

- secrets are injected by environment variables rather than committed values
- exposed ports are documented explicitly
- published ports bind to `127.0.0.1` by default for local-only exposure
- named volumes scope state to the stack
- public model access is expected to flow through OmniRoute
- ZeroClaw is integrated without adding custom runtime logic in this repository
- the control-plane service stores only platform metadata and structured audit events

## Threat considerations

- Local default credentials in `.env.example` are placeholders and must be replaced before using shared or remote environments.
- Docker Compose publishes ports to `127.0.0.1` by default for local development convenience. If a later phase needs remote reachability, the `*_HOST` variables must be reviewed deliberately.
- OmniRoute and ZeroClaw are upstream components. This repo integrates them but does not reimplement their security boundaries.
- Prometheus and Grafana are intended for trusted local-lab usage in Phase 0.
- The Phase 1 HTTP API is local-dev oriented and unauthenticated by default. The `X-Actor` header is a scaffolding hook for audit attribution, not a production auth mechanism.

## CI scanning

The repository includes workflows for:

- `gofmt`
- `go vet`
- unit tests
- `golangci-lint`
- `gosec`
- `govulncheck`
- `gitleaks`
- `trivy`
