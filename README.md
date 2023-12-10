# BC Ferries API - [bcferriesapi.ca](https://bcferriesapi.ca)

The BC Ferries API provides current data on BC Ferries sailings and schedules for all routes.

## Running Locally

To run locally you must have a postgres database set up with the tables in `db_setup.sql` included. Create a `.env` file from the `.ev.sample` and fill it with the database info.

Once that's set up just build the project with:

```
go build
```

And run it with:

```
./bc-ferries-api
```

And you're good to go!

## API Reference

### V2

Version 2 of the API encompasses data for all terminals and routes served by BC Ferries. The response is structured as an array of "route" objects, each defining departure and arrival terminals, along with a JSON object containing sailings for that specific route.

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

This API uses the following route codes used by BC Ferries:

- "TSA" -> Tsawwassen
- "SWB" -> Swartz Bay
- "SGI" -> Southern Gulf Islands
- "DUK" -> Duke Point (Nanaimo)
- "FUL" -> Fulford Harbour (Salt Spring Island)
- "HSB" -> Horseshoe Bay
- "NAN" -> Departure Bay (Nanaimo)
- "LNG" -> Langford
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
