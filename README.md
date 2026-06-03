# Local Docker

Запуск PostgreSQL с применением миграций:

```bash
docker compose up postgres migrate
```

После успешного выполнения сервис `migrate` завершится с кодом `0`, а `postgres` продолжит работать.

Повторный запуск миграций на уже заполненной схеме может упасть на существующих таблицах/индексах. Для чистого локального прогона сначала удалите volume:

```bash
docker compose down -v
```

Подключение с хоста:

```bash
psql postgres://worknet:worknet@localhost:5432/worknet?sslmode=disable
```

Остановить БД:

```bash
docker compose down
```

Остановить и удалить локальные данные:

```bash
docker compose down -v
```
