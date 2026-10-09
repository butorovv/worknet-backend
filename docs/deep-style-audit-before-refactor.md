# Deep Style Audit Before Auth Refactor

- Author: Сергей
- Mode: backend
- Date: 2026-06-04
- Status: planned

## Цель

Повторно и глубже сравнить backend/auth style в `2025_2_Yabloko`, `2026_1_ASAP` и текущем `worknet-backend` после прошлого refactor attempt. Этот документ не содержит изменений кода и нужен как план перед следующим этапом.

## Проверенные ветки и источники

- `2025_2_Yabloko`: branch `main`.
- `2026_1_ASAP`: branch `dev`.
- `worknet-backend`: branch `feature/add-auth`, с незакоммиченным результатом прошлого refactor attempt.

Изученные области:

- `cmd` / bootstrap / composition root;
- `internal/app`;
- module layout;
- delivery handlers;
- middleware;
- usecase/service;
- repository;
- DTO/request/response models;
- errors/response helpers;
- config;
- PostgreSQL/Redis clients;
- tests.

## Краткий вердикт

Прошлый refactor attempt недостаточен как полноценное приведение auth к выбранному style reference.

Он полезно упростил часть usecase/repository boundary:

- убрал `Dependencies`;
- заменил `ports.go` на `interfaces.go`;
- объединил `UserRepository` и `IdentityRepository` в `AuthRepository`;
- убрал импорт `usecase` из Postgres repository;
- поправил Redis delete semantics;
- добавил route smoke test.

Но это точечная правка. Нужен следующий этап, потому что остаются архитектурные и стилевые расхождения:

- не зафиксирован единый стиль route registration и middleware naming;
- `internal/app` уже выполняет composition root роль, но станет тяжело расти без локальных builder-функций;
- delivery package смешивает handler, route registration, cookie helpers и JSON decoding в одном файле;
- usecase validation и command/result naming еще не приведены к стабильному WorkNet style;
- repository SQL inline, без выбранного правила: inline SQL или embedded SQL;
- Redis TTL считает `time.Until(session.ExpiresAt)` от wall clock, а usecase использует injected `Clock`;
- тесты покрывают happy paths, но не покрывают большую часть error mapping, JSON strictness, logout/me и repository/session edge cases.

## Главный вывод по style reference

Нельзя слепо копировать ни один старый проект.

- `2025_2_Yabloko` лучше как structural reference: `delivery/usecase/repository/domain`, простые constructors, stdlib `http.ServeMux`, минимум лишней абстракции.
- `2026_1_ASAP` лучше как auth/session reference: opaque cookie sessions, Redis session repository, интерфейсы у потребителя, request id/access middleware, более широкая auth test surface.
- Текущий `worknet-backend/docs/architecture.md` важнее обоих старых проектов там, где старые проекты противоречат WorkNet target architecture: router-фабрики не должны создавать repositories/usecases, usecase не должен получать HTTP DTO, `internal/app` остается composition root.

## Сравнительная таблица

| Область | Как в `2025_2_Yabloko` | Как в `2026_1_ASAP` | Как сейчас в WorkNet | Что нужно поменять |
|---|---|---|---|---|
| `cmd` / bootstrap | `cmd/<service>/main.go` вызывает `service/cmd.Run()`. В `Run()` часто создаются config, DB, mux, middleware. Graceful shutdown обычно отсутствует. | Один `cmd/main.go`, внутри сразу config, logger, repos, services, handlers, routes, server, graceful shutdown. | `cmd/api/main.go` тонкий: signal context, `app.New`, `Start`, `Shutdown`. | Оставить WorkNet style. Он чище обоих references для long-running backend. Не переносить heavy bootstrap в `cmd`. |
| `internal/app` | Обычно отсутствует. Composition часто в `service/cmd/app.go`, иногда router factory создает repo/usecase. | Composition прямо в `cmd/main.go`, большой файл. | `internal/app/app.go` создает config, pgxpool, Redis, auth repo/store/usecase/handlers/middleware, server. | Оставить `internal/app` как composition root. На следующем этапе выделить приватные helpers внутри `app.go` или `app` package: `newPostgresPool`, `newRedisClient`, `newRouter`, `newAuthModule`. |
| Module layout | В каждом service: `internal/domain`, `internal/usecase`, `internal/repository`, `internal/delivery/{http,transport,middlewares}`. | Monolith-level folders: `internal/handlers`, `internal/services`, `internal/repository`, `internal/domain`, `internal/dto`, `internal/middleware`. | Modular monolith: `internal/modules/auth/{domain,usecase,repository,delivery/http}`. | Оставить WorkNet module layout. Он ближе к target architecture и лучше для будущих modules. |
| Delivery handlers | Handlers часто проверяют method/content-type, decode JSON, мапят domain errors через `ResponseSender`, логируют. Interface usecase объявлен у handler. | Handlers крупнее, много inline error response, validation helper, swagger comments. Interface service объявлен у handler. | `auth/delivery/http/handler.go` тонкий, mapping DTO -> command, strict JSON decode, cookie helpers. | Сохранить thin handler. Разнести файл: `routes.go`, `handler.go`, `json.go`, `cookies.go` или оставить только если команда хочет минимализм. Добавить content-type policy только если нужен API contract. |
| Route registration | Router-фабрики вроде `NewAuthRouter`, `NewItemRouter`; иногда внутри создают repo/usecase. Используется `http.ServeMux` и path patterns. | Routes в `cmd/main.go` через `chi.Router`. | `RegisterRoutes(mux, handler, middleware)` в delivery, dependencies приходят готовыми. | Оставить dependency injection извне. Переименовать в `NewRouter` или `RegisterAuthRoutes` только если команда выберет единый naming. |
| Middleware | Service-local `delivery/middlewares`, JWT/CSRF/CORS/access log. Auth middleware часто отдельно от handler. | Global `internal/middleware`: request id, access, auth, csrf. Cookie session auth вызывает session service. | Auth middleware лежит в `auth/delivery/http`, зависит от `AuthService`, пишет auth context. Shared middleware пока нет. | Для MVP оставить auth middleware в auth delivery. В следующий этап добавить только реально shared middleware: request id/access log/CORS, если нужен frontend contract. |
| Auth/session model | Auth service использует JWT cookie, refresh, CSRF. Не подходит напрямую. | Opaque `session_id` cookie, Redis-backed sessions, CSRF token in session. Ближе к WorkNet. | Opaque `session_id` cookie, Redis active sessions, no JWT, no CSRF. | Не переходить на JWT. CSRF/CORS решить отдельно как API/security decision, не смешивать со style refactor. |
| Usecase/service | Interfaces у потребителя. В auth usecase есть validation, hashing, JWT generation, но возвращает `delivery/transport.AuthResult` - слабое место. | Services часто принимают DTO request и возвращают DTO response - не подходит целевой WorkNet boundary. | Usecase принимает commands и возвращает application structs. После refactor есть `AuthRepository`, `SessionStore`, `PasswordHasher`, `SessionTokenGenerator`, `Clock`. | Оставить command/result boundary. Доработать naming: `Service` vs `AuthUsecase`, `AuthRepository` granularity, `CurrentUser` location. Проверить, не стал ли `AuthRepository` слишком широким после объединения. |
| Interface placement | Интерфейсы лежат рядом с usecase/handler, где потребляются. Иногда отдельный `interfaces.go`. | Интерфейсы объявлены в handlers/services файлах с `go:generate mockgen`. | Delivery interface в `handler.go`, usecase interfaces в `interfaces.go`. | Это уже близко. Нужно договориться о naming: `interfaces.go` допустимо, но термин `ports` в docs/architecture все еще используется. |
| Repository | Postgres repos принимают `*pgxpool.Pool` или interface, SQL часто через `go:embed`, errors мапятся в domain. Есть repo tests. | Repos создают свои clients по config или получают clients, используют model/mapper, SQL иногда embedded package. Redis через redigo pool. | Postgres repo получает `*pgxpool.Pool`, inline SQL, domain scan напрямую, Redis store получает `go-redis` client. | Выбрать SQL style. Для MVP можно оставить inline SQL, но тогда зафиксировать правило. Если брать Yabloko style - вынести SQL в `repository/sql/auth/*.sql` с `go:embed`. |
| DTO/request/response | `delivery/transport` содержит request/response и mapping helpers. | Global `internal/dto/<domain>`, DTO иногда проходит в service. | `auth/delivery/http/dto.go` содержит request/response, usecase DTO не получает. | Текущий WorkNet лучше references. Можно переименовать package alias/DTO names только ради единообразия, но не обязательно. |
| Errors/response helpers | `pkg/http_response.ResponseSender`, response `{ "error": "..." }`. Логирует internal/user errors. | `internal/utils/response.Send`, response `{ status, errors: [...] }`. | `platform/httpx` response `{ "error": { "code", "message" } }`; auth error mapping в delivery. | Оставить WorkNet API contract. Возможно добавить `httpx.DecodeJSON` или `httpx.WriteNoContent` только если повторится в других modules. |
| Config | Simple env config per service, defaults, weak bool parse. | Large config with typed nested structs and logger bootstrap. | `platform/config.Config`, strict required `DATABASE_URL`, typed parsing for bool/int/duration. | Текущий config нормальный. Можно сгруппировать nested structs позже, когда появятся modules. Не копировать Yabloko weak bool parse. |
| PostgreSQL client | `pgxpool.New` in service cmd/app, repository gets pool/interface. | User repository creates pool itself from config, some repos own clients. | `pgxpool` создается в `internal/app`, repository получает pool. | Оставить WorkNet. Это чище и соответствует target architecture. |
| Redis client | Auth service не использует Redis sessions. | Session repository создает Redis pool по config. | Redis client создается в `internal/app`, session store получает client. | Оставить WorkNet. Добавить health/close semantics только при необходимости. |
| Transactions | Обычно repository сам делает DB operations, иногда без explicit tx. | Repository `Create` сам открывает tx. | `CreatePasswordUser` tx внутри repository. Docs target говорит tx boundary должна быть usecase-level, если нужна business atomicity. | Для auth MVP можно оставить repository tx, но явно зафиксировать как acceptable exception или ввести `CreatePasswordUser` как storage-atomic repo method. Не делать transaction manager сейчас. |
| Tests | Есть usecase/handler/repository tests, но качество неоднородное. | Есть handler/service/repo/middleware tests с gomock/testify. | Есть usecase tests, handler/middleware tests, no repository integration tests. | Добавить focused tests перед новым refactor: JSON strictness, register/logout/me, usecase errors, Redis delete not found behavior. Integration repo tests отложить до test infra. |
| Naming | `UseCase`, `RepoPostgres`, `NewXRouter`, `transport`, `middlewares`. | `Handler`, `Service`, `Repository`, `dto`, `middleware`, `utils`. | `Handler`, `Service`, `PostgresRepository`, `RedisSessionStore`, `delivery/http`, `platform/httpx`. | Выбрать WorkNet naming: `Service`, `Repository`, `Store`, `RegisterRoutes`. Не смешивать `UseCase`/`Usecase` из Yabloko без решения команды. |

## Подробная оценка WorkNet после прошлого refactor attempt

### Что уже хорошо

- `cmd/api/main.go` очень тонкий и лучше обоих references.
- `internal/app` является настоящим composition root.
- `auth` module layout соответствует target modular monolith.
- Delivery не передает HTTP DTO в usecase.
- Usecase изолирован от HTTP/Postgres/Redis/cookie details.
- Redis sessions остаются opaque cookie-based, без JWT.
- API behavior не изменен.
- Error response format единый и проще, чем в обоих references.
- Repository больше не импортирует usecase после прошлого attempt.

### Что все еще не соответствует выбранному стилю

- `handler.go` содержит слишком много разных delivery responsibilities: interface, cookie config, route registration, handlers, cookies, JSON decoder.
- `Middleware` использует тот же `AuthService` interface, что handler. Для читаемости можно выделить минимальный `SessionValidator` interface, как в references часто делают scoped interfaces.
- Usecase constructor теперь простой, но `newService` существует только ради тестового `Clock`; это нормально, но нужно решить, хотим ли public options/dependencies или private test constructor.
- `AuthRepository` после объединения стал шире. Для MVP это терпимо, но не всегда чище, чем прежние `UserRepository`/`IdentityRepository`.
- Repository SQL inline. Это нормально для малого модуля, но отличается от Yabloko embedded SQL и ASAP SQL packages.
- No request id/access logging middleware, хотя оба references имеют такой слой.
- `httpx.WriteJSON` игнорирует encode errors и не имеет centralized JSON decoder. Это допустимо сейчас, но при росте handlers может расползтись.
- No repository integration tests for Postgres/Redis.

## Конкретные файлы WorkNet, которые надо изменить на следующем этапе

Высокий приоритет:

- `internal/modules/auth/delivery/http/handler.go`
  - Разделить responsibilities или хотя бы упорядочить: routes, handler methods, cookie helpers, JSON decode.
- `internal/modules/auth/delivery/http/middleware.go`
  - Проверить интерфейс зависимости: оставить `AuthService` или выделить `SessionValidator`.
- `internal/modules/auth/delivery/http/errors.go`
  - Проверить coverage всех domain errors и оставить единый mapping.
- `internal/modules/auth/delivery/http/dto.go`
  - Оставить DTO в delivery; возможно добавить explicit mapping function names.
- `internal/modules/auth/usecase/service.go`
  - Проверить granularity `AuthRepository`; стабилизировать constructor style.
- `internal/modules/auth/usecase/interfaces.go`
  - Зафиксировать naming `interfaces` vs `ports` в code/docs.
- `internal/modules/auth/repository/postgres.go`
  - Решить SQL style: оставить inline или вынести в embedded SQL.
- `internal/modules/auth/repository/redis_sessions.go`
  - Проверить TTL against injected `Clock` decision и добавить unit/integration coverage.
- `internal/app/app.go`
  - Разбить wiring на приватные helpers перед ростом modules.
- `internal/modules/auth/delivery/http/handler_test.go`
  - Расширить tests для register/logout/me, invalid JSON, method not allowed, cookie clearing.
- `internal/modules/auth/usecase/service_test.go`
  - Добавить negative cases и session lifecycle cases.

Средний приоритет:

- `internal/platform/httpx/response.go`
  - Добавить shared helpers только если повторятся в новых modules: strict decoder, `NoContent`, method helper.
- `internal/platform/config/config.go`
  - Сгруппировать config structs позже, не сейчас.
- `docs/architecture.md`
  - Синхронизировать терминологию `ports` vs `interfaces`, если в коде закрепляем `interfaces.go`.

Низкий приоритет:

- `internal/platform/password/bcrypt.go`
- `internal/platform/token/random.go`
  - Оставить как есть, пока нет требований по argon2 или другому token generator.

## Файлы, которые трогать нельзя в следующем style refactor

Чтобы не смешать style refactor с behavioral/schema changes:

- `db/migrations/*.sql`
  - Миграции уже задают users/auth schema. Не менять без отдельной database задачи.
- `Dockerfile`
  - Не участвует в auth style drift.
- `docker-compose.yaml`
  - Не менять, если не меняется runtime config contract.
- `go.mod`
  - Не добавлять новые зависимости ради style refactor.
- `cmd/api/main.go`
  - Он уже достаточно тонкий; менять только если нужно для compile after app refactor.
- `internal/modules/*` кроме `auth`
  - Остальные modules scaffold; не использовать их для массового рефакторинга.
- `internal/platform/password/bcrypt.go`
  - Не менять hashing algorithm в style refactor.
- `internal/platform/token/random.go`
  - Не менять token generation semantics в style refactor.
- `db/migrations/003_create_users_and_auth.*.sql`
  - Особенно не менять users/auth_identities contract.
- `docs/refactor-auth-style-report.md`
  - Исторический отчет прошлого attempt, не переписывать задним числом.

## Риски

- Можно случайно изменить public API shape: status codes, cookie attributes, error codes/messages.
- Можно сломать frontend cookie flow, если менять SameSite/Secure/CORS вместе со style refactor.
- Можно ухудшить architecture, если слепо перенести Yabloko router factories, которые создают repos/usecases внутри delivery.
- Можно ухудшить boundary, если перенести ASAP style, где services принимают DTO.
- Объединенный `AuthRepository` может стать слишком широким при добавлении OAuth/password reset.
- Разнос `handler.go` на файлы безопасен, но легко превратить в churn без behavioral value.
- Вынесение SQL в files улучшит стиль по Yabloko, но увеличит diff и потребует внимательных tests.
- Redis session tests без test infra могут стать brittle; лучше начать с unit tests вокруг fake store и оставить integration отдельно.

## Порядок безопасного рефакторинга маленькими коммитами

### Commit 1: audit-only baseline

- Добавить этот отчет.
- Никакой Go-код не менять.

### Commit 2: delivery file organization

- Разделить `internal/modules/auth/delivery/http/handler.go` на:
  - `routes.go`;
  - `handler.go`;
  - `json.go`;
  - `cookies.go`.
- Не менять function signatures и public behavior.
- Запустить `go test ./internal/modules/auth/delivery/http`.

### Commit 3: delivery test expansion

- Добавить tests для:
  - `Register`;
  - `Logout`;
  - `Me`;
  - invalid method;
  - invalid JSON / unknown field / trailing token;
  - cookie clearing.
- Не менять production code, кроме минимальных helpers если тесты покажут повтор.

### Commit 4: usecase boundary decision

- Решить, оставлять `AuthRepository` или вернуть более узкие interfaces.
- Если оставлять `AuthRepository`, закрепить это в report/docs как MVP simplification.
- Если возвращать узкие interfaces, сделать это без `CreatePasswordUserParams` и без repository import of usecase.

### Commit 5: usecase test expansion

- Добавить cases:
  - duplicate user;
  - invalid register input;
  - invalid login email/password;
  - inactive user;
  - missing session;
  - logout empty session id;
  - session store errors.

### Commit 6: repository style decision

- Выбрать один из вариантов:
  - оставить inline SQL и зафиксировать как MVP rule;
  - вынести SQL в `internal/modules/auth/repository/sql/*.sql` с `go:embed`.
- Не менять schema и business semantics.

### Commit 7: app wiring cleanup

- Разбить `internal/app/app.go` на небольшие private helpers.
- Не переносить composition в delivery.
- Проверить graceful shutdown и route registration.

### Commit 8: optional platform helpers

- Только если появится повторение:
  - strict JSON decoder в `platform/httpx`;
  - common method not allowed helper;
  - no-content helper.
- Не добавлять request id/access log/CORS в этот refactor, если нет отдельного API/runtime решения.

## Что считать success criteria

- `register/login/logout/me` сохраняют behavior.
- Cookie-based sessions сохраняются.
- PostgreSQL users/auth_identities и Redis sessions сохраняются.
- `cmd/api` остается thin.
- `internal/app` остается composition root.
- Delivery не создает repositories/usecases.
- Usecase не импортирует HTTP DTO.
- Repository не импортирует usecase.
- Tests покрывают auth delivery/usecase edge cases.
- `go test ./...` и `go vet ./...` проходят.

## Итоговый план следующего этапа

Следующий refactor нужен, но он должен быть не большим переписыванием, а серией маленьких mechanical commits. Главная цель следующего этапа - не менять business behavior, а стабилизировать style boundaries:

1. Сначала delivery organization и tests.
2. Потом usecase interface decision.
3. Потом repository SQL style decision.
4. Потом cleanup `internal/app`.
5. Только после этого рассматривать shared middleware/platform helpers.

Прошлый refactor attempt можно оставить как первый полезный шаг, но нельзя считать задачу завершенной.
