# Park Smarter consumer API plan

Hand-maintained reverse-engineering notes for **parksmarter-pp-cli**.

> **Not the partner DMS API.** IPS Group’s Data Management System (DMS) and city enforcement integrations are B2G/B2B-only. This CLI targets the **consumer mobile app** backend used by `com.ipsgroupinc.parksmarter` (Park Smarter™ US).

## Discovery summary (Sep 2026)

| Source | Finding | Status |
|--------|---------|--------|
| Play Store / App Store | Package `com.ipsgroupinc.parksmarter`, IPS Group Inc. | Verified |
| APK v4.4.0 (via apkeep) | React Native / Expo Hermes bundle | Verified |
| Hermes bundle strings | Production host `https://apiv3.parksmarter.com` | **Verified** |
| Hermes bundle strings | Staging/dev hosts `staging-parksmarter-api.ipsmeters.com`, `dev-parksmarter-api.ipsmeters.com` (Azure) | Verified (not used by CLI default) |
| DNS | `api.parksmarter.com` resolves (Cloudflare) but returns ASP.NET errors on `/api/*`; likely legacy | Unverified / avoid |
| `www.parksmarter.com` | Marketing site only; Cloudflare blocked automated fetch | N/A |
| Live probe | `GET https://apiv3.parksmarter.com/api/ApplicationValidity` → 200 JSON without auth | **Verified** |
| Live probe | Auth-gated paths return HTTP 401 without token | **Verified** |

## Base URL

**Live (CLI default):** `https://apiv3.parksmarter.com`

Do **not** use `parksmarter.com` or undocumented DMS hosts for consumer account operations.

## Auth — partially verified

The Android app (Expo) stores tokens in secure storage and calls `clientConfig.getAuthToken` before API requests.

| Endpoint | Method | Notes |
|----------|--------|-------|
| `/api/Auth/LoginWithPhoneAndPassword` | POST | Exists (200 with empty body on bad creds). APK Redux: `loginWithPhoneAndPasswordAsyncThunk`. **Request JSON field names unverified** — CLI sends PascalCase `PhoneNumber`, `Password`. |
| `/api/Auth/LoginWithCachedAuthToken` | POST | APK action verified. Body likely `{ "AuthToken": "…" }` — **unverified**. |
| `/api/Auth/LoginWithApple` | POST | Exists; out of CLI v0 scope. |
| `/api/Auth/RefreshToken` | POST | Exists; not wired in v0. |
| `/api/Auth/LogOutAllDevices` | POST | Exists; not wired in v0. |

**Transport:** CLI sends `Authorization: Bearer <token>` and duplicate `AuthToken: <token>` header (ASP.NET mobile pattern — **header name unverified** without live token).

**Token file:** `~/.config/parksmarter-pp-cli/token.json` (0600). Env override: `PARKSMARTER_TOKEN`.

### Capturing a token (if password login fails)

1. Install Park Smarter on a device you own; sign in.
2. Proxy HTTPS from the device (Charles, mitmproxy, HTTP Toolkit) with user CA installed.
3. Filter host `apiv3.parksmarter.com`.
4. Capture `POST /api/Auth/LoginWithPhoneAndPassword` response JSON **or** any authed request’s bearer header.
5. Save `{ "auth_token": "…" }` and run `parksmarter-pp-cli auth login --token-file token.json`.

Document observed request/response shapes in a PR — do not commit secrets.

## Public / low-auth endpoints

| Endpoint | Method | Verified response |
|----------|--------|-------------------|
| `/api/ApplicationValidity` | GET | `{ ForceUpgrade, RecommendUpgrade, Config, SessionId, … }` |
| `/api/State/List` | GET | `{ StatesResult: [] }` (empty without params — shape unverified) |

## Account — auth required, response shapes unverified

| Endpoint | Method | APK / probe |
|----------|--------|-------------|
| `/api/Card/GetCards` | POST | Redux: `addCardAsyncThunk`, `deleteCardAsyncThunk` |
| `/api/Vehicle/GetVehicles` | POST | Redux: `addVehicleAsyncThunk`, `updateVehicleAsyncThunk` |
| `/api/User/GetUserDetail` | GET/POST | Returns `null` without token; shape **unverified** |

CLI redacts output to **ids + last4** (cards) and plate metadata (vehicles).

## Meter / zone lookup — auth required

All are **GET** with query parameters (auth-gated reads return 401 without token; POST without body returns 411 — not used by CLI).

| Endpoint | Use |
|----------|-----|
| `/api/MeterList/GetMetersNearby` | `--lat` / `--lng` |
| `/api/MeterList/GetMeterByNumber` | `--meter-number` |
| `/api/MeterList/GetMetersByZoneNameSearch` | `--zone` |
| `/api/MeterList/GetMetersByZoneOrSpaceNameSearch` | `--zone` + `--space` |
| `/api/MeterList/GetMetersByLocationAddress` | `--address` |
| `/api/MeterList/GetMetersByLatLng` | alternate geo search (not CLI default) |
| `/api/MeterList/GetLimitedMetersByLocation` | app feature flag `UseLimitedMetersFetch` |
| `/api/MeterList/GetMetersByScannerCode` | barcode scan flow |

**Query param names** (`Latitude`, `Longitude`, `MeterNumber`, …) are inferred from ASP.NET conventions — **unverified** without authed HAR.

## Sessions — auth required

| Endpoint | Method | Use |
|----------|--------|-----|
| `/api/ParkingSession/GetActiveParkingSessions` | GET | `sessions list` |
| `/api/ParkingSession/GetPastParkingSessions` | GET | `sessions list --past` |

Session id field in JSON likely `SessionId` — mapped defensively in client.

## Parking mutations — auth required, GET semantics (path verified; query params unverified)

Live probes (Sep 2026, no credentials):

| Endpoint | GET (no auth) | POST (no body) |
|----------|---------------|----------------|
| `/api/ParkingSession/StartParkingSession` | **401** | **411** Length Required |
| `/api/ParkingSession/ExtendSession` | **401** | **411** |
| `/api/ParkingSession/StopSession` | **401** | **411** |

Mutations are **GET** with query parameters (not POST JSON). ASP.NET returns 401 without a bearer token on GET; POST without `Content-Length` returns 411 — do **not** treat POST 411 as proof of POST semantics.

| Endpoint | CLI command | Confirm phrase |
|----------|-------------|----------------|
| `/api/ParkingSession/StartParkingSession` | `parking start` | `START PARK SMARTER PARKING` |
| `/api/ParkingSession/ExtendSession` | `parking extend` | `EXTEND PARK SMARTER PARKING` |
| `/api/ParkingSession/StopSession` | `parking stop` | `STOP PARK SMARTER PARKING` |

**Query params (unverified):** `MeterNumber`, `ZoneName`, `SpaceNumber`, `Minutes`, `VehicleId`, `CardId`, `SessionId`.

Pricing preview endpoints (exist, 401 without auth):

- `/api/ParkingEstimate/GetParkingEstimateItems`
- `/api/ParkingEstimate/GetParkingEstimateMulti`

Not wired in v0 CLI; use app UI or capture HAR to confirm bodies.

## Safety gates (v0)

- **Start:** `--enable-live-parking --owner-approved --confirm "START PARK SMARTER PARKING"`
- **Extend:** `--enable-live-parking --owner-approved --confirm "EXTEND PARK SMARTER PARKING"`
- **Stop:** `--enable-live-parking --owner-approved --confirm "STOP PARK SMARTER PARKING"`
- **Live HTTP (non-`--dry-run`):** additionally requires **`--acknowledge-unverified-body`** until HAR-verified query params ship.
- **Phone/password login:** blocked unless **`--acknowledge-unverified-body`**; prefer `auth login --token-file` from your own HAR.
- **`--dry-run`:** builds query map; does not send mutating GET; must not emit `started`/`extended`/`stopped: true`.

## Known gaps / TODO

1. Confirm login JSON field names and success response token keys with authed HAR.
2. Confirm `Authorization` vs `AuthToken` header with live token.
3. Map full `StartParkingSession` required params (payment tokenization, zone policy).
4. Wire `ParkingEstimate` preview before start when shapes are confirmed.
5. Apple SSO login (`LoginWithApple`) — out of v0 scope.
6. BLE meter provisioning — out of v0 scope (app feature, not REST).

## What this is not

- IPS DMS, enforcement, or permit APIs
- Park Smarter UK (`parksmarter.uk`) — separate deployment
- Unrelated GitHub “ParkSmart” / “ParkO” student projects
- MCP integration (v0 non-goal)
