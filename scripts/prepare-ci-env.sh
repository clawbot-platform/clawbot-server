#!/bin/sh
set -eu

sed \
  -e 's/REQUIRED_SECRET_replace_me_postgres/test-postgres-password/g' \
  -e 's/REQUIRED_SECRET_replace_me_minio/test-minio-password/g' \
  -e 's/REQUIRED_SECRET_replace_me_grafana/test-grafana-password/g' \
  -e 's/REQUIRED_SECRET_replace_me_omniroute_api_key/test-omniroute-key/g' \
  -e 's/REQUIRED_SECRET_replace_me_zeroclaw_token/test-zeroclaw-token/g' \
  .env.example > .env
