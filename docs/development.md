# Development

## Local workflow

1. Copy `.env.example` to `.env`.
2. Start the shared stack with `make up`.
3. Apply migrations with `make migrate-up`.
4. Run the control-plane service with `make run-server`.
5. Optionally verify readiness with `make smoke`.
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
- Keep downstream business logic outside this repository.
- Treat downstream verticals as consumer examples, not design constraints.
