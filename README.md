# Snippr

A self-hosted snippet manager for developers. Save, search, and share code snippets — with a web UI, REST API, and CLI tool.

## Features

- **Auth** — JWT-based register/login
- **Snippet CRUD** — create, read, update, delete with tags and language metadata
- **Full-text search** — Postgres `tsvector` powered search across title and content
- **Public sharing** — generate a shareable slug link for any snippet (`/s/:slug`)
- **Web UI** — React + shadcn/ui with syntax highlighting
- **CLI tool** — `snippr save`, `snippr list`, `snippr search`, `snippr get --copy`
- **Docker** — single `docker compose up` to run everything

## Stack

| Layer | Tech |
|---|---|
| Backend | Go · Chi router · JWT |
| Database | PostgreSQL (tsvector FTS) |
| Frontend | React · Vite · shadcn/ui · Tailwind CSS |
| CLI | Go · Cobra |
| Deploy | Docker · Render |

## Quick start

```bash
# Start Postgres + server
docker compose up

# Server runs at http://localhost:8080
```

## Development

```bash
# Start Postgres
docker compose up -d postgres

# Run Go server (hot reload)
make run

# Run React dev server (with API proxy)
cd web && npm run dev
```

## CLI

```bash
# Build
make build-cli

# Login
./bin/snippr login

# Save a snippet
cat main.go | ./bin/snippr save "Server entrypoint" --lang go --tag backend

# List snippets
./bin/snippr list

# Search
./bin/snippr search "postgres"

# Get and copy to clipboard
./bin/snippr get 1 --copy
```

## Environment variables

| Variable | Default | Description |
|---|---|---|
| `PORT` | `8080` | HTTP port |
| `DATABASE_URL` | `postgres://snippr:snippr@localhost:5432/snippr?sslmode=disable` | Postgres DSN |
| `JWT_SECRET` | `dev-secret-change-in-production` | JWT signing key |

## Deploy (Render)

1. Push to GitHub
2. Create a new **Web Service** on [render.com](https://render.com) — point it at this repo, runtime **Docker**
3. Create a **PostgreSQL** database on Render and copy the internal URL
4. Set env vars: `DATABASE_URL`, `JWT_SECRET`
5. Deploy
