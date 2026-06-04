# Yabloko Style Gap Analysis

Дата: 2026-06-04

Сравнение текущего `worknet-backend` с `2025_2_Yabloko` branch `main`.

Код в рамках анализа не менялся.

## 1. Какие различия еще остались

| Область | WorkNet сейчас | 2025_2_Yabloko | Различие |
|---|---|---|---|
| Runtime layout | Modular monolith: `cmd/api`, `internal/app`, `internal/modules/auth` | Multi-service layout: `cmd/auth_service`, `auth_service/cmd`, service-local `internal` | WorkNet не повторяет multi-service структуру Yabloko |
| Composition root | `internal/app` создает Postgres, Redis, repo, usecase, handler, middleware | `auth_service/cmd/app.go` создает repo/usecase, часть router factories тоже может собирать deps в delivery в других сервисах | WorkNet wiring централизованнее |
| Auth mechanism | Opaque cookie session + Redis | JWT cookie + refresh token | Auth behavior принципиально другой |
| Routes | `/api/v1/auth/register`, `/login`, `/logout`, `/me` | `/api/v0/auth/signup`, `/login`, `/refresh`, `/logout` | Public API отличается |
| DTO boundary | `delivery/transport` есть, usecase DTO не импортирует | Auth usecase импортирует `delivery/transport` и возвращает `transport.AuthResult` | WorkNet держит boundary строже |
| Response helper | `internal/platform/httpx`, error shape `{ "error": { "code", "message" } }` | `pkg/http_response`, error shape `{ "error": "..." }` | Response contract отличается |
| Logging | Минимальный `slog` в app, auth handlers без Yabloko-style contextual logs | Yabloko активно использует `pkg/logger`, `logger.FromContext`, handler/repo logs | WorkNet менее логирован |
| Request id / trace | Нет аналога `pkg/trace` в текущем auth path | `pkg/trace` и `X-Request-ID` в response sender | Trace слой не перенесен |
| Middleware set | Только auth session middleware в auth module | CORS, CSRF, access log, rate limit, JWT/session helpers | WorkNet middleware пока минимальный |
| Content-Type policy | Strict JSON decode, но нет обязательной проверки `Content-Type: application/json` | Auth handler проверяет content type | WorkNet не требует content-type |
| Cookie config | Cookie name/secure/ttl из WorkNet config, SameSite Lax фиксирован | Auth cookie читает `COOKIE_SECURE`, `COOKIE_SAMESITE`, `COOKIE_DOMAIN` прямо из env | Cookie config model отличается |
| Repository SQL | SQL inline в Go file | SQL вынесен в `repository/sql/...` и подключается через `go:embed` | SQL layout еще не Yabloko-like |
| Repository logging | Repository без contextual logs | Repository логирует start/success/error через `logger.FromContext` | Разный уровень observability |
| Repository shape | Один `AuthRepoPostgres` для users + identities, отдельный `SessionRepoRedis` | `AuthRepoPostgres` только Postgres users/password; Redis sessions нет | WorkNet добавил session repository |
| Domain model | `UserID int64`, `User`, `AuthIdentity`, `Session` | Auth domain user с string UUID и JWT-oriented auth fields | Domain отличается под схему WorkNet |
| Usecase constructor | `NewAuthUsecase(repo, sessions, hasher, tokens, ttl)` | `NewAuthUseCase(repo, secretKey)` | Constructor shape отличается из-за session/password/token deps |
| Usecase validation | Username/email/password validation под WorkNet | Email/password validation под Yabloko auth | Business validation отличается |
| Tests | Standard library fake-based tests | В Yabloko есть gomock/testify mocks и часть stale fixtures | Test style не полностью совпадает |
| Shared package placement | `internal/platform/*` | top-level `pkg/*` плюс service-local packages | Shared helpers расположены иначе |
| Docs | WorkNet добавил audit/style/migration docs | Yabloko содержит service swagger docs | Документация разного назначения |

## 2. Какие различия являются осознанными

- WorkNet остается modular monolith, а не копирует Yabloko multi-service runtime
  layout.
- `internal/app` остается composition root: delivery не создает Postgres/Redis,
  repository или usecase.
- WorkNet сохраняет cookie-based Redis sessions и не переходит на JWT.
- WorkNet не добавляет `/refresh`, JWT claims и `jwt_token`.
- WorkNet сохраняет текущие public routes и status codes.
- WorkNet сохраняет текущий error response contract через `httpx`.
- Usecase не импортирует `delivery/transport`, даже если Yabloko auth это
  делает.
- Repository не импортирует usecase DTO/params.
- Redis session repository остается WorkNet-specific частью auth.
- `go.mod`, migrations, Docker/dev setup и DB schema не приводились к Yabloko.
- Package `delivery/http` оставлен как `authhttp` на уровне package name/import
  alias, чтобы не конфликтовать ментально с `net/http`.
- SQL оставлен inline на этом шаге, потому что перенос в embedded files не
  нужен для сохранения behavior и увеличивает diff.
- Yabloko `pkg/http_response` не перенесен, потому что меняет JSON error shape.
- Yabloko logging/trace не переносились как часть Auth MVP style migration.

## 3. Какие различия стоит убрать в будущем

- Вынести SQL из `auth_repo_postgres.go` в
  `internal/modules/auth/repository/sql/auth/*.sql` через `go:embed`, если это
  будет отдельный маленький безопасный шаг.
- Добавить WorkNet-native request id/access log middleware, если это нужно для
  backend observability; ориентироваться на Yabloko `pkg/logger`/`pkg/trace`, но
  не копировать blindly.
- Решить content-type policy для JSON endpoints. Если frontend contract это
  допускает, можно добавить проверку `Content-Type: application/json` в стиле
  Yabloko.
- Добавить repository-level tests или integration tests для Postgres/Redis,
  если появится test harness.
- Унифицировать naming вокруг `AuthUsecase`/`NewAuthUsecase` в будущих модулях,
  чтобы новые modules не смешивали `Service`, `UseCase`, `Usecase`.
- Пересмотреть docs: `docs/deep-style-audit-before-refactor.md` содержит
  исторические рекомендации, часть которых уже устарела после выбора строгого
  Yabloko reference.
- Добавить shared response sender только если он сохранит WorkNet JSON contract
  и реально уменьшит duplication в следующих modules.

## 4. Какие различия лучше оставить как WorkNet-specific

- Modular monolith layout с `internal/modules`.
- `internal/app` как единственный composition root.
- Cookie-based Redis sessions вместо JWT.
- Endpoint set `register/login/logout/me` без refresh token.
- Current WorkNet response format `{ "error": { "code", "message" } }`.
- Usecase boundary с command/result structs вместо HTTP transport DTO.
- Repository dependency direction: repository не импортирует usecase.
- `internal/platform` вместо top-level `pkg`, пока helpers используются только
  внутри backend.
- WorkNet config parsing, потому что оно строже и безопаснее Yabloko env
  parsing.
- Graceful shutdown и явное закрытие Postgres/Redis ресурсов.
- Redis-backed `SessionRepoRedis`, потому что Yabloko auth reference не покрывает
  opaque session storage.
- Domain types под текущие migrations: `UserID int64`, `users`,
  `auth_identities`.
- Текущие cookie attributes и cookie name из config.
- Отсутствие Yabloko CSRF/JWT middleware в Auth MVP, пока frontend/API contract
  не требует этого.

## Итог

WorkNet уже близок к Yabloko на уровне auth module layout, file naming,
transport/middleware split, usecase/repository naming и interface placement.

Главные оставшиеся gaps являются архитектурными и behavioral: WorkNet
сознательно сохраняет modular monolith, composition root, Redis cookie sessions,
свой API contract и более строгую dependency direction.

Коммитить текущую миграцию можно как Yabloko-style alignment для Auth MVP, а
оставшиеся потенциальные улучшения лучше делать отдельными маленькими задачами.
