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

## Docker Compose (full annotated example)

```yaml
services:
  voyage-api:
    build:
      context: .
      dockerfile: Dockerfile.api
    ports:
      - "8181:8181"
    volumes:
      # Mount your existing maildir and notmuch config (read-only for safety)
      - /home/you/.mail:/mail:ro
      - /home/you/.notmuch-config:/config/notmuch/config:ro
      # Voyage data files (trips + classifiers) — writable
      - ./trips.json:/app/trips.json
      - ./configs/classifiers.yaml:/app/configs/classifiers.yaml
    environment:
      VOYAGE_PORT: "8181"
      VOYAGE_MAIL_DIR: /mail
      VOYAGE_NOTMUCH_CONFIG: /config/notmuch/config
      VOYAGE_TRIPS_DB: /app/trips.json
      VOYAGE_CLASSIFIERS: /app/configs/classifiers.yaml
      VOYAGE_API_KEY: "${VOYAGE_API_KEY}"
      VOYAGE_ALLOWED_ORIGINS: "${VOYAGE_ALLOWED_ORIGINS:-http://localhost:1234}"
    restart: unless-stopped
```

> **Tip for home servers / Tailscale:** Set `VOYAGE_ALLOWED_ORIGINS` to your device's local IP or Tailscale address, e.g. `http://100.x.x.x:1234`.

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

## Configuration

| Variable | Description | Default |
|---|---|---|
| `VOYAGE_API_KEY` | API key — **leave empty to disable auth** (trusted local network) | *(none)* |
| `VOYAGE_PORT` | Port to listen on | `8181` |
| `VOYAGE_ALLOWED_ORIGINS` | Comma-separated CORS origins | `http://localhost:8080,...` |
| `VOYAGE_MAIL_DIR` | Path to your Maildir | *(required)* |
| `VOYAGE_NOTMUCH_CONFIG` | Path to notmuch config | *(required)* |
| `VOYAGE_TRIPS_DB` | Trips persistence file | `./trips.json` |
| `VOYAGE_CLASSIFIERS` | Email category rules YAML | `./configs/classifiers.yaml` |
| `VOYAGE_SYNC_CMD` | Mail sync command | `./scripts/sync-mail.sh` |

### Customizing email categories

Edit `configs/classifiers.yaml` to add your own airlines, hotel chains, or regional carriers. No restart required — or just restart the container. See the file for the full format.

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

## Architecture

```
mbsync → local Maildir → notmuch index → voyage-api → Voyage app
```

No cloud. No OAuth. No third-party APIs. Your email never leaves your machine.
