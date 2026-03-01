# Operations Runbook

This runbook focuses on operating the bot in Cloud Run with Firestore + GCS.

## 1. Health Signals

Service-level signals:

1. Cloud Run revision status is `Ready`.
2. No sustained error bursts in service logs.
3. Telegram users can run `/start` and browse categories.
4. Admin commands affecting Firestore/GCS complete successfully.

Application-level expected logs:

- Startup: `Listening on :<PORT>`
- Webhook setup:
  - success: `Webhook set successfully`
  - skip: `SERVICE_URL not set, skipping setWebhook`

## 2. Basic Verification Commands

```bash
# Service URL
gcloud run services describe <service-name> \
  --region <region> \
  --format='value(status.url)'

# Recent logs
gcloud run services logs read <service-name> \
  --region <region> \
  --limit 200
```

## 3. Common Incidents and Fixes

## 3.1 Bot does not receive updates

Symptoms:

- No reactions to Telegram commands.

Checks:

1. Confirm `SERVICE_URL` env is set.
2. Confirm `WEBHOOK_PATH` env is correct.
3. Confirm Cloud Run service is publicly accessible (`--allow-unauthenticated` or equivalent IAM policy).
4. Check logs for `setWebhook failed`.

Fixes:

- Update env vars and redeploy/update service.
- Ensure correct Telegram token.

## 3.2 Category list is empty

Symptoms:

- Browse flow shows no categories/items even though bot runs.

Checks:

1. Verify `categories` collection exists.
2. Confirm seed tool was executed for target project.

Fixes:

- Run `go run ./cmd/seed` with correct project credentials.

## 3.3 Photo upload fails on `/add` or `/edit`

Symptoms:

- Command returns upload error.

Checks:

1. Runtime service account has `storage.objectAdmin` on bucket.
2. `STORAGE_ID` points to existing bucket.
3. Outbound internet access is available (Telegram file download + GCS API).

Fixes:

- Correct IAM and env vars.
- Confirm bucket region/policy compatibility.

## 3.4 Firestore errors

Symptoms:

- Command failures on data read/write.

Checks:

1. Runtime service account has Firestore permissions (`roles/datastore.user` or stronger).
2. Firestore API enabled.
3. `FIRESTORE_PROJECT_ID` matches deployment project.

Fixes:

- Grant IAM role.
- Update `FIRESTORE_PROJECT_ID` and redeploy.

## 3.5 Admin role cannot be granted

Symptoms:

- `/getadmin` always fails.

Checks:

1. `ADMIN_CODE` exists and matches input exactly.
2. Secret binding/injection to runtime env is correct.

Fixes:

- Rotate and re-inject secret value.

## 4. Maintenance Tasks

## 4.1 Rotate secrets

1. Add new Secret Manager version.
2. Update Cloud Run to reference latest secret version (or keep `latest`).
3. Deploy new revision.

## 4.2 Cleanup orphaned photos

Why it happens:

- If DB update succeeds but old photo deletion fails, stale objects can remain.

Approach:

1. Export current `photo_urls` references from Firestore.
2. Compare against bucket object paths.
3. Remove unreferenced objects with lifecycle policy or batch cleanup script.

## 4.3 Safe release procedure

1. Deploy to new revision.
2. Run smoke checks (`/start`, category browse, `/cat`, `/add` with photo, `/rm`).
3. Monitor logs.
4. Roll back traffic if critical errors appear.

## 5. Observability Gaps

Current gaps in code:

- No structured logs.
- No application metrics.
- No tracing.
- No dead-letter queue for dropped updates.

Recommended improvements:

1. Add request/update correlation IDs in logs.
2. Emit counters for update enqueue drops.
3. Add explicit log events for command success/failure.
4. Integrate with Cloud Monitoring dashboards/alerts.
