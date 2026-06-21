# internal/middleware

HTTP middleware.

**`Auth(jm)`** — проверяет `Authorization: Bearer <token>`, парсит JWT, кладёт `*Claims` в контекст. Без токена → 401.

**`RequireRole(role)`** — проверяет роль из claims в контексте. Требует `Auth` перед собой. Неверная роль → 403.

**`RequestLogger(logger)`** — логирует каждый запрос через slog: метод, путь, статус, байты, длительность, request_id.

**`ClaimsFrom(ctx)`** — хелпер для хэндлеров: достаёт `*Claims` из контекста без прямой зависимости на тип ключа.
