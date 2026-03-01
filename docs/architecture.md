# Architecture

## 1. System Overview

eKapterka is a Telegram webhook bot for equipment catalog browsing and inventory rental tracking.

High-level runtime responsibilities:

1. Receive Telegram updates over HTTP webhook.
2. Enqueue updates into an in-memory queue.
3. Process updates via worker goroutines.
4. Read/write domain data in Firestore.
5. Upload/delete item photos in Google Cloud Storage (GCS).
6. Render Telegram messages and inline keyboards for user navigation.

## 2. Runtime Topology

Entrypoint: `ekapterka/cmd/bot/main.go`

Startup sequence:

1. Load required env vars: `PORT`, `BOT_TOKEN`, `WEBHOOK_PATH`, `STORAGE_ID`.
2. Initialize shared context (`context.Background()`).
3. Initialize GCS client (`internal/storage`).
4. Initialize Firestore repository client (`internal/repository`).
5. Create bot (`internal/bot/NewBot`).
6. Start update workers (`StartWorkers(1, 100)`).
7. Register HTTP webhook handler at `WEBHOOK_PATH`.
8. Start HTTP server and webhook registration goroutines.

Concurrency model:

- HTTP handler is non-blocking and only enqueues updates.
- Processing happens in a buffered in-memory channel (`updateQueue`).
- Current configuration runs one worker (`numWorkers=1`), which simplifies ordering and avoids concurrent write races in bot logic.

## 3. Component Breakdown

## 3.1 API Boundary

- Telegram -> HTTP webhook endpoint in `internal/bot.WebhookHandler`.
- Outbound Telegram API calls through `go-telegram-bot-api` client.

Webhook registration:

- `SetupWebhook` executes once on startup after a 5-second delay.
- Requires `SERVICE_URL`; if missing, registration is skipped.

## 3.2 Bot Layer (`internal/bot`)

Primary files:

- `bot.go`: bot construction, webhook setup, update routing entry.
- `worker.go`: queue and worker processing.
- `comand-handler.go`: slash command handling and admin actions.
- `callback-query-handler.go`: inline callback navigation and item-card rendering.
- `render.go`: keyboard builders and Telegram message edit/send behavior.

Responsibilities:

- Route update type (`message`, `callback_query`).
- Enforce role-based access on admin commands.
- Parse command payloads (including multiline descriptions).
- Render stateful browsing flow over categories/items.
- Marshal domain objects into Telegram-friendly text and keyboards.

## 3.3 Repository Layer (`internal/repository`)

Provides Firestore access abstractions:

- `categories.go`: category tree reads and leaf listing.
- `items.go`: CRUD and rental update operations for items.
- `users.go`: role persistence and user-state bootstrap.
- `client.go`: Firestore client initialization.

Notable design choices:

- Firestore project ID is provided via required env var `FIRESTORE_PROJECT_ID`.
- Item pagination uses `offset` + `limit+1` for has-next detection.
- Item query includes fallback without ordering if indexed order query fails.
- Backward compatibility for legacy field `rental_periods` when decoding item documents.

## 3.4 Storage Layer (`internal/storage`)

- Implements GCS object upload, delete, and public URL generation.
- Photo object path format at upload time:
  - `items/<chat_id>/<timestamp>_<telegram_file_unique_id>.<ext>`
- `cleanPath` normalizes object paths before storage operations.

## 3.5 Seed Tool (`cmd/seed`, `internal/seed`)

- Writes predefined category tree to Firestore `categories` collection.
- Uses deterministic IDs and full path metadata for each category.
- Seed tool also reads Firestore project ID from `FIRESTORE_PROJECT_ID`.

## 4. Request/Update Processing Flow

## 4.1 Webhook Update Ingestion

1. Telegram POSTs update to `WEBHOOK_PATH`.
2. `HandleUpdate` parses request into `tgbotapi.Update`.
3. Update is enqueued (`EnqueueUpdate`).
4. If queue is full, update is dropped and logged.

## 4.2 Worker Dispatch

1. Worker reads update from queue.
2. `handleUpdate` routes by type:
   - `CallbackQuery` -> callback flow.
   - `Message` command -> command flow.
   - Non-command text currently ignored.

## 4.3 Callback Browse Flow

User flow:

1. Main menu (`menu:find`) -> root categories.
2. Category selection:
   - non-leaf -> load child categories.
   - leaf -> load paginated item list.
3. Item selection -> item card message/photo with navigation controls.

## 4.4 Command Flow

User command:

- `/start`: ensures user state and shows main menu.

Admin commands:

- `/getadmin <code>`: compare with `ADMIN_CODE`, set role.
- `/add`: create item with optional photo and multiline description.
- `/edit <id>`: update category/title/description, optionally replace photo.
- `/rm`: delete item and associated GCS photos.
- `/cat`: list leaf categories.
- `/rent <id>`: append rental period.
- `/unr <id> <n>`: remove rental by sorted ordinal.
- `/cmd`: show command list based on role.

## 5. Rendering Strategy

Message rendering rules (`render.go`):

- Prefer message edit when callback interaction has an existing message ID.
- Fallback to sending a new message if edit fails.
- For photo cards, attempt media edit first, fallback to new photo message.
- Caption length is truncated to Telegram 1024-rune limit.
- HTML parse mode is used for item card formatting.

## 6. Security and Access Control

Role model:

- Roles stored per user in `users` collection (`user`, `admin`).
- Admin access enforced in bot layer via `requireAdmin`.
- Admin elevation uses shared secret (`ADMIN_CODE`) at runtime.

Storage/security notes:

- Bot token and admin code are sensitive and should be stored as secrets.
- Bucket/object permissions are externalized to GCP IAM.
- Public image URLs are generated in `https://storage.googleapis.com/<bucket>/<object>` format.

## 7. Reliability Characteristics

Current behavior:

- In-memory queue only: no persistence, no retry for dropped updates.
- Single worker by default.
- No graceful shutdown orchestration (process blocks on `select {}`).
- Log-based observability only (no metrics/tracing).

Operational implications:

- Bot restart may lose in-flight queue updates.
- Burst traffic beyond queue capacity can drop updates.
- Horizontal scaling with multiple replicas can process webhook updates independently; Telegram webhook delivery strategy should be considered during scaling policy definition.

## 8. Known Architectural Constraints

1. No repository interfaces for dependency inversion/testing.
2. No migration framework; schema evolves in-code.
3. Message language and UX text are currently mostly Russian.
4. Update queue is process-local and not distributed.

## 9. Extension Points

Recommended extension directions:

1. Replace in-memory queue with managed queue (Pub/Sub / Cloud Tasks).
2. Add structured logging + metrics.
3. Add unit and integration tests around command/callback handlers.
4. Introduce explicit domain service layer if business logic grows.
