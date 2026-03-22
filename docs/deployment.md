# Deployment

This guide documents deployment options for the current codebase, with a production-oriented Google Cloud Run path.

## 1. Deployment Targets

Supported by repository artifacts:

1. Local process (`go run ./cmd/bot`)
2. Docker container (`ekapterka/Dockerfile`)
3. Google Cloud Run (source-based deploy with `gcloud run deploy --source .`)

## 2. Runtime Environment Variables

Required:

- `PORT`: HTTP listen port.
- `BOT_TOKEN`: Telegram bot token.
- `WEBHOOK_PATH`: webhook endpoint path (e.g. `/webhook`).
- `STORAGE_ID`: GCS bucket name for item photos.
- `FIRESTORE_PROJECT_ID`: GCP project ID where Firestore is enabled.

Optional:

- `SERVICE_URL`: public base URL used for Telegram `setWebhook` call.
- `ADMIN_CODE`: secret used for `/getadmin <code>`, `/grantadmin <user_id> <code>`, `/revokeadmin <user_id> <code>`.

Notes:

- Missing required vars causes startup panic (`config.MustEnv`).
- If `SERVICE_URL` is empty, webhook registration is skipped.

## 3. Prerequisites (GCP)

1. Existing GCP project.
2. Billing enabled.
3. APIs enabled:
   - Cloud Run
   - Cloud Build
   - Artifact Registry
   - Firestore
   - Cloud Storage
   - Secret Manager
4. Telegram bot created in `@BotFather`.
5. Firestore native database created.
6. GCS bucket created for photos.

## 4. Firestore Project Selection

Firestore client initialization is controlled by `FIRESTORE_PROJECT_ID`.

Deployment requirement:

- `FIRESTORE_PROJECT_ID` must match the project where your Firestore database exists.

## 5. Recommended Secret Strategy

Store sensitive values in Secret Manager:

- `BOT_TOKEN`
- `ADMIN_CODE`

Inject them into Cloud Run as secrets using `--set-secrets`.

## 6. Cloud Run Deployment (From Scratch)

Use placeholders and replace `<...>`.

### 6.1 Set local shell variables

```bash
export PROJECT_ID="<your-gcp-project-id>"
export REGION="<your-region>"              # e.g. europe-west1
export SERVICE_NAME="tg-bot"
export WEBHOOK_PATH="/webhook"
export BUCKET_NAME="<unique-bucket-name>"
export REPO_DIR="<absolute-path-to-repo>"
export ADMIN_CODE_VALUE="<strong-admin-code>"
export FIRESTORE_PROJECT_ID="$PROJECT_ID"
```

### 6.2 Authenticate and select project

```bash
gcloud auth login
gcloud config set project "$PROJECT_ID"
```

### 6.3 Enable APIs

```bash
gcloud services enable \
  run.googleapis.com \
  cloudbuild.googleapis.com \
  artifactregistry.googleapis.com \
  firestore.googleapis.com \
  storage.googleapis.com \
  secretmanager.googleapis.com
```

### 6.4 Create Firestore (if not already present)

```bash
gcloud firestore databases create \
  --location="$REGION" \
  --type=firestore-native
```

### 6.5 Create storage bucket

```bash
gcloud storage buckets create "gs://$BUCKET_NAME" \
  --location="$REGION" \
  --uniform-bucket-level-access
```

### 6.6 Create/update secrets

```bash
printf "%s" "<telegram-bot-token>" | \
  gcloud secrets create bot-token --data-file=- 2>/dev/null || \
  printf "%s" "<telegram-bot-token>" | gcloud secrets versions add bot-token --data-file=-

printf "%s" "$ADMIN_CODE_VALUE" | \
  gcloud secrets create admin-code --data-file=- 2>/dev/null || \
  printf "%s" "$ADMIN_CODE_VALUE" | gcloud secrets versions add admin-code --data-file=-
```

### 6.7 Grant runtime IAM permissions

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

For production, prefer a dedicated service account over the default compute account.

### 6.8 Deploy first revision (without SERVICE_URL)

```bash
cd "$REPO_DIR/ekapterka"

gcloud run deploy "$SERVICE_NAME" \
  --source . \
  --region "$REGION" \
  --allow-unauthenticated \
  --set-env-vars "WEBHOOK_PATH=$WEBHOOK_PATH,STORAGE_ID=$BUCKET_NAME,FIRESTORE_PROJECT_ID=$FIRESTORE_PROJECT_ID" \
  --set-secrets "BOT_TOKEN=bot-token:latest,ADMIN_CODE=admin-code:latest"
```

### 6.9 Read service URL and update runtime env

```bash
export SERVICE_URL="$(gcloud run services describe "$SERVICE_NAME" \
  --region "$REGION" \
  --format='value(status.url)')"

gcloud run services update "$SERVICE_NAME" \
  --region "$REGION" \
  --update-env-vars "SERVICE_URL=$SERVICE_URL"
```

After this, app startup can call Telegram `setWebhook` automatically.

### 6.10 Seed categories once

```bash
cd "$REPO_DIR/ekapterka"
go run ./cmd/seed
```

### 6.11 Verify

```bash
gcloud run services describe "$SERVICE_NAME" \
  --region "$REGION" \
  --format='value(status.url)'
```

Then open the bot in Telegram and send `/start`.

## 7. Docker Deployment Notes

Container details (`ekapterka/Dockerfile`):

- Build stage: `golang:1.24-alpine`, binary compiled with `CGO_ENABLED=0`.
- Runtime stage: `gcr.io/distroless/base-debian12`.
- Default `PORT=8080`.

Local run:

```bash
cd ekapterka
docker build -t ekapterka-bot .
docker run --rm -p 8080:8080 \
  -e PORT=8080 \
  -e BOT_TOKEN=<telegram-token> \
  -e WEBHOOK_PATH=/webhook \
  -e STORAGE_ID=<bucket-name> \
  -e FIRESTORE_PROJECT_ID=<your-gcp-project-id> \
  -e SERVICE_URL=<public-url-optional> \
  -e ADMIN_CODE=<admin-code-optional> \
  ekapterka-bot
```

## 8. Post-Deployment Checklist

1. `SERVICE_URL + WEBHOOK_PATH` responds with HTTP 200 to Telegram webhook posts.
2. `/start` opens main menu.
3. Category browsing works (seed applied).
4. Admin acquisition via `/getadmin <code>` works.
5. Admin can grant/revoke other admins via `/grantadmin` and `/revokeadmin`.
6. Item create/edit/delete works including photo upload/delete.
6. Cloud Run logs show no repeated `setWebhook`/storage/firestore errors.

## 9. Rollback and Safe Changes

Cloud Run rollback:

1. List revisions: `gcloud run revisions list --service <service> --region <region>`
2. Shift traffic to known-good revision via Cloud Run UI or `gcloud run services update-traffic`.

Safe rollout tips:

- Keep `WEBHOOK_PATH` stable across revisions.
- Avoid changing Firestore schema and command behavior in same release.
- Seed categories in controlled step after schema/category changes.
