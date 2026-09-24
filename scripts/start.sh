#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
if docker compose version >/dev/null 2>&1; then compose=(docker compose); else compose=(docker-compose); fi
"${compose[@]}" up -d --build --wait --wait-timeout 180
# Kong caches upstream DNS targets. Refresh it after service containers are
# recreated so the gateway always points at the current container IPs.
"${compose[@]}" restart api-gateway
sleep 5
python3 scripts/configure-gateway.py
python3 scripts/simulate.py
python3 scripts/verify-observability.py
