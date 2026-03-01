# eKapterka

[![Go Version](https://img.shields.io/badge/Go-1.24-00ADD8?logo=go)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](./LICENSE)

eKapterka is a Telegram bot for browsing and managing outdoor gear inventory, with item data stored in Google Firestore and photos stored in Google Cloud Storage.

## Table of Contents

- [Features](#features)
- [Tech Stack](#tech-stack)
- [Project Structure](#project-structure)
- [Requirements](#requirements)
- [Configuration](#configuration)
- [Quick Start](#quick-start)
- [Running with Docker](#running-with-docker)
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

Optional:

- `SERVICE_URL` - public base URL; if set, webhook is registered as `SERVICE_URL + WEBHOOK_PATH`
- `ADMIN_CODE` - code used by `/getadmin <code>`
- `GOOGLE_APPLICATION_CREDENTIALS` - path to GCP service account JSON

Important:

Firestore project ID is currently hardcoded as `e-kapterka` in:

- `ekapterka/internal/repository/client.go`
- `ekapterka/cmd/seed/main.go`

If you use a different GCP project, update those files.

## Quick Start

1. Export environment variables.
2. Run the bot.

```bash
cd ekapterka
export PORT=8080
export BOT_TOKEN=<telegram_token>
export WEBHOOK_PATH=/webhook
export STORAGE_ID=<gcs_bucket_name>
export SERVICE_URL=<https_public_url_optional>
export ADMIN_CODE=<admin_code_optional>
export GOOGLE_APPLICATION_CREDENTIALS=<path_to_service_account_json>

go run ./cmd/bot
```

Seed categories:

```bash
cd ekapterka
export GOOGLE_APPLICATION_CREDENTIALS=<path_to_service_account_json>
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
  -e SERVICE_URL=<https_public_url_optional> \
  -e ADMIN_CODE=<admin_code_optional> \
  -e GOOGLE_APPLICATION_CREDENTIALS=/secrets/gcp.json \
  -v /local/path/gcp.json:/secrets/gcp.json:ro \
  ekapterka-bot
```

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
- Firestore project ID is hardcoded in two places (see [Configuration](#configuration)).
- `tags` exist in the model, but there are no tag management commands yet.

## License

This project is licensed under the MIT License. See [LICENSE](./LICENSE) for details.
