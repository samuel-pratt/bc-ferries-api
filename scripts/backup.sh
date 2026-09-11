#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
umask 077
mkdir -p backups
backup_tmp="$(mktemp "./backups/database-$(date -u +%Y%m%dT%H%M%SZ)-XXXXXX")"
# Incomplete dumps keep their temporary name and never get the .dump suffix.
docker compose exec -T db sh -c 'exec pg_dump -U "$POSTGRES_USER" -d "$POSTGRES_DB" --format=custom --no-owner --no-acl' > "$backup_tmp"
docker compose exec -T db pg_restore --list < "$backup_tmp" >/dev/null
mv "$backup_tmp" "$backup_tmp.dump"
echo "$backup_tmp.dump"
