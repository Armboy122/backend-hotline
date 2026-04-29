#!/usr/bin/env bash
# smoke.sh — lightweight API smoke test for backend-hotline
# Usage: BASE_URL=http://localhost:8081 USERNAME=640100 PASSWORD=secret ./scripts/smoke.sh
set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8081}"
USERNAME="${USERNAME:-}"
PASSWORD="${PASSWORD:-}"
TOKEN=""
CURRENT_YEAR="$(date +%Y)"
CURRENT_MONTH="$(date +%m)"

fail() { echo "FAIL: $*" >&2; exit 1; }

# ── Health ────────────────────────────────────────────────────────────────────
echo "→ GET ${BASE_URL}/health"
STATUS=$(curl -s -o /dev/null -w '%{http_code}' "${BASE_URL}/health")
[[ "$STATUS" == "200" ]] || fail "/health returned $STATUS (expected 200)"
echo "  ✅ /health → 200"

# ── Login (optional) ─────────────────────────────────────────────────────────
if [[ -n "$USERNAME" && -n "$PASSWORD" ]]; then
  echo "→ POST ${BASE_URL}/v1/auth/login"
  LOGIN_RESP=$(curl -sf -X POST "${BASE_URL}/v1/auth/login" \
    -H 'Content-Type: application/json' \
    -d "{\"username\":\"${USERNAME}\",\"password\":\"${PASSWORD}\"}" 2>/dev/null) || fail "login request failed"
  TOKEN=$(echo "$LOGIN_RESP" | grep -o '"accessToken":"[^"]*"' | head -1 | cut -d'"' -f4)
  [[ -n "$TOKEN" ]] || fail "could not extract accessToken from login response"
  echo "  ✅ login → got token"
else
  echo "  ⏭️  skipping login (USERNAME/PASSWORD not set)"
fi

# ── GET /v1/auth/me ──────────────────────────────────────────────────────────
if [[ -n "$TOKEN" ]]; then
  echo "→ GET ${BASE_URL}/v1/auth/me"
  ME_STATUS=$(curl -sf -o /dev/null -w '%{http_code}' \
    -H "Authorization: Bearer ${TOKEN}" \
    "${BASE_URL}/v1/auth/me" 2>/dev/null) || ME_STATUS="000"
  [[ "$ME_STATUS" =~ ^2..$ ]] || fail "/v1/auth/me returned $ME_STATUS"
  echo "  ✅ /v1/auth/me → $ME_STATUS"
fi

# ── GET /v1/tasks ─────────────────────────────────────────────────────────────
echo "→ GET ${BASE_URL}/v1/tasks?page=1&limit=1"
if [[ -n "$TOKEN" ]]; then
  TASKS_STATUS=$(curl -sf -o /dev/null -w '%{http_code}' \
    -H "Authorization: Bearer ${TOKEN}" \
    "${BASE_URL}/v1/tasks?page=1&limit=1" 2>/dev/null) || TASKS_STATUS="000"
else
  TASKS_STATUS=$(curl -sf -o /dev/null -w '%{http_code}' \
    "${BASE_URL}/v1/tasks?page=1&limit=1" 2>/dev/null) || TASKS_STATUS="000"
fi
if [[ "$TASKS_STATUS" == "401" && -z "$TOKEN" ]]; then
  echo "  ⏭️  /v1/tasks → 401 (no token, expected)"
elif [[ "$TASKS_STATUS" =~ ^2..$ ]]; then
  echo "  ✅ /v1/tasks → $TASKS_STATUS"
else
  fail "/v1/tasks returned $TASKS_STATUS"
fi

# ── GET /v1/monthly-plans/:year/:month (authenticated) ───────────────────────
if [[ -n "$TOKEN" ]]; then
  echo "→ GET ${BASE_URL}/v1/monthly-plans/${CURRENT_YEAR}/${CURRENT_MONTH}"
  MONTHLY_STATUS=$(curl -sf -o /dev/null -w '%{http_code}' \
    -H "Authorization: Bearer ${TOKEN}" \
    "${BASE_URL}/v1/monthly-plans/${CURRENT_YEAR}/${CURRENT_MONTH}" 2>/dev/null) || MONTHLY_STATUS="000"
  [[ "$MONTHLY_STATUS" =~ ^2..$ ]] || fail "/v1/monthly-plans/${CURRENT_YEAR}/${CURRENT_MONTH} returned $MONTHLY_STATUS"
  echo "  ✅ /v1/monthly-plans/${CURRENT_YEAR}/${CURRENT_MONTH} → $MONTHLY_STATUS"
fi

# ── GET /v1/dashboard/summary (optional) ─────────────────────────────────────
if [[ -n "$TOKEN" ]]; then
  echo "→ GET ${BASE_URL}/v1/dashboard/summary"
  DASH_STATUS=$(curl -sf -o /dev/null -w '%{http_code}' \
    -H "Authorization: Bearer ${TOKEN}" \
    "${BASE_URL}/v1/dashboard/summary" 2>/dev/null) || DASH_STATUS="000"
  if [[ "$DASH_STATUS" =~ ^2..$ ]]; then
    echo "  ✅ /v1/dashboard/summary → $DASH_STATUS"
  else
    echo "  ⚠️  /v1/dashboard/summary → $DASH_STATUS (non-blocking)"
  fi
fi

echo ""
echo "Smoke test passed ✅"
