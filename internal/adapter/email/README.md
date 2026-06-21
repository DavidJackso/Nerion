# internal/adapter/email

Заглушка `domain.EmailSender`. Логирует вызов, ничего не отправляет.

Заменить на реальную реализацию (SMTP, SendGrid, Resend и т.д.) без изменений в остальном коде — интерфейс `domain.EmailSender` останется прежним.

Подключить: в `internal/app/main.go` раскомментировать `email.New(logger)` и передать нужным сервисам.
