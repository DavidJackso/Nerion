# go-template

Production-ready Go HTTP server template. JWT auth, Postgres, structured logging, graceful shutdown — всё уже подключено.

## Stack

- **chi** — роутер
- **pgx** — Postgres driver (connection pool)
- **goose** — SQL миграции
- **golang-jwt** — JWT HS256
- **viper** — конфиг из файла + env

## Quick start

```bash
# 1. переименовать модуль
make rename MODULE=github.com/you/myapp

cp config.yaml.example config.yaml
# заполни db.dsn и jwt.secret в config.yaml

make migrate   # применить миграции
make run       # запустить сервер
```

Или через env (без config.yaml):

```bash
APP_DB_DSN=postgres://user:pass@localhost:5432/db \
APP_JWT_SECRET=supersecret \
make run
```

## Конфиг

| Env var | yaml ключ | Дефолт | Обязателен |
|---|---|---|---|
| `APP_DB_DSN` | `db.dsn` | — | да |
| `APP_JWT_SECRET` | `jwt.secret` | — | да |
| `APP_JWT_TTL` | `jwt.ttl` | `24h` | нет |
| `APP_HTTP_ADDR` | `http.addr` | `:8080` | нет |
| `APP_LOG_LEVEL` | `log.level` | `info` | нет |
| `APP_LOG_FORMAT` | `log.format` | `json` | нет |

Env переменные перекрывают config.yaml.

## API

| Метод | Путь | Доступ | Описание |
|---|---|---|---|
| `POST` | `/auth/login` | публичный | получить JWT |
| `GET` | `/users/{id}` | свой id или admin | получить пользователя |
| `GET` | `/users` | admin | список пользователей |
| `POST` | `/users` | admin | создать пользователя |

## Структура

```
cmd/            — точки входа (main.go, cmd/migrate/main.go)
internal/
  app/          — сборка зависимостей, lifecycle сервера
  config/       — загрузка конфига
  domain/       — интерфейсы (контракты между слоями)
  entity/       — доменные структуры
  service/      — бизнес-логика
  repository/   — работа с БД
  transport/http/— HTTP хэндлеры и роутинг
  middleware/   — JWT auth, логирование
  jwtauth/      — генерация и валидация токенов
  logger/       — slog factory
  adapter/email/— заглушка email-отправителя
migrations/     — SQL миграции (goose)
pkg/apierrors/  — типизированные HTTP-ошибки
```

## Make

```bash
make rename MODULE=github.com/you/myapp  # переименовать модуль
make run                                 # запустить
make build                               # собрать бинарь в ./bin/server
make migrate                             # применить миграции
make test                                # тесты
make lint                                # go vet
make tidy                                # go mod tidy
```
