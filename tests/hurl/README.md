# Hurl Integration Tests

This directory contains API-level integration tests runnable against a deployed Voyage backend.

## Setup

```bash
cp tests/hurl/.env.example tests/hurl/.env
# then edit tests/hurl/.env
```

Required vars:
- `BASE_URL` (e.g. `https://voyage.example.com`)
- `API_KEY`

Optional vars:
- `SEARCH_QUERY` (default `*`)
- `RUN_ID` (default: current epoch seconds)

## Run

```bash
./scripts/run-hurl.sh
```

## Files

- `01-health.hurl` — health endpoint smoke test
- `02-email-tag-flow.hurl` — search/get/tag smoke flow
- `03-trip-flow.hurl` — create/associate/list/delete trip flow
- `04-negative-cases.hurl` — basic error-path assertions
