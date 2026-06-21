# cmd/migrate

Отдельный бинарь для применения SQL-миграций через goose.

Запускается до старта основного сервера — в CI, в Dockerfile entrypoint или вручную.

```bash
APP_DB_DSN=postgres://user:pass@localhost:5432/db go run ./cmd/migrate
```

Миграции берёт из пакета `migrations` (embedded FS — файлы компилируются в бинарь).
