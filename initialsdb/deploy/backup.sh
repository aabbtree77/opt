#!/bin/bash
set -euo pipefail

APP_NAME=initialsdb

BACKUP_DIR="/opt/${APP_NAME}/backups"
TS="$(date +%Y%m%d_%H%M%S)"
OUT="${BACKUP_DIR}/${APP_NAME}_${TS}.sql.gz"

mkdir -p "$BACKUP_DIR"

CONTAINER="${APP_NAME}-db"

if ! docker ps --format '{{.Names}}' | grep -qx "$CONTAINER"; then
	echo "❌ Container $CONTAINER not running"
	exit 1
fi

docker exec "$CONTAINER" \
	pg_dump -U app app | gzip > "$OUT"

echo "✅ Backup created: $OUT"

# Keep last 14 days only
find "$BACKUP_DIR" -type f -mtime +14 -delete
