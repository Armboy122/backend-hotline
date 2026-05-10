#!/usr/bin/env bash
# smoke.sh — lightweight API smoke test for backend-hotline
# Usage: BASE_URL=http://localhost:8080 USERNAME=admin PASSWORD=secret ./scripts/smoke.sh
set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8081}"
USERNAME="${USERNAME:-}"
PASSWORD="${PASSWORD:-}"
TEAM_ID="${TEAM_ID:-1}"
CURRENT_YEAR="$(date +%Y)"
CURRENT_MONTH="$(date +%m)"
SMOKE_FROM="${SMOKE_FROM:-${CURRENT_YEAR}-${CURRENT_MONTH}-01}"
SMOKE_TO="${SMOKE_TO:-${CURRENT_YEAR}-${CURRENT_MONTH}-28}"
TOKEN=""
REFRESH_TOKEN=""
PASSED=0
FAILED=0
SKIPPED=0

print_result() {
  local status="$1" label="$2" detail="${3:-}"
  case "$status" in
    pass) PASSED=$((PASSED + 1)); echo "PASS $label${detail:+ — $detail}" ;;
    fail) FAILED=$((FAILED + 1)); echo "FAIL $label${detail:+ — $detail}" ;;
    skip) SKIPPED=$((SKIPPED + 1)); echo "SKIP $label${detail:+ — $detail}" ;;
  esac
}

contains_code() {
  local expected="$1" actual="$2"
  [[ "|$expected|" == *"|$actual|"* ]]
}

extract_json_field() {
  local field="$1"
  python3 -c 'import json,sys
field=sys.argv[1]
try:
    payload=json.load(sys.stdin)
except Exception:
    print(""); raise SystemExit(0)
data=payload.get("data") if isinstance(payload, dict) else None
if isinstance(data, dict) and data.get(field):
    print(data.get(field)); raise SystemExit(0)
if isinstance(payload, dict):
    print(payload.get(field, ""))
' "$field"
}

request_status() {
  local method="$1" path="$2" body="${3:-}" auth="${4:-no}"
  local args=(-sS -o /dev/null -w '%{http_code}' -X "$method" "${BASE_URL}${path}")
  if [[ "$auth" == "yes" ]]; then
    args+=(-H "Authorization: Bearer ${TOKEN}")
  fi
  if [[ -n "$body" ]]; then
    args+=(-H 'Content-Type: application/json' -d "$body")
  fi
  curl "${args[@]}" 2>/dev/null || true
}

hit() {
  local method="$1" path="$2" expected="$3" body="${4:-}" auth="${5:-no}" label="${6:-$method $path}"
  local status
  status=$(request_status "$method" "$path" "$body" "$auth")
  if contains_code "$expected" "$status"; then
    print_result pass "$label" "$status"
  else
    print_result fail "$label" "got $status, expected $expected"
  fi
}

hit_auth_or_skip() {
  local method="$1" path="$2" expected="$3" body="${4:-}" label="${5:-$method $path}"
  if [[ -z "$TOKEN" ]]; then
    print_result skip "$label" "USERNAME/PASSWORD not set"
    return
  fi
  hit "$method" "$path" "$expected" "$body" yes "$label"
}

# Public readiness.
hit GET /health 200 "" no "GET /health"

# Auth flow.
if [[ -n "$USERNAME" && -n "$PASSWORD" ]]; then
  LOGIN_BODY=$(python3 -c 'import json,os; print(json.dumps({"username": os.environ["USERNAME"], "password": os.environ["PASSWORD"]}))')
  LOGIN_RESP=$(curl -sS -X POST "${BASE_URL}/v1/auth/login" -H 'Content-Type: application/json' -d "$LOGIN_BODY" 2>/dev/null || true)
  TOKEN=$(printf '%s' "$LOGIN_RESP" | extract_json_field accessToken)
  REFRESH_TOKEN=$(printf '%s' "$LOGIN_RESP" | extract_json_field refreshToken)
  if [[ -n "$TOKEN" ]]; then
    print_result pass "POST /v1/auth/login" "token received"
  else
    print_result fail "POST /v1/auth/login" "no accessToken in response"
  fi
else
  print_result skip "POST /v1/auth/login" "USERNAME/PASSWORD not set"
fi

hit_auth_or_skip GET /v1/auth/me 200 "" "GET /v1/auth/me"
if [[ -n "$TOKEN" && -n "$REFRESH_TOKEN" ]]; then
  REFRESH_BODY=$(python3 -c 'import json,sys; print(json.dumps({"refreshToken": sys.argv[1]}))' "$REFRESH_TOKEN")
  hit POST /v1/auth/refresh 200 "$REFRESH_BODY" no "POST /v1/auth/refresh"
else
  print_result skip "POST /v1/auth/refresh" "no refresh token"
fi

# User management smoke: 403 is acceptable for non-admin credentials; 200 proves admin/super_admin readiness.
hit_auth_or_skip GET /v1/users '200|403' "" "GET /v1/users"

# Task daily smoke: unauth must be blocked; auth list/filter should work for scoped users.
hit GET '/v1/tasks?page=1&limit=1' 401 "" no "GET /v1/tasks without token"
hit_auth_or_skip GET '/v1/tasks?page=1&limit=1' 200 "" "GET /v1/tasks?page=1&limit=1"
hit_auth_or_skip GET "/v1/tasks/by-filter?page=1&limit=1&year=${CURRENT_YEAR}&month=${CURRENT_MONTH}" 200 "" "GET /v1/tasks/by-filter"

# Monthly plan smoke: all authenticated roles can read yearly overview/period/files/status; admin endpoints allow 403 for non-admin creds.
hit_auth_or_skip GET "/v1/monthly-plans/${CURRENT_YEAR}/overview" 200 "" "GET /v1/monthly-plans/${CURRENT_YEAR}/overview"
hit_auth_or_skip GET "/v1/monthly-plans/${CURRENT_YEAR}/${CURRENT_MONTH}" 200 "" "GET /v1/monthly-plans/${CURRENT_YEAR}/${CURRENT_MONTH}"
hit_auth_or_skip GET "/v1/monthly-plans/${CURRENT_YEAR}/${CURRENT_MONTH}/files" 200 "" "GET /v1/monthly-plans/${CURRENT_YEAR}/${CURRENT_MONTH}/files"
hit_auth_or_skip GET "/v1/monthly-plans/${CURRENT_YEAR}/${CURRENT_MONTH}/status" 200 "" "GET /v1/monthly-plans/${CURRENT_YEAR}/${CURRENT_MONTH}/status"
hit_auth_or_skip GET /v1/monthly-plans/settings '200|403' "" "GET /v1/monthly-plans/settings"

# Dashboard/report smoke and scoping filters. Dashboard routes allow super_admin/viewer; admin credentials are expected to receive 403.
hit GET /v1/dashboard/summary 401 "" no "GET /v1/dashboard/summary without token"
hit_auth_or_skip GET /v1/dashboard/summary '200|403' "" "GET /v1/dashboard/summary"
hit_auth_or_skip GET "/v1/dashboard/summary?teamId=${TEAM_ID}" '200|403' "" "GET /v1/dashboard/summary?teamId=${TEAM_ID}"
hit_auth_or_skip GET /v1/dashboard/top-jobs '200|403' "" "GET /v1/dashboard/top-jobs"
hit_auth_or_skip GET /v1/dashboard/top-feeders '200|403' "" "GET /v1/dashboard/top-feeders"
hit_auth_or_skip GET /v1/dashboard/stats '200|403' "" "GET /v1/dashboard/stats"

# Planning workstream smoke: unauthenticated requests must be blocked; authenticated requests are read/list checks.
hit GET /v1/contact-directory 401 "" no "GET /v1/contact-directory without token"
hit_auth_or_skip GET /v1/contact-directory 200 "" "GET /v1/contact-directory"
hit GET "/v1/team-plans?from=${SMOKE_FROM}&to=${SMOKE_TO}" 401 "" no "GET /v1/team-plans without token"
hit_auth_or_skip GET "/v1/team-plans?from=${SMOKE_FROM}&to=${SMOKE_TO}" 200 "" "GET /v1/team-plans?from=${SMOKE_FROM}&to=${SMOKE_TO}"
hit GET "/v1/planning/calendar/${CURRENT_YEAR}/${CURRENT_MONTH}" 401 "" no "GET /v1/planning/calendar without token"
hit_auth_or_skip GET "/v1/planning/calendar/${CURRENT_YEAR}/${CURRENT_MONTH}" 200 "" "GET /v1/planning/calendar/${CURRENT_YEAR}/${CURRENT_MONTH}"
hit GET "/v1/large-work-items?from=${SMOKE_FROM}&to=${SMOKE_TO}" 401 "" no "GET /v1/large-work-items without token"
hit_auth_or_skip GET "/v1/large-work-items?from=${SMOKE_FROM}&to=${SMOKE_TO}" 200 "" "GET /v1/large-work-items?from=${SMOKE_FROM}&to=${SMOKE_TO}"
hit GET "/v1/daily-report-drafts/sources?workDate=${SMOKE_FROM}" 401 "" no "GET /v1/daily-report-drafts/sources without token"
hit_auth_or_skip GET "/v1/daily-report-drafts/sources?workDate=${SMOKE_FROM}" 200 "" "GET /v1/daily-report-drafts/sources?workDate=${SMOKE_FROM}"

TOTAL=$((PASSED + FAILED + SKIPPED))
echo ""
echo "Results: ${PASSED} PASSED / ${FAILED} FAILED / ${SKIPPED} SKIPPED / ${TOTAL} total"

if [[ "$FAILED" -gt 0 ]]; then
  exit 1
fi
