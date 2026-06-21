# pkg

Общие утилиты без бизнес-логики и без зависимости от `domains`/`entities`.

Примеры пакетов:
- `apierrors/` — типизированные HTTP-ошибки
- `pagination/` — структуры для пагинации
- `validator/` — валидация входных данных

Если пакет зависит от `domains` или `entities` — он не `pkg`, он `adapter` или `service`.
