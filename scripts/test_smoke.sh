#!/usr/bin/env bash
# Static regression checks for scripts/smoke.sh.
set -euo pipefail

SCRIPT="${SCRIPT:-scripts/smoke.sh}"

bash -n "$SCRIPT"

if grep -F 'Authorization: Bearer ***' "$SCRIPT" >/dev/null; then
  echo "smoke script must send the real token, not a redacted placeholder" >&2
  exit 1
fi

grep -F 'Authorization: Bearer ${TOKEN}' "$SCRIPT" >/dev/null || {
  echo "smoke script must send Authorization with the real TOKEN variable" >&2
  exit 1
}

required_patterns=(
  '/health'
  '/v1/auth/login'
  '/v1/auth/me'
  '/v1/auth/refresh'
  '/v1/users'
  '/v1/tasks?page=1&limit=1'
  '/v1/tasks/by-filter'
  '/v1/monthly-plans/${CURRENT_YEAR}/overview'
  '/v1/monthly-plans/${CURRENT_YEAR}/${CURRENT_MONTH}'
  '/v1/monthly-plans/${CURRENT_YEAR}/${CURRENT_MONTH}/status'
  'GET /v1/dashboard/summary without token'
  'hit_auth_or_skip GET /v1/dashboard/summary'
  '/v1/dashboard/top-jobs'
  '/v1/dashboard/top-feeders'
  '/v1/dashboard/stats'
  '/v1/contact-directory'
  '/v1/team-plans?from=${SMOKE_FROM}&to=${SMOKE_TO}'
  '/v1/planning/calendar/${CURRENT_YEAR}/${CURRENT_MONTH}'
  '/v1/large-work-items?from=${SMOKE_FROM}&to=${SMOKE_TO}'
  '/v1/daily-report-drafts/sources?workDate=${SMOKE_FROM}'
)

for pattern in "${required_patterns[@]}"; do
  grep -F "$pattern" "$SCRIPT" >/dev/null || {
    echo "missing smoke coverage pattern: $pattern" >&2
    exit 1
  }
done

grep -F 'payload.get("data")' "$SCRIPT" >/dev/null && grep -F 'extract_json_field accessToken' "$SCRIPT" >/dev/null || {
  echo "token extraction must support wrapped data.accessToken responses" >&2
  exit 1
}

grep -F 'os.environ["REFRESH_TOKEN"]' "$SCRIPT" >/dev/null && {
  echo "refresh smoke must not read internally assigned REFRESH_TOKEN via os.environ" >&2
  exit 1
}

grep -F "|| printf '000'" "$SCRIPT" >/dev/null && {
  echo "request_status must not append fallback 000 to curl's %{http_code}" >&2
  exit 1
}

grep -F 'Results:' "$SCRIPT" >/dev/null || {
  echo "smoke script must print a pass/fail summary" >&2
  exit 1
}

echo "smoke static checks passed"
