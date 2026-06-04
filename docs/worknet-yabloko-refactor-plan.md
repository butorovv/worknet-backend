# WorkNet To Yabloko Refactor Plan

Дата: 2026-06-04

## Цель

Привести текущий `worknet-backend` Auth MVP к стилю `2025_2_Yabloko` branch
`main`, сохранив поведение WorkNet:

- `POST /api/v1/auth/register`;
- `POST /api/v1/auth/login`;
- `POST /api/v1/auth/logout`;
- `GET /api/v1/auth/me`;
- cookie-based sessions;
- PostgreSQL `users` / `auth_identities`;
- Redis-backed sessions;
- текущие migrations, Docker/dev setup и graceful shutdown.

Основные правила зафиксированы в `docs/yabloko-backend-style-guide.md`.

## Текущая структура WorkNet

```txt
cmd/api/main.go
internal/app/app.go
internal/modules/auth/
  domain/
    errors.go
    identity.go
    session.go
    user.go
  usecase/
    commands.go
    interfaces.go
    service.go
    service_test.go
  repository/
    postgres.go
    redis_sessions.go
  delivery/http/
    context.go
    dto.go
    errors.go
    handler.go
    handler_test.go
    middleware.go
internal/platform/
  config/config.go
  httpx/response.go
  password/bcrypt.go
  token/random.go
```

## Целевая структура по Yabloko style

```txt
cmd/api/main.go
internal/app/app.go
internal/modules/auth/
  domain/
    errors.go
    identity.go
    session.go
    user.go
  usecase/
    auth_usecase.go
    auth_usecase_test.go
    interfaces.go
    models.go
  repository/
    auth_repo_postgres.go
    session_repo_redis.go
    sql/auth/*.sql
  delivery/
    http/
      auth_handler.go
      auth_handler_test.go
      errors.go
    transport/
      auth_json.go
    middlewares/
      middlewares.go
      context.go
internal/platform/
  config/config.go
  httpx/response.go
  password/bcrypt.go
  token/random.go
```

## File-by-file migration

| Сейчас | Цель | Действие | Причина |
|---|---|---|---|
| `cmd/api/main.go` | без изменений | не менять | entrypoint уже минимальный |
| `internal/app/app.go` | `internal/app/app.go` | обновить imports и constructors | WorkNet composition root сохраняется |
| `internal/modules/auth/delivery/http/handler.go` | `delivery/http/auth_handler.go` | rename/split | Yabloko naming handlers |
| `internal/modules/auth/delivery/http/dto.go` | `delivery/transport/auth_json.go` | move DTO/mappers | Yabloko держит HTTP JSON models в transport |
| `internal/modules/auth/delivery/http/middleware.go` | `delivery/middlewares/middlewares.go` | move | Yabloko middlewares separated from handlers |
| `internal/modules/auth/delivery/http/context.go` | `delivery/middlewares/context.go` | move | context helpers принадлежат session middleware |
| `internal/modules/auth/delivery/http/errors.go` | `delivery/http/errors.go` | оставить и адаптировать imports | HTTP error mapping остается в handler layer |
| `internal/modules/auth/delivery/http/handler_test.go` | `delivery/http/auth_handler_test.go` | rename/update package imports | tests рядом с handler |
| `internal/modules/auth/usecase/service.go` | `usecase/auth_usecase.go` | rename type/constructor | Yabloko naming `AuthUsecase`/`NewAuthUsecase` |
| `internal/modules/auth/usecase/commands.go` | `usecase/models.go` | rename | commands/results как usecase models |
| `internal/modules/auth/usecase/interfaces.go` | без изменений | оставить | интерфейсы уже там, где используются |
| `internal/modules/auth/usecase/service_test.go` | `usecase/auth_usecase_test.go` | rename/update constructor | Yabloko test naming |
| `internal/modules/auth/repository/postgres.go` | `repository/auth_repo_postgres.go` | rename concrete type/constructor | Yabloko naming `AuthRepoPostgres` |
| `internal/modules/auth/repository/redis_sessions.go` | `repository/session_repo_redis.go` | rename concrete type/constructor | Redis session store в Yabloko-compatible naming |
| inline SQL in `postgres.go` | `repository/sql/auth/*.sql` | optional later step | Yabloko uses embedded SQL files |
| `internal/platform/httpx/response.go` | без изменений | не менять сейчас | response body contract WorkNet нельзя ломать |
| `internal/platform/config/config.go` | без изменений | не менять сейчас | WorkNet config строгий и безопаснее Yabloko |

## Файлы, которые будут переименованы

- `internal/modules/auth/delivery/http/handler.go`
  -> `internal/modules/auth/delivery/http/auth_handler.go`
- `internal/modules/auth/delivery/http/dto.go`
  -> `internal/modules/auth/delivery/transport/auth_json.go`
- `internal/modules/auth/delivery/http/middleware.go`
  -> `internal/modules/auth/delivery/middlewares/middlewares.go`
- `internal/modules/auth/delivery/http/context.go`
  -> `internal/modules/auth/delivery/middlewares/context.go`
- `internal/modules/auth/delivery/http/handler_test.go`
  -> `internal/modules/auth/delivery/http/auth_handler_test.go`
- `internal/modules/auth/usecase/service.go`
  -> `internal/modules/auth/usecase/auth_usecase.go`
- `internal/modules/auth/usecase/service_test.go`
  -> `internal/modules/auth/usecase/auth_usecase_test.go`
- `internal/modules/auth/usecase/commands.go`
  -> `internal/modules/auth/usecase/models.go`
- `internal/modules/auth/repository/postgres.go`
  -> `internal/modules/auth/repository/auth_repo_postgres.go`
- `internal/modules/auth/repository/redis_sessions.go`
  -> `internal/modules/auth/repository/session_repo_redis.go`

## Папки, которые будут добавлены

- `internal/modules/auth/delivery/transport`
- `internal/modules/auth/delivery/middlewares`
- возможно `internal/modules/auth/repository/sql/auth` отдельным шагом

## Файлы, которые будут объединены

На первом этапе не объединять файлы. WorkNet уже разделен достаточно мелко.

## Файлы, которые будут разделены

На первом этапе `handler.go` будет разделен логически:

- handler methods остаются в `auth_handler.go`;
- route registration остается рядом с handler;
- JSON DTO уходят в `delivery/transport`;
- cookie helpers остаются рядом с auth handler, чтобы не менять public
  behavior cookie.

## Interfaces

Оставить принцип Yabloko:

- `delivery/http` объявляет usecase interface, который нужен handler;
- `usecase` объявляет repository/session/password/token interfaces;
- concrete repository не объявляет interface для себя.

Текущий `usecase/interfaces.go` уже соответствует этому правилу и трогается
минимально.

## DTO / transport

Перенести HTTP DTO:

- `registerRequest` -> `transport.RegisterRequest`;
- `loginRequest` -> `transport.LoginRequest`;
- `authResponse` -> `transport.AuthResponse`;
- `userResponse` -> `transport.UserResponse`;
- `authResultResponse` -> `transport.ToAuthResponse`;
- `userResponseFromCurrent` -> `transport.ToUserResponse`.

Usecase не должен импортировать `transport`.

## Helpers

Оставить:

- `internal/platform/httpx` без изменения JSON response shape;
- `internal/platform/password`;
- `internal/platform/token`;
- strict config parsing in `internal/platform/config`.

Не переносить Yabloko `pkg/http_response` напрямую, потому что он меняет error
JSON shape.

## Файлы, которые нельзя трогать

Жестко не менять:

- `db/migrations/*.sql`;
- `Dockerfile`;
- `docker-compose.yml`;
- `go.mod`;
- `go.sum`;
- public API paths and methods.

Без необходимости не менять:

- `cmd/api/main.go`;
- `internal/platform/config/config.go`;
- `internal/platform/httpx/response.go`;
- `internal/platform/password/bcrypt.go`;
- `internal/platform/token/random.go`;
- domain entity fields.

## Риски

- Package rename `authhttp` -> `http` может создать путаницу с `net/http`;
  безопаснее оставить package name `authhttp`, но применить Yabloko file/folder
  naming.
- Перенос middleware в `delivery/middlewares` затрагивает context helper imports
  в handler tests.
- Rename constructors потребует точечных изменений в `internal/app` и tests.
- Вынесение SQL в `go:embed` может изменить whitespace только на уровне SQL,
  но риск выше, поэтому лучше делать отдельным шагом после layout/naming.
- Текущий WorkNet response helper не совпадает с Yabloko, но менять его нельзя
  без риска frontend contract.

## Порядок маленьких коммитов

1. Docs: создать Yabloko style guide и этот refactor plan.
2. Delivery layout: вынести DTO в `delivery/transport`, middleware/context в
   `delivery/middlewares`, переименовать handler test file.
3. Usecase naming: `Service` -> `AuthUsecase`, `NewService` ->
   `NewAuthUsecase`, `commands.go` -> `models.go`, update tests/app.
4. Repository naming: `PostgresRepository` -> `AuthRepoPostgres`,
   `RedisSessionStore` -> `SessionRepoRedis`, update app/tests.
5. Optional repository SQL style: вынести SQL в embedded files только если
   предыдущие проверки стабильны.
6. Run checks: `gofmt`, `goimports` if available, `go test ./...`,
   `go vet ./...`, `docker compose config`.
7. Final report: `docs/yabloko-style-migration-report.md`.

## Вердикт по прошлому refactor attempt

Прошлый refactor attempt недостаточен для цели этой задачи.

Он улучшил usecase/repository boundary и частично добавил тесты, но не привел
WorkNet к Yabloko style на уровне:

- folder layout;
- file naming;
- DTO placement;
- middleware placement;
- constructor/type naming;
- repository naming;
- tests naming and coverage shape.

Нужен следующий этап, но он должен быть маленьким и сохранять business behavior.
