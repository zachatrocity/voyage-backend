#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ENV_FILE="$ROOT_DIR/tests/hurl/.env"

if ! command -v hurl >/dev/null 2>&1; then
  echo "Error: hurl is not installed. See https://hurl.dev for install instructions."
  exit 1
fi

if [[ ! -f "$ENV_FILE" ]]; then
  echo "Missing $ENV_FILE"
  echo "Copy tests/hurl/.env.example to tests/hurl/.env and fill values first."
  exit 1
fi

set -a
source "$ENV_FILE"
set +a

: "${BASE_URL:?BASE_URL is required in tests/hurl/.env}"
: "${API_KEY:?API_KEY is required in tests/hurl/.env}"
SEARCH_QUERY="${SEARCH_QUERY:-*}"
RUN_ID="${RUN_ID:-$(date +%s)}"

COMMON_ARGS=(
  --variable "base_url=$BASE_URL"
  --variable "api_key=$API_KEY"
  --variable "search_query=$SEARCH_QUERY"
  --variable "run_id=$RUN_ID"
)

echo "Running Hurl integration tests against: $BASE_URL"
echo "Run ID: $RUN_ID"

hurl "${COMMON_ARGS[@]}" "$ROOT_DIR/tests/hurl/01-health.hurl"
hurl "${COMMON_ARGS[@]}" "$ROOT_DIR/tests/hurl/02-email-tag-flow.hurl"
hurl "${COMMON_ARGS[@]}" "$ROOT_DIR/tests/hurl/03-trip-flow.hurl"
hurl "${COMMON_ARGS[@]}" "$ROOT_DIR/tests/hurl/04-negative-cases.hurl"

echo "All Hurl integration tests passed."
