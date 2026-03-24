# Development

## Local workflow

1. Copy `.env.example` to `.env`.
2. Start the shared platform stack with `make up`.
3. Apply database migrations with `make migrate-up`.
4. Run the control-plane service with `make run-server`.
5. Optionally wait for readiness with `make smoke`.
6. Inspect services with `make ps` or `make logs`.

## Make targets

- `make up`
- `make down`
- `make restart`
- `make ps`
- `make logs`
- `make smoke`
- `make run-server`
- `make migrate-up`
- `make migrate-down`
- `make clean`
- `make lint`
- `make test`
- `make security`
- `make compose-validate`

## Contributor guidance

- Keep the repo Go-first.
- Put compose files under `deploy/compose`.
- Put versioned configs under `configs`.
- Keep docs explicit and recruiter-friendly.
- Do not add domain logic that belongs in `clawbot-trust-lab` or `clawmem`.
