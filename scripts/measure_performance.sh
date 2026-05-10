#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"
REPEATS="${REPEATS:-6}"
TOKEN="${TOKEN:-}"
USERNAME="${USERNAME:-}"
PASSWORD="${PASSWORD:-}"

login_if_needed() {
  if [[ -n "$TOKEN" || -z "$USERNAME" || -z "$PASSWORD" ]]; then
    return 0
  fi
  local body resp
  body=$(python3 - <<'PY'
import json, os
print(json.dumps({"username": os.environ["USERNAME"], "password": os.environ["PASSWORD"]}))
PY
)
  resp=$(curl -sS -X POST "$BASE_URL/v1/auth/login" -H 'Content-Type: application/json' -d "$body")
  TOKEN=$(RESP="$resp" python3 - <<'PY'
import json, os
try:
    data=json.loads(os.environ.get('RESP','{}'))
    print(data.get('data',{}).get('access_token',''))
except Exception:
    print('')
PY
)
}

measure() {
  local label="$1" path="$2" auth="${3:-public}"
  local tmp status elapsed statuses i
  tmp=$(mktemp)
  statuses=""
  for ((i=1; i<=REPEATS; i++)); do
    if [[ "$auth" == "auth" && -n "$TOKEN" ]]; then
      status=$(curl -sS -o /dev/null -w '%{http_code} %{time_total}' -H "Authorization: Bearer $TOKEN" "$BASE_URL$path" || true)
    else
      status=$(curl -sS -o /dev/null -w '%{http_code} %{time_total}' "$BASE_URL$path" || true)
    fi
    elapsed=$(awk '{printf "%.1f", $2*1000}' <<<"$status")
    printf '%s\n' "$elapsed" >> "$tmp"
    statuses+=" $(awk '{print $1}' <<<"$status")"
    sleep 0.1
  done
  python3 - "$label" "$path" "$statuses" "$tmp" <<'PY'
import sys
label, path, statuses, tmp = sys.argv[1:5]
vals=[float(x.strip()) for x in open(tmp) if x.strip()]
first=vals[0]
warm=vals[1:]
print(f"{label} path={path} statuses={statuses.strip()} avg_ms={sum(vals)/len(vals):.1f} first_ms={first:.1f} warm_avg_ms={(sum(warm)/len(warm) if warm else 0):.1f} min_ms={min(vals):.1f} max_ms={max(vals):.1f}")
PY
  rm -f "$tmp"
}

login_if_needed

measure health /health public
measure teams /v1/teams public
measure peas /v1/peas public
measure operation-centers /v1/operation-centers public

if [[ -n "$TOKEN" ]]; then
  measure auth-me /v1/auth/me auth
  measure dashboard-summary '/v1/dashboard/summary?year=2026' auth
  measure monthly-plan-status /v1/monthly-plans/2026/6/status auth
  measure monthly-plan-overview /v1/monthly-plans/2026/overview auth
else
  echo "Skipping authenticated endpoints: set TOKEN or USERNAME/PASSWORD. Secrets are never printed."
fi
