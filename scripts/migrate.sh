#!/usr/bin/env bash
set -euo pipefail

# Hotline database migrations. Goose is the schema owner for deployed DBs.
# Usage:
#   ./scripts/migrate.sh config up      # use config.yaml + .env DB_PASSWORD
#   ./scripts/migrate.sh prod up        # require GOOSE_DBSTRING or DATABASE_URL
#   ./scripts/migrate.sh config status
#   ./scripts/migrate.sh prod status

TARGET="${1:-config}"
ACTION="${2:-up}"
MIGRATION_DIR="${GOOSE_MIGRATION_DIR:-pkg/db/migrations}"

cd "$(dirname "$0")/.."

if [[ -f .env ]]; then
  set -a
  # shellcheck disable=SC1091
  source .env
  set +a
fi

if ! command -v goose >/dev/null 2>&1; then
  echo "goose CLI is required. Install with:" >&2
  echo "  go install github.com/pressly/goose/v3/cmd/goose@latest" >&2
  exit 127
fi

# Default: GOOSE_DRIVER=postgres
export GOOSE_DRIVER="${GOOSE_DRIVER:-postgres}"

config_value() {
  local key="$1"
  awk -v wanted="$key" '
    /^database:/ { in_db=1; next }
    /^[^[:space:]]/ { in_db=0 }
    in_db && $1 == wanted ":" {
      sub(/^[[:space:]]*[^:]+:[[:space:]]*/, "")
      sub(/[[:space:]]+#.*/, "")
      gsub(/^"|"$/, "")
      print
      exit
    }
  ' config.yaml
}

build_dbstring_from_config() {
  local host port user password dbname sslmode
  host="$(config_value host)"
  port="$(config_value port)"
  user="$(config_value user)"
  dbname="$(config_value dbname)"
  sslmode="$(config_value sslmode)"
  password="$(config_value password)"

  if [[ "$password" == '${DB_PASSWORD}' ]]; then
    password="${DB_PASSWORD:-}"
  fi
  if [[ -z "$host" || -z "$port" || -z "$user" || -z "$password" || -z "$dbname" || -z "$sslmode" ]]; then
    echo "cannot build DB string from config.yaml/.env; set GOOSE_DBSTRING or DATABASE_URL" >&2
    exit 2
  fi

  printf 'host=%s port=%s user=%s password=%s dbname=%s sslmode=%s' "$host" "$port" "$user" "$password" "$dbname" "$sslmode"
}

case "$TARGET" in
  config|dev|local)
    export GOOSE_DBSTRING="${GOOSE_DBSTRING:-${DATABASE_URL:-$(build_dbstring_from_config)}}"
    ;;
  prod|production)
    if [[ -z "${GOOSE_DBSTRING:-}" && -z "${DATABASE_URL:-}" ]]; then
      echo "prod migration requires explicit GOOSE_DBSTRING or DATABASE_URL" >&2
      exit 2
    fi
    export GOOSE_DBSTRING="${GOOSE_DBSTRING:-$DATABASE_URL}"
    ;;
  *)
    echo "unknown target '$TARGET' (expected config|dev|local|prod)" >&2
    exit 2
    ;;
esac

echo "Running goose ${ACTION} on target=${TARGET} dir=${MIGRATION_DIR}"
goose -dir "$MIGRATION_DIR" "$ACTION"
