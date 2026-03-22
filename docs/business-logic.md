# Business Logic

This document describes the domain behavior and business rules implemented in eKapterka.

## 1. Domain Purpose

eKapterka is an inventory and rental-assistance bot for outdoor gear.

Core business goals:

1. Let users discover available equipment by category.
2. Let admins maintain inventory records.
3. Track item rental periods for planning and conflict visibility.
4. Keep user access control simple (`user` vs `admin`).

## 2. Domain Entities and Roles

## 2.1 Roles

- `user`: can browse catalog and view item cards.
- `admin`: has full inventory/rental management commands.

Role assignment model:

- New users default to `user`.
- Role elevation is done via `/getadmin <code>` or by another admin using `/grantadmin <user_id> <code>` with runtime secret `ADMIN_CODE`.
- Role revocation is done by admins using `/revokeadmin <user_id> <code>` with the same secret.

## 2.2 Categories

Categories define navigable structure for inventory browsing:

- Hierarchical parent/child tree.
- `is_leaf` indicates categories intended to contain items.
- Ordered navigation is defined by category `order`.

Business intent:

- Browsing should feel directory-like and predictable.
- Admin-facing `/cat` command exposes valid leaf IDs for item placement.

## 2.3 Items

Each item represents a rentable or trackable inventory unit with:

- identity (document ID)
- title and description
- category placement
- optional photos
- rental windows

Business assumption:

- Item belongs to one category (`category_id`).
- Category existence/leaf validation is not strictly enforced in command logic; correctness is expected from admin workflow.

## 2.4 Rentals

A rental is a period with:

- start date
- end date
- optional admin description
- renter Telegram user ID

Business intent:

- Keep a visible booking history/future plan directly in item card.
- Allow admin note per rental (e.g., group name, trip comment).

## 3. End-User Business Flow

## 3.1 Entry

- User starts with `/start`.
- Bot initializes user state and displays main menu.

## 3.2 Browsing

1. User enters "Find equipment" flow.
2. User drills down category tree.
3. On leaf category, user sees paginated item list.
4. On item selection, user sees card with status and rentals.

Business output of item card:

- inventory identity
- readability of current availability
- human contact path for renting (quartermaster contact line)

## 3.3 Profile

- User can open profile to see items they are renting.
- Profile list is built from rentals where `user_id` matches the Telegram user.

## 4. Admin Business Flow

## 4.1 Become Admin

1. User sends `/getadmin <code>`.
2. Bot compares with `ADMIN_CODE`.
3. On success, role becomes `admin`.

Admins can also grant or revoke admin role for other users with the same secret code.

Business rationale:

- Lightweight and fast role enablement without additional admin panel.

## 4.2 Add Item (`/add`)

Format:

```text
/add
<category_id>
<title>
<description...>
```

Rules:

1. `category_id` and `title` are mandatory.
2. Description is optional and supports multiple lines.
3. If photo attached, largest Telegram photo is uploaded and linked.

Business effect:

- New inventory unit appears in category browsing.

## 4.3 Edit Item (`/edit <id>`)

Format:

```text
/edit <item_id>
<category_id>
<new title>
<new description...>
```

Rules:

1. `item_id`, `category_id`, and `title` are required.
2. Description is optional and supports multiple lines (empty clears the description).
3. If new photo attached:
   - item is updated with new photo URL
   - old photo objects are deleted after successful update

Business effect:

- Inventory metadata stays current while minimizing orphaned media.

## 4.4 Delete Item (`/rm <id>`)

Rules:

1. Item is removed from Firestore.
2. Linked photos are removed from GCS.
3. If photo deletion partially fails, item deletion is not rolled back; warning is returned.

Business tradeoff:

- Prefers inventory consistency over strict media cleanup atomicity.

## 4.5 Add Rental (`/rent <id>`)

Format:

```text
/rent <item_id>
DD.MM.YYYY
DD.MM.YYYY
<renter telegram_id>
<optional multiline admin note>
```

Rules:

1. Dates must parse as `02.01.2006`.
2. End date cannot be earlier than start date.
3. Item must exist.
4. Renter `telegram_id` must be a positive number.
5. Rental is stored as a separate document in `rentals`.

Business effect:

- Item card reflects booking window and reduces accidental double-assignment.

## 4.6 Remove Rental (`/unr <rental_id>`)

Rules:

1. Rental is resolved by Firestore document ID.
2. Selected rental document is deleted.
3. Related item `updated_at` is refreshed.

Business rationale:

- Rental records become addressable and removable without list re-numbering.

## 5. Availability Logic

Displayed availability:

- `✅ Свободно` when no rentals exist.
- `Арендовано:` with numbered periods when rentals exist.

Important implementation detail:

- Availability is informational only.
- The system currently does not block overlapping rentals.

Business implication:

- Conflict prevention is operational/manual, not strict transactional validation.

## 6. Contact Logic in Item Card

Footer line asks users to contact quartermaster.

Resolution strategy:

1. Read bot description via Telegram `getMyDescription`.
2. Take first word as contact token.
3. Cache the result for runtime.

Business intent:

- Avoid hardcoded contact and allow quick operational updates through bot profile description.

## 7. Access Control Rules

Admin-only commands:

- `/add`, `/edit`, `/rm`, `/cat`, `/rent`, `/unr`, `/grantadmin`, `/revokeadmin`

General rules:

1. Role is resolved on request.
2. If role check fails technically, command is denied.
3. Non-admin receives explicit denial message.

Business stance:

- Fail closed for privileged operations.

## 8. Data Integrity Rules (Current)

Applied rules:

1. User state auto-bootstrap on first interaction.
2. Rental date order validation (`end >= start`).
3. Mandatory field checks for core item mutation commands.

Not yet enforced at command level:

1. Category existence verification on add/edit.
2. Category leaf-only assignment enforcement.
3. Rental overlap prevention.
4. Transactional coupling between item delete/edit and media delete.

## 9. Operational Business Constraints

1. Update queue is in-memory and may drop updates when full.
2. Single worker default limits throughput but keeps flow deterministic.
3. `FIRESTORE_PROJECT_ID` must match deployment project.
4. Most user-facing text is currently Russian.

## 10. Practical Business Invariants

In normal operation, the system expects:

1. Category tree is seeded and stable.
2. Admins use `/cat` to choose valid category IDs.
3. Item cards are the primary source of rental status truth for users.
4. Admins treat rental records as planning metadata, not legal booking contracts.

## 11. Recommended Next Business Enhancements

1. Enforce category existence and `is_leaf` on add/edit.
2. Add overlap checks and conflict warnings for rental periods.
3. Add optional soft-delete/archive for items.
4. Add audit trail for role changes and inventory mutations.
5. Add explicit quartermaster profile settings instead of parsing description text.
