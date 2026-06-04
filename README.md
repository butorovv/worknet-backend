# Local Docker

Запуск backend, PostgreSQL, Redis и миграций:

```bash
docker compose up backend
```

После успешного выполнения сервис `migrate` завершится с кодом `0`, а `backend`,
`db` и `redis` продолжат работать.

Auth MVP использует cookie-based sessions:

- cookie name: `session_id`;
- `SameSite=Lax`;
- `HttpOnly=true`;
- `Secure` управляется `SESSION_COOKIE_SECURE`;
- session TTL по умолчанию: `720h`;
- active sessions хранятся только в Redis.

Повторный запуск миграций на уже заполненной схеме может упасть на существующих таблицах/индексах. Для чистого локального прогона сначала удалите volume:

```bash
docker compose down -v
```

Подключение с хоста:

```bash
psql postgres://postgres:worknet@localhost:5432/worknet?sslmode=disable
```

Остановить БД:

```bash
docker compose down
```

Остановить и удалить локальные данные:

```bash
docker compose down -v
```
