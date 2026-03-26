# Architecture

## Repository role

`clawbot-server` is the reusable platform-foundation repository in the `clawbot-platform` organization.

It provides:

- local infrastructure bootstrap
- a small generic control-plane service
- shared operational defaults for storage, messaging, routing, observability, and audit data

It does not own downstream business logic.

## Topology

The local topology is intentionally simple:

- ZeroClaw provides the runtime substrate
- OmniRoute provides model ingress and routing
- Postgres, Redis, NATS, and MinIO provide shared stateful services
- Prometheus and Grafana provide observability
- `clawbot-server` layers a small HTTP control plane on top of that foundation

## Boundary decisions

- domain logic stays out of this repo
- OmniRoute remains responsible for model routing
- ZeroClaw remains responsible for runtime behavior
- this repo owns shared bootstrap and generic control-plane scaffolding
- downstream repositories are consumers, not requirements

## Reuse model

Any downstream project can:

- consume the shared foundation stack directly
- call the control-plane API for generic run, bot, policy, and audit management
- reuse the same local development and CI patterns without copying infrastructure bootstrap into its own repository

Downstream verticals are examples of consumers only. The server remains usable without any of them.
