#!/bin/sh
set -e

if [ "$STORAGE" = "postgres" ]; then
  : "${POSTGRES_HOST:?POSTGRES_HOST is required for postgres mode}"
  : "${POSTGRES_PORT:?POSTGRES_PORT is required for postgres mode}"
  : "${POSTGRES_DB:?POSTGRES_DB is required for postgres mode}"
  : "${POSTGRES_USER:?POSTGRES_USER is required for postgres mode}"
  : "${POSTGRES_PASSWORD:?POSTGRES_PASSWORD is required for postgres mode}"

  export GOOSE_DRIVER=postgres
  export GOOSE_DBSTRING="postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=disable"
  export GOOSE_MIGRATION_DIR=/app/migrations
  export GOOSE_TABLE=goose_db_version

  echo "Running migrations..."
  goose up
fi

exec /urlshortener "$@"

