# parksmarter-pp-cli

Agent-native [Printing Press](https://github.com/mvanhorn/cli-printing-press) CLI for **Park Smarter™ (IPS Group) consumer accounts** (`com.ipsgroupinc.parksmarter`). Zone/meter lookup, sessions, and **hard-gated** start/extend/stop via the mobile consumer API — **not** IPS partner DMS / enforcement APIs.

**Author:** [Amandeep Khurana](https://github.com/amansk) (@amansk) · **License:** Apache-2.0

> Unrelated projects: this is **Park Smarter by IPS Group** (US meter pay-by-phone), not open-source “ParkSmart” repos or Park Smarter UK (`parksmarter.uk`).

## Install

```bash
go install github.com/amansk/parksmarter-pp-cli/cmd/parksmarter-pp-cli@latest
```

Or build from source:

```bash
git clone https://github.com/amansk/parksmarter-pp-cli
cd parksmarter-pp-cli
go build -o parksmarter-pp-cli ./cmd/parksmarter-pp-cli
```

## Quick start

1. Create a Park Smarter account in the mobile app (iOS/Android).
2. Login and store a bearer token:

```bash
# Phone/password via env (recommended — never echoed)
export PARKSMARTER_PHONE='+15555550100'
export PARKSMARTER_PASSWORD='your-password'
parksmarter-pp-cli auth login

# Or import a token captured from your own session (see PLAN.md)
parksmarter-pp-cli auth login --token-file ./token.json
```

3. Verify setup:

```bash
parksmarter-pp-cli doctor --json
parksmarter-pp-cli auth status
parksmarter-pp-cli account vehicles --json
parksmarter-pp-cli account cards --json
```

Saved cards output is **id + last4 only** — never full PAN or token secrets.

4. Zone / meter lookup:

```bash
parksmarter-pp-cli zones lookup --meter-number 12345 --json
parksmarter-pp-cli zones lookup --zone "Main St" --space 12 --json
parksmarter-pp-cli zones nearby --lat 32.7157 --lng -117.1611 --json
```

5. Sessions:

```bash
parksmarter-pp-cli sessions list --json
parksmarter-pp-cli sessions list --past --json
parksmarter-pp-cli sessions get <session-id> --json
```

## Live parking (hard-gated)

Preview never charges:

```bash
parksmarter-pp-cli parking preview --action start \
  --meter-number 12345 --minutes 60 --json

parksmarter-pp-cli parking preview --action extend \
  --session-id <id> --minutes 30 --json
```

Live start/extend/stop require **all three** gates (exact confirm string per action):

```bash
parksmarter-pp-cli parking start --meter-number 12345 --minutes 60 \
  --enable-live-parking --owner-approved \
  --confirm "START PARK SMARTER PARKING" --json
```

Inspect the request without mutating:

```bash
parksmarter-pp-cli parking start ... --dry-run --json
```

| Action | Confirm phrase |
|--------|----------------|
| `parking start` | `START PARK SMARTER PARKING` |
| `parking extend` | `EXTEND PARK SMARTER PARKING` |
| `parking stop` | `STOP PARK SMARTER PARKING` |

## Global flags

| Flag | Description |
|------|-------------|
| `--json` | Machine-readable JSON output |
| `--agent` | `--json --no-color --no-input` (does **not** imply `--yes`) |
| `--dry-run` | Skip mutating GETs where supported |
| `--home` | Override config dir (`$PARKSMARTER_PP_HOME` or `~/.config/parksmarter-pp-cli`) |

## Exit codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 2 | Usage / validation |
| 3 | Not found |
| 4 | Auth |
| 5 | API error |
| 7 | Rate limit / transient |

## Auth details

- Tokens saved to `~/.config/parksmarter-pp-cli/token.json` with mode **0600**.
- Override with `PARKSMARTER_TOKEN` (raw bearer).
- Login credentials via `PARKSMARTER_PHONE` / `PARKSMARTER_PASSWORD` or `--phone` / `--password`.
- **`auth status` and `doctor` never print secret values.**

## Development

```bash
go test ./...
go vet ./...
```

See [PLAN.md](./PLAN.md) for consumer API endpoint notes (verified vs unverified) and HAR capture instructions.

## Publishing

Intended for eventual `/printing-press-publish` into [mvanhorn/printing-press-library](https://github.com/mvanhorn/printing-press-library) under `library/commerce/parksmarter/`.
