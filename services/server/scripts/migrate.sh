#!/bin/bash

set -e

DB_URL="${DATABASE_URL:-postgres://tunnelkit:tunnelkit@localhost:5432/tunnelkit?sslmode=disable}"
MIGRATIONS_DIR="$(dirname "$0")/migrations"

# Wait for DB
echo "Waiting for database..."
until pg_isready -d "$DB_URL" 2>/dev/null || true; do
  sleep 1
done

echo "Running migrations..."
for f in "$MIGRATIONS_DIR"/*.up.sql; do
  if [ -f "$f" ]; then
    echo "Running $f..."
    psql "$DB_URL" -f "$f"
  fi
done

echo "Migrations complete!"
