#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"

check_endpoint() {
  local path="$1"
  local expected_prefix="${2:-2}"
  local url="${BASE_URL}${path}"
  local body_file
  body_file="$(mktemp)"

  local status
  status="$(curl -fsS -o "${body_file}" -w '%{http_code}' "${url}")"
  if [[ "${status}" != ${expected_prefix}* ]]; then
    echo "Smoke check failed: ${url} returned ${status}" >&2
    echo "--- response body ---" >&2
    cat "${body_file}" >&2
    rm -f "${body_file}"
    exit 1
  fi

  rm -f "${body_file}"
  echo "OK ${status} ${path}"
}

check_endpoint "/health"
check_endpoint "/v1/teams?limit=1&page=1"
check_endpoint "/v1/job-types?limit=1&page=1"
check_endpoint "/v1/tasks?limit=1&page=1"

echo "Smoke checks completed successfully."
