# Security

## Local security posture

The local defaults are development-oriented, but they still follow explicit-configuration and least-privilege principles:

- secrets are injected through environment variables rather than committed values
- exposed ports are documented explicitly
- published ports bind to `127.0.0.1` by default for local-only exposure
- named volumes scope state to the local stack
- public model access is expected to flow through OmniRoute
- ZeroClaw is integrated without reimplementing runtime boundaries in this repository
- the control plane stores platform metadata and structured audit events, not downstream business data

## Threat considerations

- Placeholder values in `.env.example` must be replaced before using shared or remote environments.
- If a deployment needs remote reachability, the `*_HOST` variables must be reviewed deliberately.
- OmniRoute and ZeroClaw are upstream components. This repo integrates them but does not redefine their security boundaries.
- Prometheus and Grafana are intended for trusted local development by default.
- The HTTP API is unauthenticated by default. The `X-Actor` header is an audit-attribution hook, not a production auth system.

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
