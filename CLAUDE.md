# GO-Fuels — Project Context for LLMs

## What This Project Is

GO-Fuels is a Go backend service that tracks Italian fuel prices by ingesting data from the MISE (Ministry of Economic Development) public API. It stores time-series price data in TimescaleDB and reference/metadata in MongoDB, using RabbitMQ for async job processing.

## Architecture Overview

```
┌─────────────┐     ┌───────────┐     ┌─────────────┐
│  Echo API    │────▶│ RabbitMQ  │────▶│   Worker    │
│  (port 8080) │     │ (port 5672)│     │ (consumer)  │
└──────┬───────┘     └───────────┘     └──────┬──────┘
       │                                       │
       │  reads/writes                         │ fetches MISE API
       ▼                                       │ upserts prices
┌──────────────┐                               │
│   MongoDB    │◀──── reference data           │
│ (port 27017) │      (locations, brands,      │
└──────────────┘       fuel types, users)       │
                                               ▼
                                     ┌──────────────────┐
                                     │  TimescaleDB      │
                                     │  (port 6543)      │
                                     │  price time-series│
                                     │  + ingestion jobs │
                                     └──────────────────┘
```

### Three Execution Modes

1. **API Server** (`src/main.go`) — Echo HTTP server on :8080. Connects to all three datastores. Serves REST endpoints for fuels, locations, brands, users, and ingestion jobs.
2. **Worker** (`src/cmd/worker/main.go`) — Long-running RabbitMQ consumer. Connects to Postgres + RabbitMQ only. Picks up ingestion jobs from the queue, calls the MISE API, upserts fuel prices.
3. **Ingester CLI** (`src/cmd/ingester/main.go`) — One-shot script. Connects to Postgres only. Runs a single fetch+ingest for hardcoded coordinates. Useful for testing.

## Directory Structure

```
src/
├── main.go                          # API server entry point
├── models/models.go                 # All data models (shared by API + worker)
├── ingester/ingester.go             # MISE API client + ingestion logic
├── cmd/
│   ├── ingester/main.go             # One-shot ingester CLI
│   └── worker/main.go               # RabbitMQ job worker
├── routers/
│   ├── Router.go                    # Route registration hub
│   ├── FuelRouter.go                # /fuels/* handlers
│   ├── LocationRouter.go            # /location/* handlers
│   ├── UserRouter.go                # /user/* handlers
│   ├── BrandRouter.go               # /brand/* handlers
│   └── IngestionRouter.go           # /ingestion/* handlers
└── utils/
    ├── utils.go                     # Key generation helpers
    └── dbHandlers/
        ├── Handler.go               # Aggregated DB handler (Postgres + Mongo + Rabbit)
        ├── db/PostgresHandler.go    # TimescaleDB operations via GORM
        ├── db/MongoHandler.go       # MongoDB operations
        └── rabbit/rabbit.go         # RabbitMQ publish/consume
build/
├── docker-compose.yaml              # MongoDB + TimescaleDB + RabbitMQ
└── Dockerfile                       # Stub — not yet implemented
```

## Data Model

### MongoDB Collections (reference/metadata)

| Collection          | Model            | Key Field      | Purpose                                |
|---------------------|------------------|----------------|----------------------------------------|
| `locations`         | `Location`       | `location_key` | Station name, address, coordinates     |
| `brands`            | `Brand`          | `brand_key`    | Fuel company info                      |
| `fuel_types`        | `FuelType`       | `fuel_key`     | Fuel category, unit, description       |
| `user_preferences`  | `UserPreferences`| `user_id`      | User settings, notification prefs      |

### PostgreSQL/TimescaleDB Tables (time-series + jobs)

| Table            | Model          | Primary Key          | Purpose                              |
|------------------|----------------|----------------------|--------------------------------------|
| `fuel_data`      | `FuelData`     | `(id, last_update)`  | Price records, hypertable on `last_update` |
| `ingestion_jobs` | `IngestionJob` | `id`                 | Job lifecycle tracking               |

### Cross-DB References

MongoDB and PostgreSQL are linked by string keys — **not foreign keys**:
- `FuelData.location_key` → `Location.location_key` (MongoDB)
- `FuelData.fuel_key` → `FuelType.fuel_key` (MongoDB)
- `Location.brand_key` → `Brand.brand_key` (MongoDB)

## Ingestion Pipeline

### Flow

1. Client `POST /ingestion/jobs` with `{lat, lng, radius}`
2. API creates `IngestionJob` in Postgres (status: `START`)
3. API publishes `{job_id}` to RabbitMQ queue `ingestion_jobs`
4. Worker consumes message, loads job from Postgres
5. Worker sets status `RUNNING`, calls MISE API at `https://carburanti.mise.gov.it/ospzApi/search/zone`
6. Worker upserts each station+fuel combination into `fuel_data`
7. Worker sets status `DONE` (with record count) or `FAILED` (with error)

### MISE API Data Mapping

```
StationResult.ID        → location_key: "station_{id}"
FuelPrice.FuelID+IsSelf → fuel_key:     "fuel_{fuelId}_{isSelf}"
FuelPrice.Name          → fuel_category
FuelPrice.Price         → price
StationResult.InsertDate → last_update
```

### Job Status Lifecycle

```
START → RUNNING → DONE
                → FAILED
```

## API Endpoints

### Fuels (`/fuels`)
- `POST /types` — Create fuel type (MongoDB)
- `GET /types` — List all fuel types
- `GET /types/:fuel_key` — Get fuel type by key
- `GET /types/category/:category` — Filter by category
- `POST /data` — Create fuel price record (Postgres)
- `PUT /data` — Upsert fuel price
- `GET /data/location/:location_key` — Prices at location
- `GET /data/location/:location_key/fuel/:fuel_key` — Specific price
- `GET /data/history?location_key=&fuel_key=&since=` — Price history
- `GET /data/average/:fuel_key?since=` — Average price (default 7d)
- `GET /data/latest/:fuel_key?limit=` — Latest prices (default 10)
- `PATCH /data/price` — Update single price

### Locations (`/location`)
- `POST /` — Create location (MongoDB)
- `GET /:location_key` — Get location
- `PUT /:location_key` — Update location
- `GET /brand/:brand_key` — Locations by brand
- `GET /nearby?lat=&lng=&radius=` — Nearby locations (default 5km)
- `GET /:location_key/prices` — Location + enriched prices (cross-DB)
- `GET /nearby/prices?lat=&lng=&radius=&fuel_key=` — Nearby + prices

### Brands (`/brand`)
- `POST /` — Create brand
- `GET /:brand_key` — Get brand
- `GET /` — List all brands
- `PUT /:brand_key` — Update brand

### Users (`/user`)
- `POST /preferences` — Create preferences
- `GET /preferences/:user_id` — Get preferences
- `PUT /preferences/:user_id` — Update preferences

### Ingestion (`/ingestion`)
- `POST /jobs` — Create ingestion job (publishes to RabbitMQ)
- `GET /jobs/:id` — Get job status
- `GET /jobs?limit=&offset=` — List jobs (default limit 20)

## Tech Stack & Dependencies

| Component    | Technology                          | Purpose                     |
|--------------|-------------------------------------|-----------------------------|
| HTTP         | `echo/v4`                           | REST API framework          |
| ORM          | `gorm` + `gorm/driver/postgres`     | PostgreSQL access           |
| Time-series  | TimescaleDB (pg18-ha)               | Hypertable on `fuel_data`   |
| Documents    | `mongo-driver`                      | MongoDB access              |
| Queue        | `streadway/amqp`                    | RabbitMQ publish/consume    |
| Module       | `Irlandesee/GO-Fuels`               | Go 1.25.3                   |

## Environment Variables

| Variable      | Default                                    | Used By         |
|---------------|--------------------------------------------|-----------------|
| `DB_HOST`     | `localhost`                                | API, Worker     |
| `DB_PORT`     | `6543`                                     | API, Worker     |
| `DB_NAME`     | `postgres`                                 | API, Worker     |
| `DB_USER`     | `postgres`                                 | API, Worker     |
| `DB_PASS`     | `mysecretpassword`                         | API, Worker     |
| `MONGO_URI`   | `mongodb://admin:password@localhost:27017` | API only        |
| `MONGO_DB`    | `go_fuels`                                 | API only        |
| `RABBIT_URI`  | `amqp://guest:guest@localhost:5672/`       | API, Worker     |

## Known Issues / TODOs

- `models/models.go:124-134` — `GenerateFuelKey()` and `GenerateLocationKey()` are placeholder stubs (real implementations exist in `utils/utils.go` but models.go has duplicate stubs)
- `UpsertFuelData` in `PostgresHandler.go` uses GORM's `FirstOrCreate` with `Assign` — this does not update existing records on subsequent calls; needs to use `Save` or raw SQL `ON CONFLICT ... DO UPDATE`
- `build/Dockerfile` is a stub (runs `top -b`), needs proper multi-stage Go build
- `deploy/`, `utils/customError/`, `utils/serializer/` are empty placeholder directories

## Running Locally

```bash
# Start infrastructure
docker compose -f build/docker-compose.yaml up -d

# Run API server
cd src && go run main.go

# Run worker (separate terminal)
cd src && go run cmd/worker/main.go

# One-shot ingestion test
cd src && go run cmd/ingester/main.go
```
