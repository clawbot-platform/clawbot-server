#!/bin/sh
set -eu

timeout_seconds="${1:-180}"
deadline=$(( $(date +%s) + timeout_seconds ))

while [ "$(date +%s)" -lt "${deadline}" ]; do
  if ./scripts/smoke.sh >/tmp/clawbot-stack-smoke.log 2>&1; then
    cat /tmp/clawbot-stack-smoke.log
    exit 0
  fi

  sleep 5
done

cat /tmp/clawbot-stack-smoke.log
echo "foundation stack did not become ready within ${timeout_seconds}s" >&2
exit 1
