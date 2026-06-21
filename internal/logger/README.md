# internal/logger

Создаёт `*slog.Logger`.

`New(level, format)` — принимает уровень (`debug`/`info`/`warn`/`error`) и формат (`json`/`text`).  
Неизвестный уровень → fallback на `info`.  
Пишет в stdout.

json-формат — для продакшна и парсинга лог-агрегаторами.  
text-формат — для локальной разработки.
