# eKapterka

[![Go Version](https://img.shields.io/badge/Go-1.24-00ADD8?logo=go)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](./LICENSE)

Language: [English](./README.md) | Русский

eKapterka — это Telegram-бот для просмотра и управления инвентарем туристического снаряжения, где данные о предметах хранятся в Google Firestore, а фотографии — в Google Cloud Storage.

## Содержание

- [Возможности](#возможности)
- [Технологический стек](#технологический-стек)
- [Подробная документация](#подробная-документация)
- [Структура проекта](#структура-проекта)
- [Требования](#требования)
- [Конфигурация](#конфигурация)
- [Быстрый старт](#быстрый-старт)
- [Запуск через Docker](#запуск-через-docker)
- [Деплой в Google Cloud Run (с нуля)](#деплой-в-google-cloud-run-с-нуля)
- [Команды бота](#команды-бота)
- [Модель данных (Firestore)](#модель-данных-firestore)
- [Тестирование](#тестирование)
- [Известные ограничения](#известные-ограничения)
- [Лицензия](#лицензия)

## Возможности

- Интерактивная навигация по категориям через inline-кнопки Telegram.
- Карточки предметов с названием, описанием, фото и периодами аренды.
- Поддержка ролей (`user` / `admin`) с повышением роли через секретный код.
- Админские CRUD-потоки для предметов.
- Управление периодами аренды (добавление/удаление).
- Многострочные описания в командах `/add` и `/edit`.
- Профиль пользователя показывает текущее арендованное снаряжение.

## Технологический стек

- Go `1.24`
- Telegram Bot API (`github.com/go-telegram-bot-api/telegram-bot-api/v5`)
- Google Firestore
- Google Cloud Storage
- Docker (multi-stage сборка, distroless runtime)

## Подробная документация

Полная техническая документация находится в каталоге [`docs/`](./docs):

- [`docs/architecture.md`](./docs/architecture.md)
- [`docs/business-logic.md`](./docs/business-logic.md)
- [`docs/data-model.md`](./docs/data-model.md)
- [`docs/deployment.md`](./docs/deployment.md)
- [`docs/bot-flows.md`](./docs/bot-flows.md)
- [`docs/operations.md`](./docs/operations.md)
- [`docs/development.md`](./docs/development.md)

## Структура проекта

- `ekapterka/cmd/bot` - основной entrypoint бота
- `ekapterka/cmd/seed` - entrypoint для сидирования категорий
- `ekapterka/internal/bot` - обработчики команд, callback'ов и рендеринг
- `ekapterka/internal/repository` - слой доступа к данным Firestore
- `ekapterka/internal/storage` - логика загрузки/удаления в GCS
- `ekapterka/internal/models` - доменные модели
- `ekapterka/internal/seed` - статические данные для сидирования категорий
- `ekapterka/internal/server` - HTTP-сервер для Telegram webhook

## Требования

- Go `1.24+`
- Токен Telegram-бота от `@BotFather`
- Проект в Google Cloud, в котором:
  - включен Firestore
  - создан Cloud Storage bucket для фото
  - сервисный аккаунт имеет необходимые права

## Конфигурация

Для `cmd/bot` требуются следующие переменные окружения:

Обязательные:

- `PORT` - HTTP-порт сервера (например `8080`)
- `BOT_TOKEN` - токен Telegram-бота
- `WEBHOOK_PATH` - путь webhook (например `/webhook`)
- `STORAGE_ID` - имя GCS bucket для фото
- `FIRESTORE_PROJECT_ID` - ID GCP-проекта, используемый для инициализации Firestore клиента

Опциональные:

- `SERVICE_URL` - публичный base URL; если задан, webhook регистрируется как `SERVICE_URL + WEBHOOK_PATH`
- `ADMIN_CODE` - код для `/getadmin <code>`
- `GOOGLE_APPLICATION_CREDENTIALS` - путь к JSON-ключу сервисного аккаунта GCP

Важно:

`FIRESTORE_PROJECT_ID` должен совпадать с точным ID Google Cloud проекта, в котором находится ваша база Firestore.

## Быстрый старт

1. Экспортируйте переменные окружения.
2. Запустите бота.

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

Сидирование категорий:

```bash
cd ekapterka
export GOOGLE_APPLICATION_CREDENTIALS=<path_to_service_account_json>
export FIRESTORE_PROJECT_ID=<your-gcp-project-id>
go run ./cmd/seed
```

## Запуск через Docker

Из каталога `ekapterka`:

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

## Деплой в Google Cloud Run (с нуля)

В этом разделе используются анонимизированные placeholder-значения. Замените значения в `<...>` на свои.

### 1. Подготовьте локальные переменные

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

### 2. Авторизуйтесь и выберите проект

```bash
gcloud auth login
gcloud config set project "$PROJECT_ID"
```

При необходимости предварительно подключите billing к проекту.

### 3. Включите необходимые API

```bash
gcloud services enable \
  run.googleapis.com \
  cloudbuild.googleapis.com \
  artifactregistry.googleapis.com \
  firestore.googleapis.com \
  storage.googleapis.com \
  secretmanager.googleapis.com
```

### 4. Создайте базу Firestore (если еще не создана)

```bash
gcloud firestore databases create \
  --location="$REGION" \
  --type=firestore-native
```

Если Firestore уже существует, команда завершится ошибкой — в этом случае шаг можно пропустить.

### 5. Создайте bucket для фото

```bash
gcloud storage buckets create "gs://$BUCKET_NAME" \
  --location="$REGION" \
  --uniform-bucket-level-access
```

### 6. Сохраните чувствительные значения в Secret Manager

```bash
printf "%s" "<telegram-bot-token>" | \
  gcloud secrets create bot-token --data-file=- 2>/dev/null || \
  printf "%s" "<telegram-bot-token>" | gcloud secrets versions add bot-token --data-file=-

printf "%s" "$ADMIN_CODE" | \
  gcloud secrets create admin-code --data-file=- 2>/dev/null || \
  printf "%s" "$ADMIN_CODE" | gcloud secrets versions add admin-code --data-file=-
```

### 7. Выдайте runtime-доступ к секретам и storage

Используйте сервисный аккаунт Cloud Run (по умолчанию это обычно `<project-number>-compute@developer.gserviceaccount.com`, если вы не настроили кастомный аккаунт):

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

### 8. Задеплойте сервис (первый проход)

Деплой из исходников в каталоге `ekapterka`:

```bash
cd "$REPO_DIR/ekapterka"

gcloud run deploy "$SERVICE_NAME" \
  --source . \
  --region "$REGION" \
  --allow-unauthenticated \
  --set-env-vars "WEBHOOK_PATH=$WEBHOOK_PATH,STORAGE_ID=$BUCKET_NAME,FIRESTORE_PROJECT_ID=$FIRESTORE_PROJECT_ID" \
  --set-secrets "BOT_TOKEN=bot-token:latest,ADMIN_CODE=admin-code:latest"
```

### 9. Получите URL Cloud Run и задайте `SERVICE_URL`

```bash
export SERVICE_URL="$(gcloud run services describe "$SERVICE_NAME" \
  --region "$REGION" \
  --format='value(status.url)')"
```

Обновите env переменные сервиса, чтобы бот мог автоматически зарегистрировать Telegram webhook при старте:

```bash
gcloud run services update "$SERVICE_NAME" \
  --region "$REGION" \
  --update-env-vars "SERVICE_URL=$SERVICE_URL"
```

### 10. Один раз выполните сидирование категорий

Поскольку дерево категорий Firestore необходимо для навигации, запустите seed-команду один раз с корректными credentials проекта:

```bash
cd "$REPO_DIR/ekapterka"
go run ./cmd/seed
```

### 11. Проверьте деплой

```bash
gcloud run services describe "$SERVICE_NAME" \
  --region "$REGION" \
  --format='value(status.url)'
```

Откройте бота в Telegram и выполните `/start`.

Примечания:

- `FIRESTORE_PROJECT_ID` должен указывать на GCP-проект, в котором включен Firestore.
- `SERVICE_URL` обязателен для автоматической регистрации webhook в процессе старта.
- Для production лучше использовать выделенный сервисный аккаунт вместо дефолтного compute service account.

## Команды бота

Пользовательские команды:

- `/start` - инициализирует состояние пользователя и открывает главное меню
- `/cmd` - показывает доступные команды

Админские команды:

- `/getadmin <code>` - выдает роль `admin`, если код совпадает с `ADMIN_CODE`
- `/cat` - показывает листовые категории (`ID + title`)
- `/add` - создает предмет
- `/edit <id>` - редактирует предмет
- `/rm <id>` - удаляет предмет
- `/rent <id>` - добавляет период аренды (нужен Telegram ID арендатора)
- `/unr <rental_id>` - отменяет аренду по ID документа аренды

Формат `/add` (поддерживает многострочное описание):

```text
/add
<category_id>
<title>
<description line 1>
<description line 2>
...
```

Формат `/edit` (поддерживает многострочное описание):

```text
/edit <item_id>
<category_id>
<new title>
<description line 1>
<description line 2>
...
```

Формат `/rent`:

```text
/rent <item_id>
01.01.2025
10.02.2025
<telegram_id арендатора>
<опциональная заметка для админа>
```

## Модель данных (Firestore)

Коллекция `categories`:

- `id`, `title`, `parent_id`, `order`, `is_leaf`

Коллекция `items`:

- `title`, `description`, `category_id`, `tags`, `photo_urls`, `created_at`, `updated_at`

Коллекция `rentals`:

- `item_id`, `start`, `end`, `description`, `user_id`, `username`

Коллекция `users`:

- `id`, `role`, `created_at`, `message_id`

## Тестирование

```bash
cd ekapterka
GOCACHE=/tmp/go-build-cache go test ./...
```

## Известные ограничения

- Большая часть UI-текста бота сейчас на русском.
- Поле `tags` есть в модели, но команд управления тегами пока нет.

## Лицензия

Проект распространяется под лицензией MIT. Подробности: [LICENSE](./LICENSE).
