# Yabloko Style Migration Report

Дата: 2026-06-04

## 1. Цель

Привести текущий WorkNet Auth MVP к backend style reference
`2025_2_Yabloko` branch `main`, сохранив бизнес-поведение:

- register;
- login;
- logout;
- me;
- cookie-based sessions;
- PostgreSQL users/auth_identities;
- Redis sessions;
- migrations;
- Docker/dev setup;
- graceful shutdown.

## 2. Style reference

Основной style reference: `2025_2_Yabloko` branch `main`.

Причина:

- понятный `delivery/usecase/repository/domain` layout;
- DTO вынесены в `delivery/transport`;
- middleware вынесены в `delivery/middlewares`;
- repository names вида `XRepoPostgres`;
- usecase names вида `XUsecase`;
- interfaces лежат рядом с потребителями.

## 3. Что взято из 2025_2_Yabloko

- File naming:
  - `auth_handler.go`;
  - `auth_json.go`;
  - `auth_usecase.go`;
  - `auth_repo_postgres.go`.
- Delivery layout:
  - handlers отдельно;
  - transport DTO отдельно;
  - middlewares отдельно.
- Constructor/type naming:
  - `NewAuthUsecase`;
  - `NewAuthRepoPostgres`;
  - `NewSessionRepoRedis`.
- Interface placement:
  - delivery-level usecase interface в handler package;
  - usecase-level repository/session interfaces в usecase package.
- Tests рядом с кодом и с проверкой HTTP status/body/cookie behavior.

## 4. Что не перенесено из 2025_2_Yabloko

- JWT auth и refresh-token API не перенесены, потому что WorkNet использует
  opaque cookie sessions.
- Создание repository/usecase внутри router factory не перенесено, потому что
  WorkNet `docs/architecture.md` требует `internal/app` composition root.
- Зависимость usecase от `delivery/transport` не перенесена, потому что WorkNet
  usecase не должен зависеть от HTTP DTO.
- Yabloko `pkg/http_response` не перенесен напрямую, потому что он меняет
  WorkNet error response JSON contract.
- Multi-service runtime layout не перенесен, потому что WorkNet остается modular
  monolith.
- SQL не вынесен в embedded files на этом шаге, чтобы не увеличивать diff и
  риск вокруг рабочей auth persistence.

## 5. Что взято из 2026_1_ASAP

Новый код из `2026_1_ASAP` не переносился.

Сохранена уже выбранная WorkNet auth/session mechanics, совместимая с ASAP
идеями:

- opaque session id в cookie;
- Redis-backed session store;
- middleware validates session and writes current user/session into context.

## 6. Структура до

```txt
internal/modules/auth/
  delivery/http/
    context.go
    dto.go
    errors.go
    handler.go
    handler_test.go
    middleware.go
  repository/
    postgres.go
    redis_sessions.go
  usecase/
    commands.go
    interfaces.go
    service.go
    service_test.go
```

## 7. Структура после

```txt
internal/modules/auth/
  delivery/
    http/
      auth_handler.go
      auth_handler_test.go
      errors.go
    middlewares/
      context.go
      middlewares.go
    transport/
      auth_json.go
  repository/
    auth_repo_postgres.go
    session_repo_redis.go
  usecase/
    auth_usecase.go
    auth_usecase_test.go
    interfaces.go
    models.go
```

## 8. Переименованные файлы

- `internal/modules/auth/delivery/http/handler.go`
  -> `internal/modules/auth/delivery/http/auth_handler.go`
- `internal/modules/auth/delivery/http/handler_test.go`
  -> `internal/modules/auth/delivery/http/auth_handler_test.go`
- `internal/modules/auth/delivery/http/dto.go`
  -> `internal/modules/auth/delivery/transport/auth_json.go`
- `internal/modules/auth/delivery/http/middleware.go`
  -> `internal/modules/auth/delivery/middlewares/middlewares.go`
- `internal/modules/auth/delivery/http/context.go`
  -> `internal/modules/auth/delivery/middlewares/context.go`
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

## 9. Удаленные файлы

Функционально файлы не удалялись. Старые paths исчезли из-за rename/move.

`internal/modules/auth/usecase/ports.go` уже был удален предыдущим refactor
attempt и в этом шаге не восстанавливался.

## 10. Созданные файлы

- `docs/yabloko-backend-style-guide.md`
- `docs/worknet-yabloko-refactor-plan.md`
- `docs/yabloko-style-migration-report.md`
- `internal/modules/auth/delivery/transport/auth_json.go`
- `internal/modules/auth/delivery/middlewares/context.go`
- `internal/modules/auth/delivery/middlewares/middlewares.go`
- `internal/modules/auth/delivery/http/auth_handler.go`
- `internal/modules/auth/delivery/http/auth_handler_test.go`
- `internal/modules/auth/usecase/auth_usecase.go`
- `internal/modules/auth/usecase/auth_usecase_test.go`
- `internal/modules/auth/usecase/models.go`
- `internal/modules/auth/repository/auth_repo_postgres.go`
- `internal/modules/auth/repository/session_repo_redis.go`

## 11. Сохраненное public API behavior

Сохранено:

- `POST /api/v1/auth/register`;
- `POST /api/v1/auth/login`;
- `POST /api/v1/auth/logout`;
- `GET /api/v1/auth/me`;
- status codes:
  - register success: `201`;
  - login success: `200`;
  - logout success: `204`;
  - me success: `200`;
  - invalid method: `405`;
  - invalid JSON: `400`;
  - missing/invalid auth: `401`;
- response helper `internal/platform/httpx` и текущий error JSON shape;
- cookie name from config;
- cookie path `/`;
- `HttpOnly`;
- `Secure` from config;
- `SameSite=Lax`;
- logout cookie clearing with `MaxAge=-1`;
- Redis key prefix `auth:session:`;
- graceful shutdown in `internal/app`.

## 12. Запущенные проверки

- `gofmt -w ...`
- `goimports -w ...`
- `go test ./internal/modules/auth/delivery/http`
- `go test ./internal/modules/auth/...`
- `go test ./...`
- `go vet ./...`
- `docker compose config`

## 13. Что прошло

Прошли:

- `go test ./internal/modules/auth/delivery/http`
- `go test ./internal/modules/auth/...`
- `go test ./...`
- `go vet ./...`
- `docker compose config`

## 14. Оставшиеся риски

- Git видит renames как delete/add до staging, потому что изменения еще не
  staged.
- SQL пока остается inline в repository; это осознанно оставлено для отдельного
  безопасного шага.
- `internal/platform/httpx` стилистически отличается от Yabloko
  `pkg/http_response`, но менять его сейчас рискованно для frontend contract.
- `internal/app` остается WorkNet composition root, хотя часть Yabloko router
  factories делает wiring внутри delivery.

## 15. Что проверить вручную

- Через running stack:
  - register sets session cookie;
  - login sets session cookie;
  - me returns current user with cookie;
  - logout clears cookie and deletes Redis session;
  - повторный me после logout возвращает unauthorized.
- Frontend contract:
  - error JSON shape не сломан;
  - cookie name matches expected environment.
