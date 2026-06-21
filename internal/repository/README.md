# internal/repository

Postgres-реализации интерфейсов из `domain`. Голый pgx, без ORM.

Каждый метод — один SQL-запрос. Ошибки pgx маппятся в `apierrors`:
- `pgx.ErrNoRows` → `apierrors.ErrNotFound`
- `pgconn.PgError` код `23505` (unique violation) → `apierrors.ErrConflict`

Получает `*pgxpool.Pool` из `app` при инициализации.
