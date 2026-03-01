# Bot Flows

This document describes user and admin interaction flows exactly as implemented.

## 1. Update Routing

Entry point: `Bot.handleUpdate`

Routing rules:

1. `CallbackQuery != nil` -> callback handler.
2. `Message == nil` -> ignore update.
3. If message is a command (or command in caption entities) -> command handler.
4. Plain text messages are currently ignored (`handleTextMessage` is not active).

## 2. Main Menu and Navigation

Main menu buttons:

- `🔍 Найти снаряжение` -> `menu:find`
- `👤 Мой профиль` -> `menu:profile`

Browse navigation callbacks:

- `menu:find`, `search:root` -> show root child categories.
- `search:category:<id>` -> open category or show leaf items.
- `search:items:<category_id>:<page>` -> paginated items.
- `search:item:<item_id>:p:<page>` -> item card.
- Back buttons route to parent category/root/main menu depending on context.

Pagination:

- Page size is fixed to 10 items.
- Prev/next arrows rendered based on page index and `hasNext`.

## 3. Commands

Supported commands:

- `/start`
- `/add`
- `/edit`
- `/cat`
- `/rm`
- `/rent`
- `/unr`
- `/cmd`
- `/getadmin`

## 3.1 `/start`

- Ensures user state in Firestore (`users` document).
- Sends main menu.

## 3.2 `/cmd`

- Always includes `/start`.
- Includes admin commands only if user role is `admin`.

## 3.3 `/getadmin <code>`

- Compares input code to env var `ADMIN_CODE`.
- On success sets user role to `admin`.
- On missing/invalid code returns error message.

## 3.4 `/cat`

- Admin-only.
- Lists all leaf categories as `<code>ID</code> title` in HTML parse mode.

## 3.5 `/add`

Admin-only format:

```text
/add
<category_id>
<title>
<description line 1>
<description line 2>
...
```

Behavior:

- Accepts command from message text or photo caption.
- Description is multiline (`join(lines[3:], "\n")`).
- If photo exists, uploads largest Telegram photo to GCS and stores public URL.
- Persists item in Firestore.

Validation:

- Requires at least 4 lines.
- `category_id` and `title` must be non-empty.

## 3.6 `/edit <id>`

Admin-only format:

```text
/edit <item_id>
<category_id>
<new title>
<description line 1>
<description line 2>
...
```

Behavior:

- Parses item ID from first command line.
- Description is multiline.
- If new photo attached, uploads new photo and replaces `photo_urls`.
- If photo replaced, attempts to delete old GCS objects.

Validation:

- Requires at least 4 lines.
- Requires non-empty `item_id`, `category_id`, `title`.
- Item must exist.

## 3.7 `/rm <id>`

Admin-only.

Behavior:

- Reads item ID from command arguments.
- Fallback parser reads first line fields from text/caption when args are missing.
- Deletes item document and then deletes linked photo objects.

Note:

- Error hint says `/delete <id>`, but registered command is `/rm`.

## 3.8 `/rent <id>`

Admin-only format:

```text
/rent <item_id>
01.01.2025
10.02.2025
<optional multiline admin note>
```

Behavior:

- Parses dates using `02.01.2006` format.
- Ensures `end >= start`.
- Appends rental using Firestore `ArrayUnion`.

## 3.9 `/unr <id> <number>`

Admin-only.

Behavior:

- Fetches item rentals.
- Sorts rentals by start date, then end date.
- Removes rental by 1-based index in sorted view.
- Writes full updated rentals array back to Firestore.

## 4. Item Card Rendering

Card composition:

1. Item ID (always)
2. Category ID (admin only)
3. Title
4. Description (if non-empty)
5. Availability:
   - no rentals -> `✅ Свободно`
   - rentals -> sorted list with date ranges
6. Rental descriptions shown only to admins.
7. Contact footer: "Для того чтобы арендовать, пишите каптерщику <contact>"

Contact source:

- Derived from first word of bot description via Telegram `getMyDescription`.
- Value cached with `sync.Once`.
- If unavailable, fallback value is currently empty string.

Rendering mode:

- HTML parse mode.
- If item has photo URL, sends/edits photo message with caption.
- Caption is truncated to 1024 runes for Telegram limits.

## 5. Role Enforcement

Admin-only commands call `requireAdmin`:

- Resolves role from Firestore (`GetUserRole`).
- On role lookup failure returns access-check error.
- Non-admin users receive explicit forbidden response.

## 6. Error Handling Pattern

General pattern:

- Errors are handled inline and returned to user as short message.
- Internal errors are logged with contextual metadata.
- In callback/message editing, if edit fails the bot falls back to sending a new message.

## 7. Localization

Current user-facing text is mostly Russian.

Implication:

- If introducing multi-language support, all command help and callback responses should be centralized first.
