# Development Guide

## 1. Repository Layout

Top-level:

- `README.md`: project overview and quick usage.
- `docs/`: detailed technical documentation.
- `ekapterka/`: Go module with bot implementation.

Go module key areas:

- `cmd/bot`: runtime entrypoint.
- `cmd/seed`: category seed utility.
- `internal/bot`: update handling and Telegram UI rendering.
- `internal/repository`: Firestore persistence.
- `internal/storage`: GCS integration.
- `internal/models`: domain structs.

## 2. Local Prerequisites

- Go 1.24+
- Access to Firestore and GCS in target GCP project
- Telegram bot token

Env vars required for local bot run:

- `PORT`
- `BOT_TOKEN`
- `WEBHOOK_PATH`
- `STORAGE_ID`
- `FIRESTORE_PROJECT_ID`

Optional:

- `SERVICE_URL`
- `ADMIN_CODE`
- `GOOGLE_APPLICATION_CREDENTIALS`

## 3. Local Run

```bash
cd ekapterka
export PORT=8080
export BOT_TOKEN=<telegram-token>
export WEBHOOK_PATH=/webhook
export STORAGE_ID=<bucket-name>
export FIRESTORE_PROJECT_ID=<your-gcp-project-id>
export GOOGLE_APPLICATION_CREDENTIALS=<path-to-sa-json>
go run ./cmd/bot
```

Seed categories:

```bash
cd ekapterka
export FIRESTORE_PROJECT_ID=<your-gcp-project-id>
go run ./cmd/seed
```

## 4. Test/Build Commands

Run tests:

```bash
cd ekapterka
GOCACHE=/tmp/go-build-cache go test ./...
```

Build binary:

```bash
cd ekapterka
go build -o bot ./cmd/bot
```

Build container:

```bash
cd ekapterka
docker build -t ekapterka-bot .
```

## 5. Coding Constraints in Current Design

1. `FIRESTORE_PROJECT_ID` must be provided and point to a project with Firestore enabled.
2. Bot logic is strongly coupled to concrete repository/storage clients.
3. Queue is process-local and not suitable for exactly-once semantics.
4. UI text is mostly Russian and spread across handlers.

## 6. High-Risk Change Areas

When modifying, test these carefully:

1. Callback payload formats (`search:*`, `menu:*`) because navigation depends on exact string patterns.
2. Command parsers for multiline input and caption-command support.
3. Photo replacement flow (`/edit`) because it combines DB update and object deletion.
4. Rental create/remove flows (`/rent`, `/unr`) because they touch a separate `rentals` collection and item timestamps.

## 7. Suggested Development Workflow

1. Run `go test ./...` before and after changes.
2. For bot-flow changes, run manual smoke tests in Telegram:
   - `/start`
   - browse categories
   - view item card with/without photo
   - admin flows (`/getadmin`, `/grantadmin`, `/revokeadmin`, `/add`, `/edit`, `/rent`, `/unr`, `/rm`)
3. Keep docs updated when command syntax or env vars change.

## 8. Contribution Conventions (Recommended)

Not enforced by tooling yet, but recommended:

1. Keep user-facing behavior changes backward-compatible when possible.
2. Add migration notes if changing Firestore schema.
3. Avoid introducing hardcoded environment-specific values.
4. Update both `README.md` and `docs/` for operationally relevant changes.
