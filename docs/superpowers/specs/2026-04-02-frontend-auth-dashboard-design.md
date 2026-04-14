# GO-Fuels Frontend, Auth & Dashboard — Design Spec

**Date:** 2026-04-02
**Status:** Approved

## Overview

Add a React frontend to GO-Fuels that lets authenticated users manage fuel price preferences and visualize historical trends. Authentication via Keycloak (OIDC/JWT), geocoding via Nominatim, user data on MongoDB.

## Architecture

```
React SPA (Vite, :5173)
    │
    ├── OIDC redirect ──→ Keycloak (:8180)
    │
    └── API calls ──→ Backend Go (Echo, :8080)
                           │
                           ├── JWT validation (JWKS from Keycloak)
                           ├── Nominatim geocoding (rate-limited, cached)
                           ├── MongoDB (users, locations, brands, fuel_types)
                           ├── TimescaleDB (fuel_data, ingestion_jobs)
                           └── RabbitMQ (ingestion job queue)
```

The backend Go acts as BFF (Backend for Frontend). The React SPA calls only the Go backend — never external services directly. Keycloak is exposed only for OIDC login/register redirects.

## 1. Authentication — Keycloak + JWT

### Keycloak Setup

- Added to `docker-compose.yaml` on port `8180` (avoids conflict with Echo on `8080`)
- Image: `quay.io/keycloak/keycloak:26.2`, `start-dev` mode
- Realm `go-fuels` auto-imported from a JSON export file at startup
- Client `go-fuels-frontend`: public OIDC client, redirect to `http://localhost:5173/*`
- Client scopes: `email`, `profile`

### JWT Flow

1. User clicks Login in React SPA
2. `react-oidc-context` redirects to Keycloak login/register page
3. Keycloak returns access token + refresh token to the SPA
4. SPA sends `Authorization: Bearer <token>` on every API call
5. Backend Go validates JWT signature via Keycloak JWKS endpoint (`/realms/go-fuels/protocol/openid-connect/certs`)
6. No per-request call to Keycloak — signature verification only

### User Sync (Keycloak to MongoDB)

On first authenticated API call, if no user document exists in MongoDB for the JWT `sub` claim, the backend auto-creates one. Fields populated from JWT claims: `user_id` (sub), `email`, `display_name` (preferred_username or name).

### Roles

Single role `user` for v1. Role field present in the data model but no RBAC enforcement beyond "is authenticated".

### Environment Variables (Backend)

| Variable | Default | Purpose |
|----------|---------|---------|
| `KEYCLOAK_URL` | `http://localhost:8180` | Keycloak base URL |
| `KEYCLOAK_REALM` | `go-fuels` | Realm name |

## 2. Geocoding — Nominatim

### Interface

```go
type Geocoder interface {
    Search(ctx context.Context, query string) ([]GeoResult, error)
}
```

Default implementation: `NominatimGeocoder`. Can be swapped for ISTAT static dataset or Google Maps Geocoding API in the future without changing consumers.

### Nominatim Call

```
GET https://nominatim.openstreetmap.org/search
    ?q=Milano
    &countrycodes=it
    &featuretype=city
    &format=json
    &limit=5
```

Filtered to centro abitato types only (city/town/village). Italy only (`countrycodes=it`).

### Rate Limiting

Nominatim requires max 1 request/second. Backend enforces this with a `time.Ticker`-based rate limiter. Additionally, geocoded locations are cached in MongoDB — the same city name won't trigger a second Nominatim call.

### User Flow

1. User types city name in frontend search box
2. Frontend calls `GET /api/locations/geocode?q=Milano`
3. Backend calls Nominatim, filters results, returns candidates
4. User selects the correct result
5. Frontend calls `POST /api/users/me/locations` with selected location
6. Backend checks if location exists in MongoDB (by Nominatim `place_id`, used as stable key). If not, creates it. Associates it with user preferences.

## 3. Data Model — MongoDB Extensions

### Collection `users` (new, replaces `user_preferences`)

```json
{
  "user_id": "keycloak-sub-uuid",
  "email": "user@example.com",
  "display_name": "Mario Rossi",
  "preferred_locations": [
    {
      "location_key": "city_nominatim_158946",
      "name": "Milano",
      "added_at": "2026-04-02T10:00:00Z"
    }
  ],
  "preferred_fuels": ["fuel_1_true", "fuel_2_false"],
  "created_at": "2026-04-02T09:00:00Z",
  "updated_at": "2026-04-02T10:00:00Z"
}
```

- `preferred_locations`: max 3 entries, validated server-side
- `preferred_fuels`: optional filter; empty means "show all"
- `user_id`: matches Keycloak JWT `sub` claim

### Collection `locations` (extended)

```json
{
  "location_key": "city_nominatim_158946",
  "name": "Milano",
  "region": "Lombardia",
  "type": "city",
  "lat": 45.4642,
  "lng": 9.1900,
  "radius": 5,
  "source": "nominatim",
  "created_at": "2026-04-02T10:00:00Z"
}
```

- `type: "city"` — geocoded locations from Nominatim
- `type: "station"` — fuel stations from MISE (existing, key format `station_{id}`)
- Geographic link: stations within a city are found by coordinates + radius query on TimescaleDB/MISE data

### Migration

The existing `user_preferences` collection is deprecated. New `users` collection takes over. Old data can be migrated if needed but there are no production users yet.

## 4. API — New Endpoints

All `/api/*` endpoints require a valid JWT (except `/api/auth/config`).

### Auth

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/auth/config` | Returns Keycloak OIDC config (realm URL, client_id). Public. |

### Geocoding

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/locations/geocode?q=` | Search Italian cities via Nominatim. Returns candidate list. |

### User Profile & Preferences

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/users/me` | Current user profile (auto-created on first call) |
| `PUT` | `/api/users/me` | Update display_name |
| `GET` | `/api/users/me/locations` | List preferred locations |
| `POST` | `/api/users/me/locations` | Add preferred location (max 3) |
| `DELETE` | `/api/users/me/locations/:location_key` | Remove preferred location |
| `GET` | `/api/users/me/fuels` | List preferred fuels |
| `PUT` | `/api/users/me/fuels` | Update preferred fuels list |

### Dashboard Data

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/dashboard/locations/:location_key/prices` | Current prices for a location (stations within radius) |
| `GET` | `/api/dashboard/locations/:location_key/trends?fuel_key=&since=` | Historical trend for location + fuel |
| `POST` | `/api/dashboard/locations/:location_key/ingest` | Trigger ingestion job for location's coordinates |

### Existing Endpoints

All existing endpoints (`/fuels/*`, `/location/*`, `/brand/*`, `/user/*`, `/ingestion/*`) remain unchanged and unauthenticated.

## 5. Frontend React

### Stack

- **Vite + React + TypeScript**
- **React Router** for routing
- **Recharts** for charts/trends
- **react-oidc-context** + **oidc-client-ts** for OIDC auth with Keycloak
- **fetch** native for API calls (no Axios)

### Pages

| Page | Route | Description |
|------|-------|-------------|
| Login | `/login` | Redirect to Keycloak |
| Dashboard | `/` | Card per preferred location: current prices + mini sparkline chart |
| Location Detail | `/locations/:key` | Full price table + historical trend charts (Recharts). Fuel filter toggle. |
| Settings | `/settings` | Add/remove locations (search autocomplete), manage preferred fuels |

### Layout

Sidebar navigation: Dashboard, Settings, Logout. Responsive but desktop-first.

### First Visit Flow

1. User arrives at `/` — not authenticated — redirect to Keycloak
2. User registers/logs in on Keycloak
3. Redirect back to `/` — backend auto-creates MongoDB profile
4. Dashboard is empty — prompt: "Aggiungi il tuo primo luogo"
5. User navigates to Settings, searches "Milano", selects, saves
6. Returns to Dashboard — sees fuel price data

### API Module

Centralized `api.ts` module that:
- Reads access token from `react-oidc-context`
- Injects `Authorization: Bearer <token>` header on every request
- Handles 401 by triggering token refresh or re-login
- Base URL configurable via env var (`VITE_API_URL`, default `http://localhost:8080`)

### Directory Structure

```
frontend/
├── index.html
├── package.json
├── vite.config.ts
├── src/
│   ├── main.tsx
│   ├── App.tsx
│   ├── api.ts                  # Centralized API client
│   ├── auth/
│   │   └── AuthProvider.tsx    # OIDC context wrapper
│   ├── pages/
│   │   ├── Dashboard.tsx
│   │   ├── LocationDetail.tsx
│   │   └── Settings.tsx
│   ├── components/
│   │   ├── LocationCard.tsx    # Dashboard card per location
│   │   ├── PriceTable.tsx      # Fuel prices table
│   │   ├── TrendChart.tsx      # Recharts line chart
│   │   ├── CitySearch.tsx      # Autocomplete city search
│   │   └── Layout.tsx          # Sidebar + content layout
│   └── types/
│       └── index.ts            # TypeScript types matching API models
```

## 6. Infrastructure — Docker Compose Addition

Added to `build/docker-compose.yaml`:

```yaml
keycloak:
  image: quay.io/keycloak/keycloak:26.2
  command: start-dev --import-realm
  ports:
    - "8180:8080"
  environment:
    KEYCLOAK_ADMIN: admin
    KEYCLOAK_ADMIN_PASSWORD: admin
  volumes:
    - ./keycloak/realm-export.json:/opt/keycloak/data/import/realm-export.json
```

Realm configuration auto-imported on startup. No manual Keycloak setup required after `docker compose up`.

Frontend runs via `cd frontend && npm run dev` on port `5173` (not containerized in v1).

## 7. Scope

### In Scope (v1)

- Keycloak with pre-configured realm in docker-compose
- JWT validation middleware in Go backend
- Geocoder interface + Nominatim implementation with rate limiting
- User sync Keycloak to MongoDB on first login
- CRUD user preferences (max 3 locations, preferred fuels)
- New `/api/*` authenticated endpoints
- React SPA: login, dashboard, location detail, settings
- Historical trend charts with Recharts
- Trigger ingestion jobs from frontend

### Out of Scope (v1)

- RBAC / admin roles
- Alternative geocoders (ISTAT, Google Maps) — interface ready, not implemented
- Push/email notifications on price changes
- Frontend containerization
- CI/CD pipeline
- Frontend E2E tests
- SSR / SEO
