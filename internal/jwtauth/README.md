# internal/jwtauth

JWT токены. Алгоритм HS256.

`Generate(userID, role)` — создаёт подписанный токен с `user_id`, `role` и `exp`.  
`Parse(tokenStr)` — валидирует подпись, проверяет метод подписи, возвращает `*Claims`.

`Tokenizer` — интерфейс. `transport` и `middleware` зависят от него, не от `*Manager` напрямую — удобно для тестов.

Секрет и TTL задаются при создании через `New(secret, ttl)`.
