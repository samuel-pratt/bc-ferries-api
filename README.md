# BC Ferries API

[bcferriesapi.ca](https://bcferriesapi.ca) is a free, read-only JSON API for current BC Ferries schedules, vessel capacity, service notices, and sailing status. No API key is required.

> [!IMPORTANT]
> This is an independent project and is not affiliated with BC Ferries. Data reflects the BC Ferries website when it was collected and may be delayed or unavailable. Check the response freshness metadata before using it for time-sensitive travel decisions.

## Quick start

```sh
curl https://bcferriesapi.ca/v2/capacity/
```

All public API responses use JSON and allow cross-origin requests.

## API reference

### Version 2

V2 is the current API. It includes every route served by BC Ferries and adds operational and freshness information without removing fields expected by existing clients.

| Method | Endpoint | Description |
| --- | --- | --- |
| `GET` | [`/v2/`](https://bcferriesapi.ca/v2/) | Capacity and non-capacity routes |
| `GET` | [`/v2/capacity/`](https://bcferriesapi.ca/v2/capacity/) | Routes for which BC Ferries reports vessel fill data |
| `GET` | [`/v2/noncapacity/`](https://bcferriesapi.ca/v2/noncapacity/) | Remaining routes and schedules |
| `GET` | [`/healthcheck/`](https://bcferriesapi.ca/healthcheck/) | Basic service health check |

The combined endpoint has this top-level shape:

```json
{
  "capacityRoutes": [],
  "nonCapacityRoutes": [],
  "meta": {
    "schemaVersion": "2.1",
    "generatedAt": "2026-08-21T12:00:00Z",
    "stale": false,
    "sources": {}
  }
}
```

The capacity and non-capacity endpoints each return a `routes` array and the same `meta` object. The example timestamps and empty arrays above illustrate structure only.

### Route fields

Every V2 route can include:

| Field | Type | Description |
| --- | --- | --- |
| `routeCode` | string | Identifier for the terminal pair |
| `fromTerminalCode` | string | Departure terminal code |
| `toTerminalCode` | string | Arrival terminal code |
| `sailingDuration` | string | Published crossing duration |
| `sailings` | array | Scheduled sailings for the route |
| `observedAt` | timestamp | When this route was last observed |
| `notices` | array | Service notices affecting the route, when available |
| `terminal` | object | Terminal details and webcam metadata, when available |

Capacity sailings include `time`, `arrivalTime`, `sailingStatus`, `fill`, `carFill`, `oversizeFill`, `vesselName`, and `vesselStatus`. Non-capacity sailings omit the fill fields. Either kind may also include:

- `status`: normalized `scheduled`, `delayed`, `departed`, `arrived`, or `cancelled` state.
- `statusText`: the source wording shown by BC Ferries.
- `delayMinutes`: a parsed delay, when one can be determined.
- `observedAt`: when the source information was observed.

Retain `statusText` when displaying operational information: normalized fields are interpretations of the public website, not an official operational feed.

### Freshness metadata

The `meta` object describes the whole response:

| Field | Description |
| --- | --- |
| `schemaVersion` | Response schema version |
| `generatedAt` | Time the API assembled the response |
| `stale` | Whether one or more required sources are older than their freshness threshold |
| `sources` | Per-source status, observation time, and last successful collection time |

Each source reports a `status` of `fresh`, `stale`, or `unavailable`. Consumers should check both `meta.stale` and the relevant route or sailing `observedAt` value.

JSON Schemas are available in [`schemas/`](schemas/):

- [`bc-ferries-api-v2-schema.json`](schemas/bc-ferries-api-v2-schema.json)
- [`bc-ferries-api-v2-capacity-schema.json`](schemas/bc-ferries-api-v2-capacity-schema.json)
- [`bc-ferries-api-v2-noncapacity-schema.json`](schemas/bc-ferries-api-v2-noncapacity-schema.json)

### Capacity terminal codes

| From | Supported destinations |
| --- | --- |
| `TSA` | `SWB`, `SGI`, `DUK` |
| `SWB` | `TSA`, `FUL`, `SGI` |
| `HSB` | `NAN`, `LNG`, `BOW` |
| `DUK` | `TSA` |
| `LNG` | `HSB` |
| `NAN` | `HSB` |

### Errors and availability

The service currently returns JSON for successful requests and permits CORS with `Access-Control-Allow-Origin: *`. If capacity data is temporarily unavailable, the capacity endpoint may return the JSON string `"BC Ferries Data Currently Down"`. Clients should tolerate missing optional fields and unavailable or stale source data.

### Version 1 (legacy)

V1 remains available for older clients:

```text
GET /api/
GET /api/{departure-terminal}/
GET /api/{departure-terminal}/{destination-terminal}/
```

Example:

```sh
curl https://bcferriesapi.ca/api/TSA/SWB/
```

Supported codes are `TSA`, `SWB`, `SGI`, `DUK`, `FUL`, `HSB`, `NAN`, `LNG`, and `BOW`. New integrations should use V2.

## Run locally

### Requirements

- Docker
- Docker Compose
- Go 1.19 or newer, only when running outside Docker

Clone the repository and create the environment file:

```sh
git clone https://github.com/samuel-pratt/bc-ferries-api.git
cd bc-ferries-api
cp .env.example .env
```

Replace the example database password in `.env`, then start the services:

```sh
docker compose up --build -d
```

The default local endpoints are:

- `http://localhost:8081/healthcheck/`
- `http://localhost:8081/v2/`

Follow logs with `docker compose logs -f api`. Stop the stack without deleting its database volume with `docker compose down`.

The API creates a separate `route_operational_data` table idempotently at startup, preserving the five-column legacy route tables for rollback. [`migrations/003_operational_data.sql`](migrations/003_operational_data.sql) is also available for operators who run migrations separately. The migration role needs `CREATE` in the schema. See [DEPLOYMENT.md](DEPLOYMENT.md) for the controlled migration, database upgrade, and rollback procedure.

## Used by

- [BC Ferry Times](https://apps.apple.com/ca/app/id1615899209) — iOS app for ferry schedules and capacities.
- [MMM BC Ferries](https://github.com/stonecrown/MMM-BCFerries) — MagicMirror² route information module.
- [Cascadia Crossing](https://apps.apple.com/app/1643019956) — iOS app for border waits, ferry schedules, and capacities.

If you build something with the API, feel free to open an issue or pull request to add it here.
