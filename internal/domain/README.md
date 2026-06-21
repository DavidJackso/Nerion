# internal/domain

Только интерфейсы. Никаких реализаций, никаких импортов из других внутренних пакетов кроме `entity`.

`UserRepository` — контракт между `service` и `repository`.  
`UserService` — контракт между `transport` и `service`.  
`EmailSender` — контракт для адаптеров отправки почты.

Правило: если `transport` хочет вызвать `service` — он делает это через интерфейс из `domain`. Реализации подставляет `app`.
