#!/bin/sh
set -eu

INPUT_FILE="${1:-.env.example}"
OUTPUT_FILE="${2:-.env}"

if [ ! -f "$INPUT_FILE" ]; then
  echo "Missing $INPUT_FILE"
  exit 1
fi

sed \
  -e 's/REQUIRED_SECRET_replace_me_postgres/test-postgres-password/g' \
  -e 's/REQUIRED_SECRET_replace_me_minio/test-minio-password/g' \
  -e 's/REQUIRED_SECRET_replace_me_grafana/test-grafana-password/g' \
  -e 's/REQUIRED_SECRET_replace_me_omniroute_api_key/test-omniroute-key/g' \
  -e 's/REQUIRED_SECRET_replace_me_zeroclaw_token/test-zeroclaw-token/g' \
  "$INPUT_FILE" > "$OUTPUT_FILE"

echo "Prepared CI environment file: $OUTPUT_FILE"
