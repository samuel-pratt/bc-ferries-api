# BC Ferries deployment

Deploy this repository independently. Its API binds to **127.0.0.1:8081**; put the existing HTTPS reverse proxy in front of it. Do not replace your working proxy configuration or expose PostgreSQL.

## Before deploying an existing installation

1. Record the current commit and API image ID (`docker compose images -q api`), and tag that image with a rollback tag. Save the old Compose file and protected environment file outside Git.
2. Identify the real database container and its volume with `docker inspect <database-container> --format '{{json .Mounts}}'`. Check its PostgreSQL major with `docker exec <database-container> postgres --version`.
3. Set `BC_DB_VOLUME` to that exact volume and `POSTGRES_IMAGE` to the same major. Keep the existing credentials/database name. **Do not apply the example's PostgreSQL 17 image to a PostgreSQL 13 volume.**
4. Run `bash scripts/backup.sh` using the correct Compose project. Copy the resulting dump off the droplet, and test restoring it into a separate database.
5. Build and test before switching traffic:

```sh
bash scripts/verify.sh
docker compose config --quiet
docker compose build --pull api
docker compose up -d db
docker compose run --rm api migrate
docker compose up -d --no-deps api
curl --fail http://127.0.0.1:8081/healthcheck/
curl --fail http://127.0.0.1:8081/readyz/
curl --fail http://127.0.0.1:8081/v2/
docker compose logs --tail=100 api
```

Do not print `docker compose config` without `--quiet`: the rendered configuration includes secrets. Existing deployments must keep the original Compose project name (use `COMPOSE_PROJECT_NAME` if necessary).

Readiness requires some cached capacity sailings; it is not proof every route is fresh. Verify route counts and the oldest route observation against the previous deployment. Check both capacity and non-capacity routes, and check that schedules refresh after Chrome runs. Compare legacy `/api/` and `/v2/` response fields used by your clients. A failed scrape retains old data; clients should respect `meta.stale` and each route's `observedAt`. This remains unofficial scraped information, not a navigation or safety system.

## New installation

Copy `.env.example` to `.env`, use a unique password (`openssl rand -hex 32`), and `chmod 600 .env`. Keep the example's PostgreSQL 17 setting, explicitly create the volume with `docker volume create bcferries_pg17`, then follow the build/start commands above. `init.sql` initializes the two route tables only on an empty volume.

## Application database role

After initial migrations, provision a non-superuser runtime role:

```sh
docker compose cp scripts/app-role.sql db:/tmp/app-role.sql
docker compose exec db sh -c 'psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -f /tmp/app-role.sql'
```

The script prompts securely for a password. In `.env`, preserve the original administrator credentials as `DB_ADMIN_USER` and `DB_ADMIN_PASS`; set `DB_USER=ferries_app` and `DB_PASS` to the new role password. Recreate the API and verify readiness. The API never receives the administrator credentials. Future migrations that require ownership changes should run through the administrator, followed by updated grants.

## Database safety and rollback

The operational-data migration creates **a separate table**. It does not add columns to or rewrite the legacy five-column route tables. Writes update route data and metadata in one transaction. Existing rows without metadata remain readable and are reported as stale until refreshed. The integration test checks an old-style five-column `SELECT *` after migration.

If an earlier preview already added columns to the original tables, that compatibility guarantee does not apply: inspect the table definitions first. Do not drop columns automatically.

To roll back the API, use the saved image and Compose file with the same database major and volume. Leave the companion table in place. Do not run `down -v`, prune volumes, or restore a backup over the live database as a routine rollback.

PostgreSQL 13 is end-of-life. Upgrade it as a **separate maintenance operation**: stop API writers, take a final dump, create a new PostgreSQL 17 volume, restore into a separately started empty database using `pg_restore --exit-on-error --no-owner --no-acl`, verify row counts and the application, then switch the configured volume/image. The standard init script creates tables, so restore into a separate empty database, not over those initialized tables. Keep the old stopped volume unchanged for rollback; never attach a new major directly to the old data directory. Recreate application-role grants after restore.

## Operations

- Schedule `bash scripts/backup.sh` daily via cron/systemd; alert on failures. Keep encrypted off-host copies. Test restoration regularly. The script validates archive structure, not a full restore.
- Alert on readiness failures, old observations, HTTP errors, memory pressure, disk usage, and backup age. Liveness alone cannot detect stale scraped data.
- One scraper process per database. Jobs do not overlap themselves; HTTP requests and browser navigations have deadlines.
- The API runs non-root with no Linux capabilities. Chromium's nested sandbox is disabled in Docker; container isolation remains enabled. Keep the host/runtime and image packages patched. Scraper destinations are fixed in code.
- Build on the droplet or CI, not all three builds simultaneously on a small server. Containers have bounded logs; reserve disk space for images, PostgreSQL and backups.
