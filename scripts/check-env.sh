#!/usr/bin/env sh
set -eu

ENV_FILE="${1:-.env}"
VALIDATE_OPTIONAL_STACK="${VALIDATE_OPTIONAL_STACK:-0}"

if [ ! -f "$ENV_FILE" ]; then
  echo "Missing $ENV_FILE. Copy .env.example to $ENV_FILE first."
  exit 1
fi

core_required_vars="
COMPOSE_PROJECT_NAME
POSTGRES_IMAGE
POSTGRES_HOST
POSTGRES_PORT
POSTGRES_DB
POSTGRES_USER
POSTGRES_PASSWORD
REDIS_IMAGE
REDIS_HOST
REDIS_PORT
NATS_IMAGE
NATS_HOST
NATS_PORT
NATS_MONITOR_PORT
STACK_SMOKE_TIMEOUT
APP_ENV
SERVER_ADDRESS
LOG_LEVEL
AUTO_MIGRATE
SHUTDOWN_TIMEOUT
DATABASE_URL
"

optional_required_vars="
MINIO_IMAGE
MINIO_HOST
MINIO_PORT
MINIO_CONSOLE_PORT
MINIO_ROOT_USER
MINIO_ROOT_PASSWORD
PROMETHEUS_IMAGE
PROMETHEUS_HOST
PROMETHEUS_PORT
GRAFANA_IMAGE
GRAFANA_HOST
GRAFANA_PORT
GRAFANA_ADMIN_USER
GRAFANA_ADMIN_PASSWORD
OMNIROUTE_IMAGE
OMNIROUTE_HOST
OMNIROUTE_PORT
OMNIROUTE_BASE_URL
ZEROCLAW_VERSION
ZEROCLAW_HOST
ZEROCLAW_PORT
ZEROCLAW_PROVIDER
ZEROCLAW_API_URL
ZEROCLAW_API_KEY
ZEROCLAW_DEFAULT_MODEL
ZEROCLAW_GATEWAY_TOKEN
ZEROCLAW_RUST_LOG
"

# shellcheck disable=SC1090
set -a
case "$ENV_FILE" in
  */*) . "$ENV_FILE" ;;
  *) . "./$ENV_FILE" ;;
esac
set +a

validate_vars() {
  required_vars="$1"

  for var in $required_vars; do
    eval "value=\${$var:-}"
    if [ -z "$value" ]; then
      echo "ERROR: Required variable '$var' is missing or empty."
      exit 1
    fi

    case "$value" in
      *REQUIRED_SECRET*|*replace_me*)
        echo "ERROR: Variable '$var' still contains a placeholder value."
        exit 1
        ;;
    esac
  done
}

validate_vars "$core_required_vars"

if [ "$VALIDATE_OPTIONAL_STACK" = "1" ]; then
  validate_vars "$optional_required_vars"
  echo "Environment file validation passed for core + optional stack: $ENV_FILE"
else
  echo "Environment file validation passed for core stack: $ENV_FILE"
fi
