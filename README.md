# Snippr

A self-hosted snippet manager for developers. Save, search, and share code snippets — with a web UI, REST API, and CLI tool.

**Live demo:** [snippr-vp2w.onrender.com](https://snippr-vp2w.onrender.com)

## Features

- **Auth** — JWT-based register/login with httpOnly session cookie
- **Snippet CRUD** — create, read, update, delete with tags and language metadata
- **Full-text search** — Postgres `tsvector` powered search across title and content
- **Public sharing** — generate a shareable slug link for any snippet (`/s/:slug`)
- **Workspaces** — shared spaces with admin/member roles; public or invite-only
- **Live board** — real-time snippet feed per workspace via Server-Sent Events
- **Web UI** — React + shadcn/ui with dark/light theme and syntax highlighting
- **CLI tool** — install once, run as `snippr` from anywhere
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

# Run Go server
make run

# Run React dev server (with API proxy)
cd web && npm run dev
```

## CLI

Install the CLI once and run it as `snippr` from anywhere:

```bash
# Install to $GOPATH/bin (add $GOPATH/bin to your PATH if not already)
make install

# Or just build locally
make build-cli   # outputs bin/snippr
```

Add to your shell profile if needed:
```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

```bash
# Authenticate
snippr login

# Save a snippet (reads from stdin)
cat main.go | snippr save "Server entrypoint" --lang go --tag backend

# List snippets
snippr list

# Search
snippr search "postgres"

# Get and copy to clipboard
snippr get 1 --copy
```

## Workspaces

Workspaces let you share snippets with a team.

- **Create** a workspace from the web UI — choose public or private
- **Private** workspaces require an invite code to join
- **Public** workspaces are discoverable and joinable by anyone
- **Admin** (creator) can edit the workspace, remove members, and delete it
- **Live board** tab shows snippet activity in real time — no refresh needed

## Environment variables

| Variable | Default | Description |
|---|---|---|
| `PORT` | `8080` | HTTP port |
| `DATABASE_URL` | `postgres://snippr:snippr@localhost:5432/snippr?sslmode=disable` | Postgres DSN |
| `JWT_SECRET` | `dev-secret-change-in-production` | JWT signing key |
| `APP_ENV` | `development` | Set to `production` to enable Secure cookie flag |

Copy `.env.example` to `.env` and fill in your values.

## Deploy (Render)

1. Push to GitHub
2. Create a new **Web Service** on [render.com](https://render.com) — point it at this repo, runtime **Docker**
3. Create a **PostgreSQL** database on Render and copy the internal URL
4. Set env vars: `DATABASE_URL`, `JWT_SECRET`, `APP_ENV=production`
5. Deploy
