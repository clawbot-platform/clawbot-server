#!/bin/sh
set -eu

template=/etc/zeroclaw/config.toml.tmpl
config_dir=/home/zeroclaw/.zeroclaw
config="${config_dir}/config.toml"

if [ ! -f "$template" ]; then
  echo "ERROR: missing template: $template" >&2
  exit 1
fi

if [ ! -x /usr/local/bin/zeroclaw ]; then
  echo "ERROR: /usr/local/bin/zeroclaw is missing or not executable" >&2
  exit 127
fi

mkdir -p "$config_dir" /workspace
chown -R zeroclaw:zeroclaw "$config_dir" /workspace

envsubst < "$template" > "$config"
chown zeroclaw:zeroclaw "$config"
chmod 600 "$config"

exec runuser -u zeroclaw -- /usr/local/bin/zeroclaw gateway --host 0.0.0.0 --port 3000
