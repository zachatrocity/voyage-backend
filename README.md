<div align="center">

# ✉️ Voyage Backend ✈️

**Self-hosted travel email aggregator — your mail, your server, no cloud.**

</div>

## What it does

Voyage indexes your local email via [notmuch](https://notmuchmail.org/) and exposes a REST API that the Voyage app uses to surface travel confirmations, organize them into trips, and build itineraries — all without touching a third-party service.

## Requirements

- **notmuch** configured and indexing your mail
- **mbsync** (or any IMAP sync tool) pulling mail into a local Maildir
- **Docker** (recommended) or Go 1.23+

## Quick Start with Docker Compose

1. **Copy the example env file:**

```bash
cp .env.example .env
```

2. **Edit `.env`** with your paths and a strong API key:

```env
VOYAGE_API_KEY=change-me-to-a-long-random-secret

VOYAGE_MAIL_DIR=/home/you/.mail
VOYAGE_NOTMUCH_CONFIG=/home/you/.notmuch-config

# Optional — adjust if needed
VOYAGE_PORT=8181
VOYAGE_ALLOWED_ORIGINS=http://localhost:1234
```

3. **Start the stack:**

```bash
docker compose up -d
```

4. **Verify:**

```bash
curl -H "X-API-Key: change-me-to-a-long-random-secret" http://localhost:8181/api/v1/search?q=confirmation
```

The `/health` endpoint is always public (no key required):
```bash
curl http://localhost:8181/health
```

---

## Docker Compose

```yaml
services:
  # Mail sync sidecar — pulls Gmail via mbsync, indexes with notmuch
  voyage-mail:
    image: ghcr.io/zachatrocity/voyage-mail:latest
    volumes:
      - ./mail:/mail
      - ./config:/config
    environment:
      EMAIL_PASSWORD: "${EMAIL_PASSWORD}"
    restart: unless-stopped

  # API server
  voyage-api:
    image: ghcr.io/zachatrocity/voyage-backend:latest
    ports:
      - "8181:8181"
    volumes:
      - ./mail:/mail:ro
      - ./config/notmuch/config:/config/notmuch/config:ro
    environment:
      VOYAGE_API_KEY: "${VOYAGE_API_KEY}"
    depends_on:
      - voyage-mail
    restart: unless-stopped
```

### Required `.env` values

| Variable | Description |
|---|---|
| `EMAIL_PASSWORD` | Gmail app password for mbsync |
| `VOYAGE_API_KEY` | API key — generate with `openssl rand -hex 32`. Leave empty to disable auth (trusted local network only). |

### Optional env vars

| Variable | Description | Default |
|---|---|---|
| `SYNC_FREQUENCY` | How often to sync mail (`15m`, `1h`, etc.) | `15m` |
| `VOYAGE_PORT` | Port to listen on | `8181` |
| `VOYAGE_ALLOWED_ORIGINS` | Comma-separated CORS origins (set to your app's IP/Tailscale address) | *(none)* |
| `VOYAGE_MAIL_DIR` | Path to Maildir inside container | `/mail` |
| `VOYAGE_NOTMUCH_CONFIG` | Path to notmuch config inside container | `/config/notmuch/config` |
| `VOYAGE_TRIPS_DB` | Trips persistence file | `./trips.json` |
| `VOYAGE_CLASSIFIERS` | Email category rules YAML — falls back to built-in defaults if not mounted | `./configs/classifiers.yaml` |
| `VOYAGE_SYNC_CMD` | Mail sync command | `./scripts/sync-mail.sh` |

> **Local dev (build from source):** Comment out `image:` and use a `build:` block — `Dockerfile.api` for the API, `Dockerfile` for the mail sidecar.

---

## API

All endpoints under `/api/v1` require the `X-API-Key` header when `VOYAGE_API_KEY` is set.

```
# Auth header
X-API-Key: your-secret-key

# Search emails (notmuch query)
GET  /api/v1/search?q=confirmation&limit=50

# Get single email
GET  /api/v1/email/:id

# Tag email
POST /api/v1/email/:id/tags/:tag

# Trips (coming soon — see issues)
GET  /api/v1/trips
POST /api/v1/trips
POST /api/v1/email/:id/trip/:tripId
GET  /api/v1/trips/:id/emails

# Travel discovery
GET  /api/v1/emails/travel

# Mail sync
POST /api/v1/sync
GET  /api/v1/sync/status

# Always public
GET  /health
GET  /docs
```

Full interactive docs at `http://localhost:8181/docs`.

---

## Running without Docker

```bash
# Set env vars
export VOYAGE_MAIL_DIR=~/.mail
export VOYAGE_NOTMUCH_CONFIG=~/.notmuch-config
export VOYAGE_API_KEY=your-secret

go run ./cmd/api
```

---

## Hurl integration tests (against deployed backend)

These tests are designed to run against a real deployed backend using your own API key.

1. Copy env template:

```bash
cp tests/hurl/.env.example tests/hurl/.env
```

2. Edit `tests/hurl/.env` with your real values (`BASE_URL`, `API_KEY`).

3. Run tests:

```bash
./scripts/run-hurl.sh
```

What it covers right now:
- health check
- search + get email
- basic tag flow (re-applies an existing tag; low-mutation)
- negative cases (400/404)
- optional mutating trip create/associate/list/delete flow (`RUN_MUTATION=1`)

> `tests/hurl/.env` is gitignored so you can safely keep machine-specific/test secrets locally.
> By default, mutating trip tests are skipped to avoid creating test noise on your live backend.

---

## Architecture

```
mbsync → local Maildir → notmuch index → voyage-api → Voyage app
```

No cloud. No OAuth. No third-party APIs. Your email never leaves your machine.

---

<div align="center">🤙 built with vibes 🤙</div>
