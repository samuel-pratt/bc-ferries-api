# BC Ferries API

[bcferriesapi.ca](https://bcferriesapi.ca) provides free JSON data for BC Ferries schedules, sailing capacity, and service notices. No API key is needed. This is an unofficial API; data can be delayed or unavailable.

## Using the API

```sh
curl https://bcferriesapi.ca/v2/
```

| Endpoint | Returns |
| --- | --- |
| `GET /v2/` | `capacityRoutes`, `nonCapacityRoutes`, and `specialServices` |
| `GET /v2/capacity/` | Routes with capacity data, in a `routes` array |
| `GET /v2/noncapacity/` | Collected schedules in `routes`, plus `specialServices` |
| `GET /api/` | Legacy upcoming sailings for a limited set of terminal pairs |
| `GET /api/{departure-terminal}/` | Legacy sailings from one terminal |
| `GET /api/{departure-terminal}/{destination-terminal}/` | Legacy sailings for one terminal pair |
| `GET /healthcheck/` | Server liveness |
| `GET /readyz/` | Database access and availability of cached capacity sailings (not a freshness check) |
| `GET /` | Website |

Routes include terminal codes, sailing duration, and a `sailings` array. Sailings include departure/arrival times, vessel details, and capacity percentages where available. Additional fields provide notices, terminal details, sailing status, and observation times. Notices include an optional plain-text `body` and `bodyObservedAt` when the official notice can be fetched; keep the `url` as a fallback. Notice text is cached for up to two minutes. V2 has no per-route endpoint: filter the returned arrays by `routeCode` (for example, `TSASWB`).

Sailing times use BC local time; `serviceDate` and `arrivalDate` identify their dates. `observedAt` timestamps use UTC. For compatibility, capacity `time` becomes the actual departure after sailing; `scheduledTime` preserves the original time. `actualDepartureTime`, `estimatedArrivalTime`, and `actualArrivalTime` are separate when available. Schedule `time`/`arrivalTime` remain scheduled times; `duration` and `stops` describe each trip.

Capacity `fill`, `carFill`, and `oversizeFill` are percentages **full**, not space remaining. Use their respective `capacityKnown`, `carCapacityKnown`, and `oversizeCapacityKnown` flags: legacy numeric values alone cannot distinguish missing data from zero. Current conditions refresh every minute; schedules refresh every four hours and just after midnight Pacific time.

Check `meta.stale`, `meta.complete`, `meta.issues`, and route `dataStatus` before using cached data. `no-service` means a verified empty schedule for `serviceDate`; `unavailable` means collection failed or the source did not publish a usable schedule, and older cached sailings may remain. Fresh data is not necessarily complete or correct upstream. Ignore unfamiliar fields and handle empty results. When no cached capacity sailings exist, `/v2/capacity/` returns the JSON string `"BC Ferries Data Currently Down"`. Responses allow cross-origin requests.

For example, `https://bcferriesapi.ca/api/TSA/SWB/` returns Tsawwassen–Swartz Bay sailings.

### Coverage

V2 includes Cortes Island–Heriot Bay (`COR` → `HRB`) and Tsawwassen–Departure Bay (`TSA` → `NAN`). Route discovery checks the official indexes; terminal codes are available in the responses. Legacy `/api/` coverage is unchanged.

Schedule routes also expose `currentConditions` with live sailings, notices, terminal details, and its own observation time. `not-published` means BC Ferries does not expose a live sailing table for that direction. Matching live sailings update schedule status without replacing scheduled times; `statusObservedAt` identifies that observation. `unknown` is not a confirmed departure or cancellation.

The independently operated [Gambier/Keats service](https://www.bcferries.com/routes-fares/schedules/gambier-keats) appears in `specialServices`, using its published `columns`, `rows`, date range, and `notes` rather than invented terminal codes. Read the weekday/holiday column headings and request-only/connection footnotes; these are not unconditional departures. Expired or unavailable timetables are flagged.

Dangerous-goods-only sailings are excluded from passenger schedules. Data depends on what BC Ferries publishes; it is not a booking, fare, or guaranteed real-time vessel-position service. `/readyz/` checks availability, not data completeness.

Check [BC Ferries schedules](https://www.bcferries.com/routes-fares/schedules) and [current conditions](https://www.bcferries.com/current-conditions) for travel decisions.

## Running it

Requires Docker with Docker Compose. For a new installation:

```sh
git clone https://github.com/samuel-pratt/bc-ferries-api.git
cd bc-ferries-api
cp .env.example .env
```

Set a database password in `.env`, then start the API and database:

```sh
chmod 600 .env
docker volume create bcferries_pg17
docker compose up --build -d
```

The website is at `http://localhost:8081/` and the API is at `http://localhost:8081/v2/`. `/healthcheck/` checks that the server is running; `/readyz/` checks database access and cached capacity data. A fresh install needs time to collect sailings.

Use `docker compose logs -f api` for logs and `docker compose down` to stop without deleting data. For public hosting, put an HTTPS reverse proxy in front of port 8081, which binds to localhost.

**Existing installations:** keep your existing credentials, Compose project name, `BC_DB_VOLUME`, and PostgreSQL major version in `POSTGRES_IMAGE`. Back up first with `bash scripts/backup.sh`. Do not point a newer PostgreSQL major at an old volume or use `docker compose down -v`. The API applies its metadata migration automatically.

Run `bash scripts/verify.sh` for tests using disposable containers.

## Used by

- [BC Ferry Times](https://apps.apple.com/ca/app/id1615899209)
- [MMM BC Ferries](https://github.com/stonecrown/MMM-BCFerries)
- [Cascadia Crossing](https://apps.apple.com/app/1643019956)
