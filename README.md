# BC Ferries API - [bcferriesapi.ca](https://bcferriesapi.ca)

The BC Ferries API provides current data on BC Ferries sailings and schedules for all routes.

---

## Prerequisites

Ensure the following are installed on your system:

1. [Docker](https://www.docker.com/products/docker-desktop)
2. [Docker Compose](https://docs.docker.com/compose/)
3. [Go (1.19+)](https://go.dev/dl/) (optional for local development)

---

## Setup

### 2. Clone the repository

```
git clone https://github.com/samuel-pratt/bc-ferries-api.git
cd bc-ferries-api
```

### 2. `.env` File

Create a `.env` file in the project root from the `.env.sample`. Below is an example:

```env
# Database Configuration
DB_USER=username
DB_PASS=password
DB_NAME=dbname
DB_HOST=db
DB_PORT=5432
DB_SSL=disable
```

### 3. Build and start the container

```
docker-compose up --build
```

This will:

- Start a PostgreSQL database service (db).
- Build and run the Go application (api).

Visit these routes to test if setup was successful:

http://localhost:8080/healthcheck/ (API health check)
http://localhost:8080/v2/ (Main endpoint)


---

## API Reference

### V2

Version 2 of the API includes data for all terminals and routes served by BC Ferries. The response is structured as an array of "route" objects, each defining departure and arrival terminals, along with a JSON object containing sailings for that specific route.

#### Endpoints:

- Root Endpoint: `https://www.bcferriesapi.ca/v2/`
- Capacity Endpoint: `https://www.bcferriesapi.ca/v2/capacity/`
- Non-Capacity Endpoint: `https://www.bcferriesapi.ca/v2/noncapacity/`

The root `/v2/` route provides data for both capacity and non-capacity sailings. Non-capacity includes information on all BC Ferries routes, while capacity data covers routes with vessel fill data reported by BC Ferries.

#### Capacity Route Codes:

- **"TSA"**: Routes to terminals "SWB", "SGI", "DUK"
- **"SWB"**: Routes to terminals "TSA", "FUL", "SGI"
- **"HSB"**: Routes to terminals "NAN", "LNG", "BOW"
- **"DUK"**: Route to terminal "TSA"
- **"LNG"**: Route to terminal "HSB"
- **"NAN"**: Route to terminal "HSB"

### V1

The old version of this API uses the following route codes used by BC Ferries:

- "TSA" -> Tsawwassen
- "SWB" -> Swartz Bay
- "SGI" -> Southern Gulf Islands
- "DUK" -> Duke Point (Nanaimo)
- "FUL" -> Fulford Harbour (Salt Spring Island)
- "HSB" -> Horseshoe Bay
- "NAN" -> Departure Bay (Nanaimo)
- "LNG" -> Langdale
- "BOW" -> Bowen Island

The API endpoint format is:

`https://www.bcferriesapi.ca/api/<departure-terminal>/<destination-terminal>`

You can specify departure and destination terminals to get relevant data.

#### Available Departure Terminals:

- "TSA", "SWB", "HSB", "DUK", "LNG", "NAN", "FUL", "BOW"

#### Available Destination Terminals:

Please note that the destination terminal must correspond to the departure terminal, incorrect pairings will return an error.

- "TSA": ["SWB", "SGI", "DUK"]
- "SWB": ["TSA", "FUL", "SGI"]
- "HSB": ["NAN", "LNG", "BOW"]
- "DUK": ["TSA"]
- "LNG": ["HSB"]
- "NAN": ["HSB"]
- "FUL": ["SWB"]
- "BOW": ["HSB"]

## Used By

Projects using the BC Ferries API:

- [BC Ferry Times](https://apps.apple.com/ca/app/id1615899209): iOS app for the latest ferry schedules and capacities.
- [MMM BC Ferries](https://github.com/stonecrown/MMM-BCFerries): MagicMirror2 module for BC Ferries route info.
- [Cascadia Crossing](https://apps.apple.com/app/1643019956): iOS app for border crossing times, ferry schedules, and capacities.
