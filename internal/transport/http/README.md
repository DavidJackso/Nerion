# internal/transport/http

HTTP-сервер на chi. Роутинг, middleware, хэндлеры.

**Middleware стек** (применяется ко всем маршрутам):
- `RequestID` — генерирует уникальный ID запроса
- `RealIP` — берёт реальный IP из заголовков прокси
- `RequestLogger` — slog-совместимое логирование запросов
- `Recoverer` — перехватывает panic, возвращает 500
- `CleanPath` — нормализует URL пути
- `Timeout(8s)` — отменяет контекст по таймауту

**Маршруты:**
- `POST /auth/login` — публичный
- `GET /users/{id}` — требует JWT, только свой id или admin
- `GET /users`, `POST /users` — требует JWT + роль admin

Хэндлеры работают через `domain.UserService` — не напрямую с репозиторием.  
Ошибки маппятся через `writeError`: `*apierrors.APIError` → нужный HTTP статус, остальное → 500.
