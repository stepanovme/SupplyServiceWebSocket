# SupplyServiceWS

WebSocket сервис на Go для отправки событий по заявкам из MySQL.

## События

- `request_created` - вставка в `request`
- `request_updated` - обновление `request` и любые insert/update/delete в `request_items`, `request_files`, `request_log`
- `request_deleted` - удаление строки из `request`
- `request_notification` - insert/update в `request_log`

Каждое событие отправляется в `/ws` и содержит:

- тип события
- `request_id`
- источник изменения (`source_table`, `source_action`)
- полный snapshot заявки (`request`, `items`, `files`, `logs`)
- для `request_deleted` отправляется `deleted_data` (данные удаленной заявки из `OLD`)
- для `request_notification` дополнительно блок `notification`

## Конфиг

Создай файл `.env` (пример в `.env.example`).

Обязательные группы БД:

- `authorization_service`
- `supply_service`
- `reference_service`

## Запуск

```bash
go mod tidy
go run ./cmd/server
```

HTTP endpoints:

- `GET /health`
- `GET /ws`

### WSS (TLS)

Если есть `cert.pem` и `key.pem`, включи в `.env`:

```env
TLS_ENABLED=true
TLS_CERT_FILE=./cert.pem
TLS_KEY_FILE=./key.pem
```

Тогда сервер поднимется по HTTPS/WSS на `APP_PORT`.
Пример WebSocket URL: `wss://your-domain/ws`

## SQL инфраструктура событий

- Файл миграции: `migrations/001_request_events.sql`
- Сервис также сам проверяет и создает таблицу `request_events` + triggers в `supply_service` при старте.

Важно: у MySQL пользователя должны быть права на `CREATE TRIGGER` и `CREATE TABLE`.
