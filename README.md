# Nerion

Бэкенд для управления данными. Создаёшь пространство, описываешь схему таблиц, закидываешь записи — и получаешь REST API с ключами. Что-то среднее между Airtable и headless CMS, только своё и без подписки.

Написано на Go. PostgreSQL, chi, pgx, goose, JWT.

---

## Что умеет

**Пространства (Spaces)** — изолированные воркспейсы со своей схемой данных. Каждый space получает отдельную Postgres-схему. Можно звать участников, раздавать роли.

**Схема** — создаёшь таблицы и поля через API. Типы: `text`, `number`, `date`, `boolean`, `enum`, `email`, `file`, `relation` и ещё несколько. Поля можно менять, DDL пересоздаётся.

**Записи** — CRUD поверх твоих таблиц. Пагинация, сортировка, полнотекстовый поиск.

**Публичный API** — выдаёшь API-ключ с нужным scope (`read` / `read_write`), и любой клиент ходит напрямую на `/api/{space}/{table}`. OpenAPI-спека генерируется автоматически.

**Файлы** — загрузка через `/files/upload`. Хранилище переключается через конфиг: локальная папка для разработки, S3/minio для прода.

**Lists** — публичные представления таблиц. Фиксируешь таблицу, опционально фильтр, публикуешь — и раздаёшь ссылку без авторизации.

**PDF** — загружаешь шаблон, маппишь поля, генерируешь пачку документов по записям.

**Приглашения** — invite по email, принять можно только с нужным адресом.

**Аудит** — всё что происходит через API-ключи пишется в лог, смотришь через `/spaces/{slug}/audit`.

---

## Быстрый старт

```bash
cp config.yaml.example config.yaml
# поправь db.dsn и jwt.secret

make migrate
make run
```

Или без файла конфига:

```bash
APP_DB_DSN=postgres://user:pass@localhost:5432/nerion \
APP_JWT_SECRET=что-нибудь-длинное \
make run
```

Сервер на `:8080`.

### Docker

```bash
cp .env.example .env   # заполни переменные

docker compose --profile migrate run migrate
docker compose up -d app
```

---

## Конфиг

| Переменная | Обязательна | По умолчанию |
|---|---|---|
| `APP_DB_DSN` | да | — |
| `APP_JWT_SECRET` | да | — |
| `APP_JWT_TTL` | нет | `24h` |
| `APP_HTTP_ADDR` | нет | `:8080` |
| `APP_LOG_LEVEL` | нет | `info` |
| `APP_LOG_FORMAT` | нет | `json` |
| `APP_STORAGE_S3_BUCKET` | нет | — (без этого — локальное хранилище) |
| `APP_STORAGE_S3_ENDPOINT` | нет | — |
| `APP_STORAGE_S3_ACCESS_KEY` | нет | — |
| `APP_STORAGE_S3_SECRET_KEY` | нет | — |
| `APP_STORAGE_UPLOAD_DIR` | нет | `./uploads` |
| `APP_STORAGE_PRESIGN_TTL` | нет | `1h` |

Env-переменные перекрывают `config.yaml`.

---

## Auth

Два токена: короткий JWT (доступ) + opaque refresh-токен (хранится как SHA-256 хэш в БД).

```
POST /auth/register       — регистрация, отправляет письмо с подтверждением
POST /auth/login          — вход, возвращает пару токенов
POST /auth/refresh        — ротация refresh-токена
POST /auth/logout         — отзыв сессии
POST /auth/verify-email   — подтверждение email
POST /auth/password/reset-request
POST /auth/password/reset
GET  /auth/me             — текущий пользователь из JWT
```

Для защищённых маршрутов — `Authorization: Bearer <token>`.  
Для публичного API — `X-Api-Key: <key>`.

---

## Структура

```
cmd/               — main.go, cmd/migrate/main.go
internal/
  app/             — сборка зависимостей, lifecycle
  config/          — viper
  domain/          — интерфейсы репозиториев и сервисов
  entity/          — структуры данных
  service/         — логика
  repository/      — SQL-запросы
  transport/http/  — хэндлеры, роутинг
  middleware/      — auth, роли, логирование
  jwtauth/         — HS256
  adapter/
    email/         — заглушка (подключи свой SMTP)
    storage/       — local + S3
migrations/        — SQL (goose, embedded)
pkg/apierrors/     — типизированные ошибки
```

---

## Make

```bash
make run     # запустить
make build   # ./bin/server
make migrate # применить миграции
make test    # тесты
make lint    # go vet
make tidy    # go mod tidy
make rename MODULE=github.com/you/myapp  # сменить module path
```

---

## Лицензия

MIT
