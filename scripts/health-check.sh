#!/usr/bin/env bash
# health-check.sh - Polls OpsTrack's /health endpoint until it responds or times out.
#
# Usage:
#   scripts/health-check.sh [base_url] [timeout_seconds]
#
# Examples:
#   scripts/health-check.sh                              # http://localhost:8080
#   scripts/health-check.sh http://localhost:8080 60
#   scripts/health-check.sh https://opstrack.example.com  # any reachable deployment

set -uo pipefail

BASE_URL="${1:-http://localhost:8080}"
TIMEOUT="${2:-30}"
INTERVAL=2
elapsed=0

echo "Checking ${BASE_URL}/health (timeout: ${TIMEOUT}s)..."

while [ "$elapsed" -lt "$TIMEOUT" ]; do
  if response="$(curl -fsS "${BASE_URL}/health" 2>/dev/null)"; then
    echo "OK: $response"
    exit 0
  fi
  sleep "$INTERVAL"
  elapsed=$((elapsed + INTERVAL))
  echo "  ...not ready yet (${elapsed}s elapsed)"
done

echo "FAILED: ${BASE_URL}/health did not become healthy within ${TIMEOUT}s" >&2
exit 1
