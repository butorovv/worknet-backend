# Yabloko Style Diff Review

Дата: 2026-06-04

## Scope

Проверен текущий unstaged/untracked diff `worknet-backend` после
Yabloko-style migration.

Цель review: понять, можно ли коммитить текущие изменения.

Код в рамках review не менялся.

## Git status summary

Важно: изменения не staged, поэтому Git показывает file moves как `D` + `??`,
а `git diff --stat` не включает содержимое untracked новых файлов.

Измененные / переехавшие code areas:

- `internal/app/app.go`
- `internal/modules/auth/delivery/http`
- `internal/modules/auth/delivery/transport`
- `internal/modules/auth/delivery/middlewares`
- `internal/modules/auth/usecase`
- `internal/modules/auth/repository`

Новые docs:

- `docs/deep-style-audit-before-refactor.md`
- `docs/refactor-auth-style-report.md`
- `docs/yabloko-backend-style-guide.md`
- `docs/worknet-yabloko-refactor-plan.md`
- `docs/yabloko-style-migration-report.md`
- `docs/yabloko-style-diff-review.md`

Старый `internal/modules/auth/usecase/ports.go` удален, но его interfaces
перенесены/сведены в `internal/modules/auth/usecase/interfaces.go`.

## File moves / renames

Ожидаемые Yabloko-style moves:

| Old path | New path | Оценка |
|---|---|---|
| `delivery/http/handler.go` | `delivery/http/auth_handler.go` | Ок |
| `delivery/http/handler_test.go` | `delivery/http/auth_handler_test.go` | Ок |
| `delivery/http/dto.go` | `delivery/transport/auth_json.go` | Ок |
| `delivery/http/middleware.go` | `delivery/middlewares/middlewares.go` | Ок |
| `delivery/http/context.go` | `delivery/middlewares/context.go` | Ок |
| `usecase/service.go` | `usecase/auth_usecase.go` | Ок |
| `usecase/service_test.go` | `usecase/auth_usecase_test.go` | Ок |
| `usecase/commands.go` | `usecase/models.go` | Ок |
| `repository/postgres.go` | `repository/auth_repo_postgres.go` | Ок |
| `repository/redis_sessions.go` | `repository/session_repo_redis.go` | Ок |

Лишних переименований не найдено. Все moves соответствуют
`docs/yabloko-backend-style-guide.md`.

## Public auth API behavior

Случайной смены public API behavior не найдено.

Сохранено:

- routes:
  - `POST /api/v1/auth/register`;
  - `POST /api/v1/auth/login`;
  - `POST /api/v1/auth/logout`;
  - `GET /api/v1/auth/me`;
- success status codes:
  - register: `201`;
  - login: `200`;
  - logout: `204`;
  - me: `200`;
- method mismatch: `405`;
- invalid JSON / unknown field / trailing token: `400`;
- auth errors через текущий `httpx` response shape;
- cookie attributes:
  - configured name;
  - path `/`;
  - `HttpOnly`;
  - `Secure` from config;
  - `SameSite=Lax`;
  - logout clear with `MaxAge=-1`;
- Redis session key prefix `auth:session:`;
- `internal/app` остается composition root.

## Dependency direction

Нарушений dependency direction не найдено.

Проверено:

- `domain` не импортирует delivery/usecase/repository/platform/httpx.
- `usecase` не импортирует `delivery/transport`.
- `repository` больше не импортирует `usecase`.
- `delivery/http` импортирует `delivery/transport`, `delivery/middlewares`,
  `usecase`, `httpx`.
- `internal/app` связывает concrete repository/usecase/handler/middleware.

Это соответствует WorkNet `docs/architecture.md` и правилам style guide.

## Routes / middleware / constructors

Routes:

- `RegisterRoutes` продолжает регистрировать те же paths.
- `logout` и `me` остаются под `RequireAuth`.

Middleware:

- session cookie читается по configured name;
- missing/empty cookie возвращает `401 unauthorized`;
- valid session кладет `CurrentUser` и `SessionID` в context;
- handler читает context через `middlewares.CurrentUserFromContext` и
  `middlewares.SessionIDFromContext`.

Constructors:

- `NewAuthRepoPostgres(db)` используется в `internal/app`;
- `NewSessionRepoRedis(redisClient, "auth:session:")` используется в
  `internal/app`;
- `NewAuthUsecase(...)` используется в `internal/app`;
- `NewHandler(...)` и `middlewares.NewMiddleware(...)` получают готовые deps.

Сломанных constructor calls не найдено.

## Tests

Добавленное/расширенное покрытие соответствует задаче:

- register success + cookie;
- login success + cookie;
- login error format;
- logout success + cookie clearing;
- me success;
- invalid method;
- invalid JSON;
- unknown JSON field;
- trailing token;
- missing cookie middleware;
- route connection smoke test.

Запущенные проверки:

- `go test ./...` — passed.
- `go vet ./...` — passed.
- `docker compose config` — passed.

Дополнительно запускался `go list ./...`; пакеты были перечислены, но команда
выдала sandbox/cache warning про запись в Go module stat cache:
`operation not permitted`. Это не проявилось в `go test`/`go vet`.

## Соответствие style guide

Соответствует:

- DTO вынесены в `delivery/transport/auth_json.go`;
- middleware вынесен в `delivery/middlewares`;
- handler file naming `auth_handler.go`;
- usecase naming `AuthUsecase` / `NewAuthUsecase`;
- repository naming `AuthRepoPostgres` / `SessionRepoRedis`;
- usecase не зависит от HTTP DTO;
- repository не импортирует usecase;
- response contract WorkNet сохранен вместо прямого переноса Yabloko
  `pkg/http_response`.

Осознанные отклонения:

- package name `authhttp` сохранен вместо `http`; это допустимо style guide,
  чтобы не путать с `net/http`.
- SQL остался inline; style guide разрешает вынести SQL отдельным безопасным
  шагом, если риск низкий.

## Leftovers fix

После первичного review исправлены только leftovers из этого документа.

Изменения:

- `AuthService` в `delivery/http/auth_handler.go` сужен до методов, которые
  handler реально вызывает: `Register`, `Login`, `Logout`.
- `ValidateSession` остался только в middleware-level interface, где он реально
  используется.
- `CookieConfig` оставлен в единственном месте:
  `internal/modules/auth/delivery/errors.go`.
- `internal/app` создает один `delivery.CookieConfig` и передает его в handler
  и middleware.
- Error mapping объединен в `delivery.WriteUsecaseError`.
- `delivery/http/errors.go` удален как duplicate helper.
- `middlewares.SessionNotFound()` удален; handler использует
  `domain.ErrSessionNotFound` напрямую.

Поведение не менялось:

- routes остались прежними;
- status codes остались прежними;
- cookie attributes остались прежними;
- response JSON shape остался прежним;
- migrations/docker/go.mod не менялись.

Повторные проверки после leftovers fix:

- `gofmt` — passed.
- `goimports` — passed.
- `go test ./...` — passed.
- `go vet ./...` — passed.
- `docker compose config` — passed.

## Риски

- Unstaged rename/move лучше stage-ить аккуратно, чтобы Git распознал renames и
  review diff был читаемым.
- `docs/deep-style-audit-before-refactor.md` содержит старые рекомендации,
  которые частично противоречат текущему выбранному Yabloko direction. Это не
  кодовый риск, но в docs можно увидеть исторический конфликт.

## Вердикт

Можно коммитить.

Первичные blockers для чистого style commit устранены. Public auth API,
dependency direction, routes, middleware, constructors и tests выглядят
корректно.
