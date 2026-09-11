#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
# A unique, disposable Compose project: never references production .env or volumes.
project="release-check-bc-$$"
compose=(docker compose --env-file /dev/null -p "$project" -f compose.test.yml)
cleanup() { "${compose[@]}" down --remove-orphans; }
trap cleanup EXIT
"${compose[@]}" build tests
"${compose[@]}" run --rm tests
docker build -t bc-ferries-api:release-check .
docker run --rm --network none --cap-drop ALL --security-opt no-new-privileges --entrypoint chromium bc-ferries-api:release-check --headless --no-sandbox --disable-dev-shm-usage --dump-dom 'data:text/html,<html><body>release-smoke</body></html>' | grep -q release-smoke
