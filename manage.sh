#!/bin/sh

set -eu

ROOT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
MODE=simple

if [ "${1:-}" = "simple" ] || [ "${1:-}" = "standard" ]; then
  MODE=$1
  shift
fi

COMMAND=${1:-help}
if [ "$#" -gt 0 ]; then
  shift
fi

COMPOSE_FILE="$ROOT_DIR/deploy/$MODE/compose.yaml"
ENV_FILE="$ROOT_DIR/.env"

random_secret() {
  if command -v openssl >/dev/null 2>&1; then
    openssl rand -hex 32
    return
  fi
  od -An -N32 -tx1 /dev/urandom | tr -d ' \n'
}

init_env() {
  if [ -f "$ENV_FILE" ]; then
    chmod 600 "$ENV_FILE"
    echo "Existing .env kept unchanged: $ENV_FILE"
    return
  fi

  umask 077
  jwt_secret=$(random_secret)
  settings_key=$(random_secret)
  meili_key=$(random_secret)
  mysql_password=$(random_secret)
  mysql_root_password=$(random_secret)
  {
    echo "PAOPAO_BIND_ADDRESS=127.0.0.1"
    echo "PAOPAO_PORT=8008"
    echo "PAOPAO_PUBLIC_HOST=localhost:8008"
    echo "PAOPAO_PUBLIC_SECURE=false"
    echo "PAOPAO_JWT_SECRET=$jwt_secret"
    echo "PAOPAO_ADMIN_SETTINGS_KEY=$settings_key"
    echo "PAOPAO_MEILI_MASTER_KEY=$meili_key"
    echo "PAOPAO_MYSQL_PASSWORD=$mysql_password"
    echo "PAOPAO_MYSQL_ROOT_PASSWORD=$mysql_root_password"
    echo "PAOPAO_IMAGE=paopao-ce:local"
    echo "REDIS_TAG=7.2-alpine"
    echo "MEILI_TAG=v1.49.0"
    echo "MYSQL_TAG=8.0"
  } >"$ENV_FILE"
  chmod 600 "$ENV_FILE"
  echo "Created secure deployment settings: $ENV_FILE"
  echo "Back up this file. Its admin-settings key is required to decrypt saved secrets."
}

env_value() {
  awk -v prefix="$1=" '
    index($0, prefix) == 1 { value = substr($0, length(prefix) + 1) }
    END { print value }
  ' "$ENV_FILE"
}

require_secret() {
  name=$1
  value=$(env_value "$name")
  case "$value" in
    ""|replace-*|REPLACE-*|change-me*|CHANGE-ME*|OVERRIDDEN-*)
      echo "Unsafe or placeholder value for $name in $ENV_FILE." >&2
      echo "Remove .env and run ./manage.sh init, or replace it with a random value." >&2
      exit 1
      ;;
  esac
  if [ "${#value}" -lt 32 ]; then
    echo "$name must contain at least 32 characters." >&2
    exit 1
  fi
}

require_bool() {
  name=$1
  value=$(env_value "$name")
  case "$value" in
    true|false) ;;
    *)
      echo "$name must be true or false in $ENV_FILE." >&2
      exit 1
      ;;
  esac
}

validate_env() {
  chmod 600 "$ENV_FILE"
  require_secret PAOPAO_JWT_SECRET
  require_secret PAOPAO_ADMIN_SETTINGS_KEY
  require_secret PAOPAO_MEILI_MASTER_KEY
  if [ "$MODE" = "standard" ]; then
    require_secret PAOPAO_MYSQL_PASSWORD
    require_secret PAOPAO_MYSQL_ROOT_PASSWORD
  fi
  require_bool PAOPAO_PUBLIC_SECURE
}

# Compose gives the caller's shell higher interpolation priority than
# --env-file. Reload protected and deployment-owned values explicitly so a
# stale or hostile shell variable cannot bypass the validation above.
load_managed_environment() {
  PAOPAO_JWT_SECRET=$(env_value PAOPAO_JWT_SECRET)
  PAOPAO_ADMIN_SETTINGS_KEY=$(env_value PAOPAO_ADMIN_SETTINGS_KEY)
  PAOPAO_MEILI_MASTER_KEY=$(env_value PAOPAO_MEILI_MASTER_KEY)
  PAOPAO_PUBLIC_HOST=$(env_value PAOPAO_PUBLIC_HOST)
  PAOPAO_PUBLIC_SECURE=$(env_value PAOPAO_PUBLIC_SECURE)
  export PAOPAO_JWT_SECRET PAOPAO_ADMIN_SETTINGS_KEY PAOPAO_MEILI_MASTER_KEY
  export PAOPAO_PUBLIC_HOST PAOPAO_PUBLIC_SECURE

  if [ "$MODE" = "standard" ]; then
    PAOPAO_MYSQL_PASSWORD=$(env_value PAOPAO_MYSQL_PASSWORD)
    PAOPAO_MYSQL_ROOT_PASSWORD=$(env_value PAOPAO_MYSQL_ROOT_PASSWORD)
    export PAOPAO_MYSQL_PASSWORD PAOPAO_MYSQL_ROOT_PASSWORD
  fi
}

set_build_metadata() {
  if command -v git >/dev/null 2>&1 && git -C "$ROOT_DIR" rev-parse --is-inside-work-tree >/dev/null 2>&1; then
    : "${PAOPAO_VERSION:=$(git -C "$ROOT_DIR" describe --tags --always --dirty)}"
    : "${PAOPAO_COMMIT:=$(git -C "$ROOT_DIR" rev-parse --short HEAD)}"
    : "${PAOPAO_BUILD_DATE:=$(git -C "$ROOT_DIR" show -s --format=%cI HEAD)}"
  else
    : "${PAOPAO_VERSION:=local}"
    : "${PAOPAO_COMMIT:=unknown}"
    : "${PAOPAO_BUILD_DATE:=unknown}"
  fi
  export PAOPAO_VERSION PAOPAO_COMMIT PAOPAO_BUILD_DATE
}

compose() {
  if [ ! -f "$ENV_FILE" ]; then
    echo "Missing .env. Run ./manage.sh init first." >&2
    exit 1
  fi
  validate_env
  load_managed_environment
  set_build_metadata
  docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" "$@"
}

case "$COMMAND" in
  init)
    init_env
    validate_env
    ;;
  up)
    init_env
    compose up -d --build --wait --wait-timeout 240
    compose ps
    ;;
  down)
    compose down --remove-orphans
    ;;
  restart)
    compose restart app
    compose up -d --no-deps --no-recreate --wait --wait-timeout 240 app
    compose ps
    ;;
  status|ps)
    compose ps
    ;;
  logs)
    if [ "$#" -eq 0 ]; then
      set -- app
    fi
    compose logs --tail=200 -f "$@"
    ;;
  config|check)
    init_env
    compose config --quiet
    echo "$MODE deployment configuration is valid."
    ;;
  promote-admin)
    if [ "$#" -ne 1 ]; then
      echo "Usage: ./manage.sh [simple|standard] promote-admin USERNAME" >&2
      exit 1
    fi
    compose exec -T app /app/paopao admin promote --username "$1"
    ;;
  help)
    echo "Usage: ./manage.sh [simple|standard] COMMAND"
    echo "Commands: init, up, down, restart, status, logs [service], check, promote-admin USERNAME"
    echo "Use restart for saved admin settings; it waits for app health before returning."
    echo "Run up again after changing .env or Compose settings."
    ;;
  *)
    echo "Unknown command: $COMMAND" >&2
    echo "Usage: ./manage.sh [simple|standard] COMMAND" >&2
    exit 2
    ;;
esac
