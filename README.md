# eKapterka

[![Go Version](https://img.shields.io/badge/Go-1.24-00ADD8?logo=go)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](./LICENSE)

Language: English | [Русский](./README.ru.md)

eKapterka is a Telegram bot for browsing and managing outdoor gear inventory, with item data stored in Google Firestore and photos stored in Google Cloud Storage.

## Table of Contents

- [Features](#features)
- [Tech Stack](#tech-stack)
- [Detailed Documentation](#detailed-documentation)
- [Project Structure](#project-structure)
- [Requirements](#requirements)
- [Configuration](#configuration)
- [Quick Start](#quick-start)
- [Running with Docker](#running-with-docker)
- [Deploy to Google Cloud Run (from scratch)](#deploy-to-google-cloud-run-from-scratch)
- [Bot Commands](#bot-commands)
- [Data Model (Firestore)](#data-model-firestore)
- [Testing](#testing)
- [Known Limitations](#known-limitations)
- [License](#license)

## Features

- Interactive category navigation using Telegram inline keyboards.
- Item cards with title, description, photos, and rental periods.
- Admin role support (`user` / `admin`) with role upgrade via secret code.
- Admin CRUD flows for items.
- Rental period management (add/remove rental windows).
- Multi-line descriptions in `/add` and `/edit` command flows.

## Tech Stack

- Go `1.24`
- Telegram Bot API (`github.com/go-telegram-bot-api/telegram-bot-api/v5`)
- Google Firestore
- Google Cloud Storage
- Docker (multi-stage build, distroless runtime)

## Detailed Documentation

For full technical documentation, see the [`docs/`](./docs) directory:

- [`docs/architecture.md`](./docs/architecture.md)
- [`docs/business-logic.md`](./docs/business-logic.md)
- [`docs/data-model.md`](./docs/data-model.md)
- [`docs/deployment.md`](./docs/deployment.md)
- [`docs/bot-flows.md`](./docs/bot-flows.md)
- [`docs/operations.md`](./docs/operations.md)
- [`docs/development.md`](./docs/development.md)

## Project Structure

- `ekapterka/cmd/bot` - main bot entrypoint
- `ekapterka/cmd/seed` - category seeding entrypoint
- `ekapterka/internal/bot` - command handlers, callback handlers, rendering
- `ekapterka/internal/repository` - Firestore data access layer
- `ekapterka/internal/storage` - GCS upload/delete logic
- `ekapterka/internal/models` - domain models
- `ekapterka/internal/seed` - static category seed data
- `ekapterka/internal/server` - HTTP server for Telegram webhook

## Requirements

- Go `1.24+`
- A Telegram bot token from `@BotFather`
- A Google Cloud project with:
  - Firestore enabled
  - A Cloud Storage bucket for item photos
  - Service account credentials with required permissions

## Configuration

Set these environment variables for `cmd/bot`:

Required:

- `PORT` - HTTP server port (for example `8080`)
- `BOT_TOKEN` - Telegram bot token
- `WEBHOOK_PATH` - webhook route (for example `/webhook`)
- `STORAGE_ID` - GCS bucket name for photos
- `FIRESTORE_PROJECT_ID` - GCP project ID used to initialize Firestore client

Optional:

- `SERVICE_URL` - public base URL; if set, webhook is registered as `SERVICE_URL + WEBHOOK_PATH`
- `ADMIN_CODE` - code used by `/getadmin <code>`
- `GOOGLE_APPLICATION_CREDENTIALS` - path to GCP service account JSON

Important:

Set `FIRESTORE_PROJECT_ID` to the exact Google Cloud project ID where your Firestore database is located.

## Quick Start

1. Export environment variables.
2. Run the bot.

```bash
cd ekapterka
export PORT=8080
export BOT_TOKEN=<telegram_token>
export WEBHOOK_PATH=/webhook
export STORAGE_ID=<gcs_bucket_name>
export FIRESTORE_PROJECT_ID=<your-gcp-project-id>
export SERVICE_URL=<https_public_url_optional>
export ADMIN_CODE=<admin_code_optional>
export GOOGLE_APPLICATION_CREDENTIALS=<path_to_service_account_json>

go run ./cmd/bot
```

Seed categories:

```bash
cd ekapterka
export GOOGLE_APPLICATION_CREDENTIALS=<path_to_service_account_json>
export FIRESTORE_PROJECT_ID=<your-gcp-project-id>
go run ./cmd/seed
```

## Running with Docker

From the `ekapterka` directory:

```bash
docker build -t ekapterka-bot .
docker run --rm -p 8080:8080 \
  -e PORT=8080 \
  -e BOT_TOKEN=<telegram_token> \
  -e WEBHOOK_PATH=/webhook \
  -e STORAGE_ID=<gcs_bucket_name> \
  -e FIRESTORE_PROJECT_ID=<your-gcp-project-id> \
  -e SERVICE_URL=<https_public_url_optional> \
  -e ADMIN_CODE=<admin_code_optional> \
  -e GOOGLE_APPLICATION_CREDENTIALS=/secrets/gcp.json \
  -v /local/path/gcp.json:/secrets/gcp.json:ro \
  ekapterka-bot
```

## Deploy to Google Cloud Run (from scratch)

This section uses anonymized placeholders. Replace values in `<...>` with your own.

### 1. Prepare local variables

```bash
export PROJECT_ID="<your-gcp-project-id>"
export REGION="<your-region>"               # e.g. europe-west1
export REPO_DIR="<absolute-path-to-repo>"   # e.g. /home/user/eKapterka
export SERVICE_NAME="tg-bot"
export WEBHOOK_PATH="/webhook"
export BUCKET_NAME="<unique-bucket-name>"   # must be globally unique
export ADMIN_CODE="<strong-admin-code>"
export FIRESTORE_PROJECT_ID="$PROJECT_ID"
```

### 2. Authenticate and select project

```bash
gcloud auth login
gcloud config set project "$PROJECT_ID"
```

If needed, link billing to the project before continuing.

### 3. Enable required APIs

```bash
gcloud services enable \
  run.googleapis.com \
  cloudbuild.googleapis.com \
  artifactregistry.googleapis.com \
  firestore.googleapis.com \
  storage.googleapis.com \
  secretmanager.googleapis.com
```

### 4. Create Firestore database (if not created yet)

```bash
gcloud firestore databases create \
  --location="$REGION" \
  --type=firestore-native
```

If Firestore already exists, this command will fail and you can skip this step.

### 5. Create a bucket for item photos

```bash
gcloud storage buckets create "gs://$BUCKET_NAME" \
  --location="$REGION" \
  --uniform-bucket-level-access
```

### 6. Store sensitive values in Secret Manager

```bash
printf "%s" "<telegram-bot-token>" | \
  gcloud secrets create bot-token --data-file=- 2>/dev/null || \
  printf "%s" "<telegram-bot-token>" | gcloud secrets versions add bot-token --data-file=-

printf "%s" "$ADMIN_CODE" | \
  gcloud secrets create admin-code --data-file=- 2>/dev/null || \
  printf "%s" "$ADMIN_CODE" | gcloud secrets versions add admin-code --data-file=-
```

### 7. Grant runtime access to secrets and storage

Use the Cloud Run service account (by default this is usually `<project-number>-compute@developer.gserviceaccount.com` unless you configured a custom one):

```bash
export PROJECT_NUMBER="$(gcloud projects describe "$PROJECT_ID" --format='value(projectNumber)')"
export RUNTIME_SA="${PROJECT_NUMBER}-compute@developer.gserviceaccount.com"

gcloud projects add-iam-policy-binding "$PROJECT_ID" \
  --member="serviceAccount:$RUNTIME_SA" \
  --role="roles/secretmanager.secretAccessor"

gcloud storage buckets add-iam-policy-binding "gs://$BUCKET_NAME" \
  --member="serviceAccount:$RUNTIME_SA" \
  --role="roles/storage.objectAdmin"
```

### 8. Deploy the service (first pass)

Deploy from source in `ekapterka` directory:

```bash
cd "$REPO_DIR/ekapterka"

gcloud run deploy "$SERVICE_NAME" \
  --source . \
  --region "$REGION" \
  --allow-unauthenticated \
  --set-env-vars "WEBHOOK_PATH=$WEBHOOK_PATH,STORAGE_ID=$BUCKET_NAME,FIRESTORE_PROJECT_ID=$FIRESTORE_PROJECT_ID" \
  --set-secrets "BOT_TOKEN=bot-token:latest,ADMIN_CODE=admin-code:latest"
```

### 9. Get Cloud Run URL and set `SERVICE_URL`

```bash
export SERVICE_URL="$(gcloud run services describe "$SERVICE_NAME" \
  --region "$REGION" \
  --format='value(status.url)')"
```

Update service env so bot can self-register Telegram webhook on startup:

```bash
gcloud run services update "$SERVICE_NAME" \
  --region "$REGION" \
  --update-env-vars "SERVICE_URL=$SERVICE_URL"
```

### 10. Seed categories once

Because Firestore category tree is required for navigation, run the seed command once with project credentials:

```bash
cd "$REPO_DIR/ekapterka"
go run ./cmd/seed
```

### 11. Verify deployment

```bash
gcloud run services describe "$SERVICE_NAME" \
  --region "$REGION" \
  --format='value(status.url)'
```

Open the bot in Telegram and run `/start`.

Notes:

- `FIRESTORE_PROJECT_ID` must point to the GCP project where Firestore is enabled.
- `SERVICE_URL` is required for automatic webhook registration inside the bot startup flow.
- For production, prefer a dedicated service account instead of the default compute service account.

## Bot Commands

User commands:

- `/start` - initialize user state and open main menu
- `/cmd` - show available commands

Admin commands:

- `/getadmin <code>` - grant yourself `admin` role if code matches `ADMIN_CODE`
- `/cat` - list leaf categories (`ID + title`)
- `/add` - create item
- `/edit <id>` - edit item
- `/rm <id>` - delete item
- `/rent <id>` - add rental period
- `/unr <id> <number>` - remove rental period by index

`/add` format (supports multi-line description):

```text
/add
<category_id>
<title>
<description line 1>
<description line 2>
...
```

`/edit` format (supports multi-line description):

```text
/edit <item_id>
<category_id>
<new title>
<description line 1>
<description line 2>
...
```

`/rent` format:

```text
/rent <item_id>
01.01.2025
10.02.2025
<optional admin note>
```

## Data Model (Firestore)

Collection `categories`:

- `id`, `title`, `parent_id`, `path`, `level`, `order`, `is_leaf`

Collection `items`:

- `title`, `description`, `category_id`, `tags`, `photo_urls`, `created_at`, `updated_at`, `rentals[]`

`rentals[]` entry:

- `start`, `end`, `description`

Collection `users`:

- `id`, `role`, `created_at`, `message_id`

## Testing

```bash
cd ekapterka
GOCACHE=/tmp/go-build-cache go test ./...
```

## Known Limitations

- Most bot UI text is currently in Russian.
- `tags` exist in the model, but there are no tag management commands yet.

## License

This project is licensed under the MIT License. See [LICENSE](./LICENSE) for details.
