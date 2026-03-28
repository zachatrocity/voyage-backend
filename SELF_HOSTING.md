# Self-Hosting Voyage

Voyage is a self-hosted travel email aggregator. It reads your email via [notmuch](https://notmuchmail.org/) and surfaces travel-related messages through a REST API.

## Prerequisites

- **notmuch** installed and configured with your mail
- **mbsync** (isstrunc) or another IMAP sync tool already pulling mail into a Maildir
- **Go 1.23+** (if running directly) or **Docker** (if using containers)

## Environment Variables

| Variable | Description | Default |
|---|---|---|
| `VOYAGE_PORT` | Port the API server listens on | `8181` |
| `VOYAGE_ALLOWED_ORIGINS` | Comma-separated CORS origins | `http://localhost:8080,http://localhost:1234,http://localhost:8181` |
| `VOYAGE_MAIL_DIR` | Path to your Maildir | *(none)* |
| `VOYAGE_NOTMUCH_CONFIG` | Path to your notmuch config file | *(none)* |
| `VOYAGE_TRIPS_DB` | Path to trips JSON database | `./trips.json` |
| `VOYAGE_CLASSIFIERS` | Path to classifiers YAML config | `./configs/classifiers.yaml` |
| `VOYAGE_SYNC_CMD` | Command to run for mail sync | `./scripts/sync-mail.sh` |
| `VOYAGE_API_KEY` | API key for authentication (empty = auth disabled) | *(none)* |

## Example `.env` file

```env
VOYAGE_PORT=8181
VOYAGE_ALLOWED_ORIGINS=https://voyage.example.com,http://localhost:1234
VOYAGE_MAIL_DIR=/home/user/.mail
VOYAGE_NOTMUCH_CONFIG=/home/user/.notmuch-config
VOYAGE_TRIPS_DB=./trips.json
VOYAGE_CLASSIFIERS=./configs/classifiers.yaml
VOYAGE_SYNC_CMD=./scripts/sync-mail.sh
VOYAGE_API_KEY=your-secret-key-here
```

## Docker Compose

1. Copy the example `.env` file above and adjust the values.
2. Run:

```bash
docker compose up -d
```

The API will be available at `http://localhost:8181`.

To view logs:

```bash
docker compose logs -f voyage-api
```

## Running Directly

```bash
# Set environment variables (or source your .env)
export VOYAGE_MAIL_DIR=/home/user/.mail
export VOYAGE_NOTMUCH_CONFIG=/home/user/.notmuch-config

# Build and run
go run ./cmd/api
```

The server starts on port 8181 by default.

## Tailscale / Home Network Tips

If you're running Voyage on a home server and want to access it from other devices:

- **Tailscale**: Install Tailscale on your server and client devices. Access Voyage at `http://<tailscale-ip>:8181`. Set `VOYAGE_ALLOWED_ORIGINS` to include your Tailscale hostname or IP.
- **Reverse proxy**: If using Caddy or nginx, proxy to `localhost:8181` and set `VOYAGE_ALLOWED_ORIGINS` to your public domain.
- **Firewall**: If exposing directly, ensure only port 8181 is open and consider setting `VOYAGE_API_KEY` for authentication.

## Classifiers

The file at `configs/classifiers.yaml` defines email category rules (flights, hotels, car rentals, etc.). Edit it to add your own providers or keywords. See the default file for the format.
