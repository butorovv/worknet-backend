# Auth Yabloko Style Audit

- Author: Георгий
- Mode: backend
- Date: 2026-06-04
- Session: `sessions/2026-06-04/2333-backend-georgiy.md`
- Status: audit-only

## Цель

Сравнить текущий auth-модуль WorkNet backend с `auth_service` из
`go-park-mail-ru/2025_2_Yabloko`, branch `main`, и подготовить target structure
для безопасного рефакторинга без изменения public API и business behavior.

Этот документ не меняет код. Он фиксирует аудит, рекомендации и план будущих
маленьких refactor steps.

## Источники

- WorkNet backend: `../worknet-backend`, branch `feature/add-auth`,
  latest commit `b66fdf9 Merge pull request #4 from butorovv/feature/add-auth2`.
- Yabloko: `../2025_2_Yabloko`, branch `main`,
  latest commit `da11ea7 РК2 (#24)`.
- ASAP для chi-only справки: `../2026_1_ASAP`, branch `dev`,
  latest commit `1f15702 Merge pull request #55 from go-park-mail-ru/feature/add-channel`.

Ограничение источника: в локальном `2025_2_Yabloko` есть untracked `tags.json`,
но `auth_service` файлы прочитаны с branch `main`.

## Constraints

- Не трогать DB schema и migrations.
- Не трогать Redis contract.
- Не трогать Docker и `docker-compose.yaml`.
- Не трогать `go.mod` / `go.sum` без отдельного решения.
- Не менять routes behavior, cookie attributes, public API и business behavior
  `register/login/logout/me/session validation`.
- Не внедрять `chi-router` без отдельного подтверждения.
- Не делать глобальный backend refactor за пределами auth и composition root.

## Краткий вердикт

Текущий WorkNet auth уже во многом приведен к Yabloko-like структуре:

- есть `domain`, `usecase`, `repository`, `delivery/http`,
  `delivery/transport`, `delivery/middlewares`;
- файлы уже называются `auth_usecase.go`, `auth_repo_postgres.go`,
  `session_repo_redis.go`, `auth_handler.go`, `auth_json.go`;
- usecase не импортирует HTTP DTO;
- repository не импортирует usecase;
- composition root находится в `internal/app`, а не внутри delivery.

Главное отличие: WorkNet сознательно строже Yabloko по dependency direction и
сохраняет cookie-based Redis sessions вместо JWT. Это нужно оставить. Дальнейший
рефактор стоит делать как серию маленьких независимых шагов, а не как перенос
Yabloko один в один.

## Accepted strict Yabloko-style decisions

После дополнительной сверки с Yabloko и текущим WorkNet принято двигаться ближе
к Yabloko-style в следующих местах:

- Handler naming:
  - `Handler` -> `AuthHandler`;
  - `NewHandler` -> `NewAuthHandler`;
  - package name `authhttp` оставить, потому что он лучше различает локальный
    package и stdlib `net/http`.
- Handler interface naming:
  - `AuthUsecase` в `delivery/http` -> `AuthUseCaseInterface`;
  - handler field оставить `uc`;
  - набор методов интерфейса не менять.
- Usecase naming:
  - `AuthUsecase` -> `authUseCase`;
  - `NewAuthUsecase` -> `NewAuthUseCase`;
  - это ближе к Yabloko, где concrete implementation unexported, а constructor
    exported. Для WorkNet это приемлемо, потому что модуль живет внутри
    `internal`, а внешние слои получают значение через constructor и narrow
    interfaces.
- Route registration:
  - текущий private `registerAuthRoutes` в `internal/app/app.go` оставить до
    отдельного routing step;
  - `chi` рассматривать отдельным implement step, не смешивать с naming.
- Repository interfaces:
  - `usecase/interfaces.go` оставить как место для ports конкретного auth
    usecase/module;
  - это не глобальная папка DTO/interfaces для всех usecases проекта.
- Usecase models:
  - `RegisterCommand`, `LoginCommand`, `LogoutCommand`,
    `ValidateSessionCommand`, `AuthResult`, `CurrentUser` считать
    application/usecase-level models;
  - usecase не должен импортировать `delivery/transport`, даже если так сделано
    в Yabloko.
- Repository SQL:
  - inline SQL можно вынести в embedded `.sql` files по стилю Yabloko;
  - это отдельный step после naming и отдельно от `chi`.

## Текущая структура WorkNet auth

```txt
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
  delivery/
    errors.go
    http/
      auth_handler.go
      auth_handler_test.go
    middlewares/
      context.go
      middlewares.go
    transport/
      auth_json.go
```

Composition root и entrypoint:

```txt
cmd/api/main.go
internal/app/app.go
```

## Структура Yabloko auth_service

```txt
auth_service/
  cmd/app.go
  internal/
    config/
    domain/
      auth.go
      errors.go
    usecase/
      auth_usecase.go
      auth_usecase_test.go
      mock/
    repository/
      auth_repo_postgres.go
      db.go
      sql/auth/*.sql
    delivery/
      http/
        auth_handler.go
        auth_handler_test.go
        mock/
      middlewares/
        middlewares.go
      transport/
        auth_json.go
```

Yabloko полезен как reference по file layout и naming. Но он не подходит для
слепого переноса, потому что:

- auth использует JWT cookie и refresh-token flow;
- usecase импортирует `delivery/transport`;
- cookie settings частично читаются напрямую из env в handler;
- `cmd/app.go` совмещает config, DB, router, middlewares и server start;
- `auth_service` является отдельным service, а WorkNet backend остается
  modular monolith.

## Сравнение WorkNet и Yabloko

| Область | Yabloko | Почему так в Yabloko | WorkNet сейчас | Подходит WorkNet | Переносить | Риск |
|---|---|---|---|---|---|---|
| Папки | `internal/{domain,usecase,repository,delivery}` внутри service | Отдельный `auth_service` имеет свой полный vertical slice | `internal/modules/auth/{domain,usecase,repository,delivery}` | Да, но в modular monolith форме | Уже перенесено по смыслу | Низкий |
| Файлы | `auth_usecase.go`, `auth_repo_postgres.go`, `auth_handler.go`, `auth_json.go` | Явный domain prefix упрощает навигацию | WorkNet уже использует эти имена | Да | Оставить | Низкий |
| Package names | `http`, `transport`, `middlewares`, `usecase`, `repository` | Короткие package names внутри service | WorkNet: `authhttp` package name для `delivery/http`, остальные похожи | Да; `authhttp` удобнее, чтобы не путать с `net/http` | Не менять | Низкий |
| Handler/delivery | Handler читает JSON, проверяет method/content-type, ставит JWT cookie, мапит errors | Delivery владеет HTTP protocol и cookie behavior | Handler тонкий, мапит DTO в command, ставит `session_id`, строгий JSON decode | Да | Handler не дробить; route wiring вынести ближе к `internal/app` | Низкий |
| Usecase/service | Concrete `authUseCase`, constructor `NewAuthUseCase`; interface у handler | Usecase содержит auth сценарии и validation | Package `usecase`; после Step 1 handler уже использует `uc`, но concrete usecase еще называется `AuthUsecase` / `NewAuthUsecase` | Да, naming нужно довести до strict Yabloko-style | `AuthUsecase` -> `authUseCase`, `NewAuthUsecase` -> `NewAuthUseCase`, handler interface -> `AuthUseCaseInterface` | Низкий |
| Repository | `AuthRepoPostgres`, embedded SQL, методы `CreateUser/GetUser...` | SQL вынесен из Go, repo отвечает за persistence | `AuthRepoPostgres` с inline SQL, `SessionRepoRedis` для Redis sessions | Частично | SQL не выносить сейчас; repo names оставить | Средний при выносе SQL |
| Domain/entity/errors | `User` содержит ID/email/hash; errors текстовые | Минимальная auth domain под JWT flow | `User`, `Session`, `AuthIdentity`, typed status/errors | Да, WorkNet богаче из-за schema/session | Оставить | Низкий |
| Middleware | Service-level middlewares: access log, CSRF, CORS, rate limit | Auth service сам собирает HTTP boundary | WorkNet auth middleware только session validation | Да для MVP | Не переносить CSRF/CORS/rate limit в style refactor | Средний |
| Constructors | Простые `NewAuthRepoPostgres`, `NewAuthUseCase`, `NewAuthHandler` | Dependencies передаются явно | WorkNet аналогично, dependencies создаются в `internal/app` | Да | Оставить | Низкий |
| Error mapping | Handler switch + `ResponseSender` | Ошибки логируются и превращаются в HTTP response | `delivery.WriteUsecaseError` + `platform/httpx` | Да, WorkNet API contract другой | Оставить WorkNet format | Высокий при переносе Yabloko response |
| Dependency direction | Usecase импортирует `delivery/transport` | В Yabloko это упрощает возврат JWT DTO, но смешивает слои | WorkNet usecase не импортирует delivery | WorkNet лучше | Не переносить Yabloko direction | Высокий |
| Composition root | `auth_service/cmd/app.go` создает config, DB, repo, usecase, mux | Отдельный service bootstrap | `internal/app/app.go`, `cmd/api/main.go` тонкий | Да, WorkNet чище | Оставить | Средний при переносе |
| Routing | `http.ServeMux`, `NewAuthRouter(mux, base, uc)` | Service-local router factory | Route registration уже перенесен в `internal/app` private helper `registerAuthRoutes` | Да для composition root; `chi` требует отдельного решения | Следующий routing step: отдельная migration to `chi` с тестами public routes | Средний |

## Ответы на ключевые вопросы

### 1. `service` или `usecase`

В текущем actual code WorkNet слой бизнес-логики уже находится в package
`usecase`:

- package: `internal/modules/auth/usecase`;
- concrete type: `AuthUsecase`;
- constructor: `NewAuthUsecase`;
- file: `auth_usecase.go`.

После Step 1 слово `service` убрано из handler/app/test naming. До этого оно
оставалось как local naming:

- `AuthService` interface в `delivery/http/auth_handler.go`;
- `service` field внутри handler;
- `authService` variable в `internal/app/app.go`;
- `service` variable в tests.

Почему так могло появиться: в Go-проектах handler часто зависит от абстрактного
`service`, даже если папка называется `usecase`. Это короткое consumer-side имя,
но оно создает терминологический шум для WorkNet, где целевая architecture
говорит именно `usecase`.

Как в Yabloko:

- concrete usecase называется `authUseCase`;
- constructor `NewAuthUseCase`;
- handler interface называется `AuthUseCaseInterface`;
- handler field называется `uc`.

Обновленная рекомендация для WorkNet:

- не создавать папку `auth/service`;
- оставить package `auth/usecase`;
- привести usecase naming ближе к Yabloko:
  - concrete implementation `AuthUsecase` -> `authUseCase`;
  - constructor `NewAuthUsecase` -> `NewAuthUseCase`;
- привести handler naming ближе к Yabloko:
  - `Handler` -> `AuthHandler`;
  - `NewHandler` -> `NewAuthHandler`;
  - handler-side interface `AuthUsecase` -> `AuthUseCaseInterface`;
- оставить field name `uc` и local variables `uc/authUsecase`;
- не менять бизнес-логику и signatures сценариев `Register`, `Login`,
  `Logout`, `ValidateSession`.

Риск: низкий, если это только naming. Проверка: `go test ./...`, `go vet ./...`.

### 2. Где держать interfaces repository/usecase

Yabloko держит `AuthRepository` прямо в `auth_usecase.go`. Это удобно для
маленького JWT usecase с одним repository.

В WorkNet interfaces сейчас лежат в `usecase/interfaces.go`:

- `AuthRepository`;
- `SessionStore`;
- `PasswordHasher`;
- `SessionTokenGenerator`;
- `Clock`.

Для WorkNet отдельный `interfaces.go` оправдан, потому что auth usecase зависит
не от одного Postgres repo, а от нескольких ports: PostgreSQL users/identities,
Redis sessions, hasher, token generator, clock.

Можно ли писать interfaces прямо в `auth_usecase.go`: да, технически можно, но
файл станет менее читаемым. Для WorkNet лучше оставить `interfaces.go`.

Нужно ли отдельное место для интерфейсов, если repo слой разделен на
`auth_repo` и `session_repo`: да. Usecase interfaces должны жить у потребителя,
то есть в `usecase`, а не рядом с concrete repository. Отдельный файл помогает
держать границу видимой.

Рекомендация:

- оставить `usecase/interfaces.go`;
- не возвращать `ports.go`, чтобы naming был ближе к Yabloko и старым проектам;
- не дробить `interfaces.go` на несколько файлов, пока список маленький.

Риск изменения: если переносить interfaces между файлами без изменения package,
риск низкий. Но в этом нет необходимости.

### 3. Модели auth

#### Delivery DTO / request / response

Файл: `internal/modules/auth/delivery/transport/auth_json.go`.

- `RegisterRequest`: public JSON body для `register`.
- `LoginRequest`: public JSON body для `login`.
- `AuthResponse`: public JSON response для `register/login/me`.
- `UserResponse`: public JSON user shape.

Зачем нужны: это HTTP contract. Они не должны протекать в usecase, потому что
JSON tags, response shape и field visibility относятся к delivery layer.

Дублирование:

- `UserResponse` похож на `usecase.CurrentUser`;
- это допустимое дублирование ради boundary между application result и HTTP
  response.

Что можно упростить:

- не убирать `UserResponse`;
- можно оставить mapping functions `ToAuthResponse` / `ToUserResponse`;
- не дробить `auth_handler.go` на `json.go`/`cookies.go` сейчас: это не решает
  задачу interfaces/structs и дает лишний file churn.

#### Usecase input/output models

Файл: `internal/modules/auth/usecase/models.go`.

- `RegisterCommand`;
- `LoginCommand`;
- `LogoutCommand`;
- `ValidateSessionCommand`;
- `CurrentUser`;
- `AuthResult`.

Это application/usecase-level DTO/models. Они описывают input/output
application scenarios, а не HTTP JSON contract и не persistent/domain entity.

Зачем нужны:

- команды отделяют usecase API от HTTP DTO;
- `AuthResult` содержит `SessionID` и `ExpiresAt`, которые нужны delivery для
  cookie, но не возвращаются в JSON body;
- `CurrentUser` представляет application-level current user без domain methods
  и без HTTP tags.
- usecase не импортирует `delivery/transport`, чтобы сохранить dependency
  direction `delivery -> usecase -> domain`, даже если Yabloko использует более
  короткий путь через `transport.AuthResult`.

Дублирование:

- `CurrentUser` частично дублирует `domain.User` и `transport.UserResponse`;
- `RegisterCommand` и `LoginCommand` похожи на request DTO;
- это полезное дублирование, потому что domain entity и HTTP DTO не должны
  становиться application API usecase.

Что можно упростить:

- не объединять `CurrentUser` и `UserResponse`;
- можно переименовать `CurrentUser` в `UserResult`, но пользы мало;
- оставить `models.go` как единый файл для commands/results.

#### Domain / entity models

Файлы: `internal/modules/auth/domain/*.go`.

- `UserID`;
- `UserStatus`;
- `User`;
- `AuthIdentity`;
- `Session`;
- domain errors.

Зачем нужны:

- `User` и `Session` являются реальными domain/application entities для auth;
- `UserStatus` нужен для `IsActive`;
- domain errors используются для stable mapping в delivery;
- `ProviderPassword` используется repository для `auth_identities`.

Что спорно:

- `AuthIdentity` struct сейчас не используется в коде напрямую;
- `ProviderPassword` используется.

Рекомендация:

- оставить `AuthIdentity` на этом этапе, потому что она отражает schema
  `auth_identities` и может понадобиться при развитии auth identities;
- если команда хочет strict dead-code cleanup, удалить `AuthIdentity` можно
  отдельным низкорисковым шагом, но это не нужно для Yabloko-style refactor.

#### Database / repository models

Файлы:

- `internal/modules/auth/repository/auth_repo_postgres.go`;
- `internal/modules/auth/repository/session_repo_redis.go`.

PostgreSQL repository сейчас сканирует сразу в `domain.User`. Отдельных DB
models нет.

Redis repository имеет private model:

- `redisSessionPayload`.

Зачем нужен `redisSessionPayload`: Redis JSON shape не обязан совпадать с
`domain.Session`; private model защищает Redis contract от случайных изменений
domain struct.

Рекомендация:

- оставить `redisSessionPayload`;
- не добавлять Postgres DB models, пока SQL простой и scan идет в domain.User
  без nullable/derived fields;
- если появятся сложные joins/nullables, тогда добавить private repository
  models + mappers.

### 4. Repository layer split

Текущее разделение:

- `AuthRepoPostgres`: PostgreSQL users + password identities.
- `SessionRepoRedis`: Redis active sessions.

Это разделение стоит оставить.

Почему:

- Postgres и Redis имеют разные clients, failure modes и lifecycle;
- sessions по WorkNet contract не живут в SQL;
- Redis contract нельзя трогать в текущей задаче;
- `SessionRepoRedis` лучше не смешивать с `AuthRepoPostgres`, иначе auth repo
  станет storage god-object.

Нужно ли иметь два interfaces в usecase:

- да: `AuthRepository` и `SessionStore`;
- `AuthRepository` отвечает за durable user/password identity;
- `SessionStore` отвечает за ephemeral active sessions.

Стоит ли дробить `AuthRepository` на `UserRepository` и `IdentityRepository`:

- сейчас нет;
- `CreatePasswordUser` выполняет storage-atomic creation user + password
  identity в одной transaction;
- разделение на два interfaces может потребовать transaction manager или
  усложнит rollback semantics;
- если появятся OAuth/password reset/account linking, тогда можно вернуться к
  более тонким ports.

Как назвать в стиле Yabloko:

- concrete types уже названы Yabloko-like:
  `AuthRepoPostgres`, `SessionRepoRedis`;
- interface `SessionStore` можно оставить, потому что это behavior name для
  usecase, а не имя concrete repository;
- переименование `SessionStore` в `SessionRepository` даст мало пользы.

### 5. Auth middleware

Текущее создание:

- `internal/app/app.go` создает `authMiddleware := authmw.NewMiddleware(authUsecase, cookieConfig)`;
- private `registerAuthRoutes` в `internal/app/app.go` принимает готовый
  middleware;
- `delivery/middlewares.Middleware.RequireAuth(next)` читает cookie
  `session_id`, вызывает `ValidateSession`, кладет `CurrentUser` и `SessionID`
  в request context.

Почему middleware структура с методами:

- структура хранит `validator` и `cookies`;
- constructor удобен для DI в `internal/app`;
- tests могут создать middleware с fake usecase/session validator и cookie
  config;
- метод `RequireAuth` читается как reusable dependency.

Как в Yabloko:

- auth_service имеет набор function middlewares: `CSRFMiddleware`,
  `CSRFTokenMiddleware`, `CorsMiddleware`, `RateLimit`;
- JWT auth в handler flow, а не отдельный WorkNet-like session validator;
- route registration оборачивает handlers через `rateLimitHandler`.

ASAP ближе к WorkNet по session middleware:

- `AuthMiddleware(sessionService) func(http.Handler) http.Handler`;
- middleware function читает `session_id`, получает user id, пишет context.

Что лучше оставить в WorkNet:

- оставить middleware в `auth/delivery/middlewares`;
- оставить constructor-based DI;
- можно переименовать generic `Middleware` в `AuthMiddleware`;
- interface `SessionValidator` уже является правильным target/current naming,
  потому что middleware зависит только от `ValidateSession`.

Можно ли сделать проще:

- да, заменить struct на function
  `RequireAuth(validator SessionValidator, cookies CookieConfig) func(http.Handler) http.Handler`;
- это немного короче, но не сильно улучшает код;
- текущая структура не мешает тестируемости и понятна в composition root.

Рекомендация: не переписывать middleware shape радикально. Максимум — naming
cleanup (`SessionValidator`, `AuthMiddleware`, `auth_middleware.go`).

### 6. Запуск приложения

`cmd/api/main.go`:

- минимальный entrypoint;
- создает signal context;
- вызывает `app.New`;
- запускает server;
- выполняет graceful shutdown.

`internal/app/app.go`:

- composition root;
- загружает config;
- создает `pgxpool`;
- создает Redis client;
- создает auth repositories;
- создает auth usecase (`NewAuthUsecase` сейчас, `NewAuthUseCase` target);
- создает handler и middleware;
- регистрирует routes;
- владеет HTTP server и закрытием ресурсов.

Как это соотносится с Yabloko:

- Yabloko держит bootstrap в `auth_service/cmd/app.go`;
- там же создаются DB, repo, usecase, mux, middlewares и server;
- graceful shutdown не выражен так явно, как в WorkNet.

Рекомендация:

- оставить WorkNet схему;
- не переносить Yabloko bootstrap один в один;
- оставить auth route registration в composition root через private helper
  `registerAuthRoutes` внутри `internal/app/app.go` или, если `app.go` станет
  слишком большим, в `internal/app/routes.go`;
- handler при этом остается набором HTTP methods, JSON decode и cookie helpers;
- если `app.go` станет расти, можно выделить private helpers внутри `internal/app`:
  `newPostgresPool`, `newRedisClient`, `newAuthModule`, `newRouter`;
- это делать отдельным маленьким refactor step, не смешивая с naming/SQL/routing.

Риск упрощения запуска: средний. Любая ошибка в composition root может сломать
startup, resource cleanup или route registration. Проверка: `go test ./...`,
`go vet ./...`, `docker compose config`, manual smoke при необходимости.

### 7. Идея взять chi-router из ASAP

ASAP использует:

- dependency `github.com/go-chi/chi v1.5.5`;
- `chi.NewRouter()`;
- `mux.Use(...)`;
- `mux.Route("/api/v1/auth", func(mux chi.Router) { ... })`;
- `mux.With(authMiddleware, csrfMiddleware).Post(...)`;
- path params через `chi.URLParam` в handlers.

Плюсы `chi` для WorkNet:

- method-specific routes (`Post`, `Get`) без ручных method checks;
- route groups для `/api/v1/auth`;
- middleware chaining читается компактнее;
- path params пригодятся будущим modules.

Минусы сейчас:

- потребует менять `go.mod/go.sum`, что запрещено без отдельного решения;
- изменит routing layer во всем backend startup;
- текущие auth routes простые и не требуют path params;
- придется проверить `405` behavior и response body, чтобы не сломать public API;
- появится external dependency ради небольшой выгоды.

Какие файлы затронет:

- `go.mod`;
- `go.sum`;
- `internal/app/app.go`;
- `internal/modules/auth/delivery/http/auth_handler.go` и app route helper;
- route tests в `auth_handler_test.go`;
- возможно будущие handlers других modules.

Рекомендация:

- не внедрять `chi` в Step A naming или Step B embedded SQL;
- делать `chi` отдельным Step C: "routing migration to chi";
- перед Step C отдельно согласовать `go.mod/go.sum` diff и route-test scope;
- до Step C оставить текущий router wiring behavior.

## Interface and struct audit

### Interfaces

| Interface | Где объявлен | Кто реализует | Кто потребляет | Нужен ли | Рекомендация | Dependency direction |
|---|---|---|---|---|---|---|
| `authhttp.AuthUsecase` сейчас / `authhttp.AuthUseCaseInterface` target | `internal/modules/auth/delivery/http/auth_handler.go` | `*usecase.AuthUsecase` сейчас / `*usecase.authUseCase` target; test fake `fakeAuthUsecase` | `authhttp.AuthHandler` target | Да | В Step A переименовать в `AuthUseCaseInterface`; field `uc` оставить | Не нарушает: delivery зависит от usecase contract, interface объявлен у потребителя |
| `middlewares.SessionValidator` | `internal/modules/auth/delivery/middlewares/middlewares.go` | `*usecase.AuthUsecase` сейчас / `*usecase.authUseCase` target; test fake `fakeAuthUsecase` | `middlewares.Middleware` | Да | Оставить; имя точнее, потому что middleware нужен только `ValidateSession` | Не нарушает: middleware не знает concrete usecase |
| `usecase.AuthRepository` | `internal/modules/auth/usecase/interfaces.go` | `*repository.AuthRepoPostgres`; `fakeUsers` в tests | `usecase.AuthUsecase` сейчас / `usecase.authUseCase` target | Да | Оставить в `usecase/interfaces.go`; не дробить сейчас на `UserRepository`/`IdentityRepository` | Правильно: usecase объявляет storage port, repository его реализует |
| `usecase.SessionStore` | `internal/modules/auth/usecase/interfaces.go` | `*repository.SessionRepoRedis`; `fakeSessions` в tests | `usecase.AuthUsecase` сейчас / `usecase.authUseCase` target | Да | Оставить; имя лучше, чем `SessionRepository`, потому что Redis sessions являются ephemeral store | Правильно |
| `usecase.PasswordHasher` | `internal/modules/auth/usecase/interfaces.go` | `*password.BcryptHasher`; `fakeHasher` в tests | `usecase.AuthUsecase` сейчас / `usecase.authUseCase` target | Да | Оставить; concrete bcrypt не протекает в usecase | Правильно |
| `usecase.SessionTokenGenerator` | `internal/modules/auth/usecase/interfaces.go` | `*token.RandomGenerator`; `fakeTokens` в tests | `usecase.AuthUsecase` сейчас / `usecase.authUseCase` target | Да | Оставить; позволяет тестировать token generation без crypto randomness | Правильно |
| `usecase.Clock` | `internal/modules/auth/usecase/interfaces.go` | `RealClock`; `fakeClock` в tests | `usecase.AuthUsecase` сейчас / `usecase.authUseCase` target | Да | Оставить, но только как testability seam; не выносить за пределы usecase | Правильно |

#### Вывод по interfaces

- Лишних interfaces в production path не найдено.
- Два delivery-level interfaces нужны для тестируемости и narrow dependencies,
  но handler naming в Step A стоит довести до strict Yabloko-style:
  - `authhttp.AuthUsecase` -> `authhttp.AuthUseCaseInterface`;
  - `middlewares.SessionValidator` оставить.
- `usecase/interfaces.go` оправдан: auth usecase потребляет несколько ports
  конкретного auth module: Postgres auth repo, Redis session store, hasher,
  token generator, clock. Это не глобальное место для interfaces всего проекта.
- Заменять delivery interfaces на concrete usecase type не стоит:
  тесты handler/middleware станут тяжелее, а dependency direction станет менее
  выразительной.
- Заменять usecase interfaces на concrete repository/hasher/token types нельзя:
  это протащит infrastructure details в application layer.

### Structs

| Struct | Слой | Кто создает | Кто использует | Зачем нужна | Дублирование | Рекомендация |
|---|---|---|---|---|---|---|
| `authhttp.Handler` сейчас / `authhttp.AuthHandler` target | delivery/http | `internal/app` через `authhttp.NewHandler` сейчас / `authhttp.NewAuthHandler` target | Route handlers / tests | Хранит auth usecase dependency и cookie config для HTTP methods | Нет | В Step A переименовать в `AuthHandler`; field `uc` оставить |
| `usecase.AuthUsecase` сейчас / `usecase.authUseCase` target | usecase | `internal/app` через `NewAuthUsecase` сейчас / `NewAuthUseCase` target; tests через helper | delivery interfaces, tests | Содержит auth scenarios and application rules | Нет | В Step A переименовать в `authUseCase` + `NewAuthUseCase`, ближе к Yabloko |
| `repository.AuthRepoPostgres` | repository | `internal/app` | `usecase.AuthRepository` port | PostgreSQL users/password identities adapter | Нет | Оставить; inline SQL не трогать |
| `repository.SessionRepoRedis` | repository | `internal/app` | `usecase.SessionStore` port | Redis active sessions adapter | Нет | Оставить; Redis contract не менять |
| `middlewares.Middleware` | delivery/middlewares | `internal/app` через `authmw.NewMiddleware` | route wrapping | Хранит session validator and cookie config | Нет | Оставить или переименовать в `AuthMiddleware`; function-only rewrite не нужен |
| `delivery.CookieConfig` | delivery shared | `internal/app` from `config.Config` | handler and middleware | Delivery-level cookie policy subset | Частично дублирует config | Оставить: не надо передавать весь app config в delivery |
| `config.Config` | platform/config | `config.Load` | `internal/app` | Runtime config для app dependencies | Нет | Оставить; не менять |
| `app.App` | app/composition root | `app.New` | `cmd/api/main.go` | Владеет server/resource lifecycle | Нет | Оставить |
| `transport.RegisterRequest` | delivery/transport | JSON decoder | `Handler.Register` | Public request DTO | Не дублирует usecase полностью: есть JSON tags | Оставить |
| `transport.LoginRequest` | delivery/transport | JSON decoder | `Handler.Login` | Public request DTO | Не дублирует usecase полностью | Оставить |
| `transport.AuthResponse` | delivery/transport | mapper `ToAuthResponse` | HTTP responses | Public response DTO | Частично дублирует `usecase.AuthResult`, но без session fields | Оставить |
| `transport.UserResponse` | delivery/transport | mapper `ToUserResponse` | `AuthResponse`, `Me` | Public user response DTO | Частично дублирует `usecase.CurrentUser` | Оставить ради API boundary |
| `usecase.RegisterCommand` | usecase | `AuthHandler.Register` target | `authUseCase.Register` target | Application input | Похож на request DTO | Оставить ради boundary |
| `usecase.LoginCommand` | usecase | `AuthHandler.Login` target | `authUseCase.Login` target | Application input | Похож на request DTO | Оставить |
| `usecase.LogoutCommand` | usecase | `AuthHandler.Logout` target | `authUseCase.Logout` target | Application input | Минимальная wrapper struct | Оставить для единообразия commands |
| `usecase.ValidateSessionCommand` | usecase | middleware | `authUseCase.ValidateSession` target | Application input | Минимальная wrapper struct | Оставить |
| `usecase.CurrentUser` | usecase | `currentUserFromDomain` | middleware context, response mapper | Application output for authenticated user | Дублирует часть `domain.User` | Оставить: domain entity не должна становиться response DTO |
| `usecase.AuthResult` | usecase | `authUseCase` target | handler response/cookie logic | Application output with user + session metadata | Дублирует response частично | Оставить: `SessionID` не публичный JSON |
| `domain.User` | domain | repository scan | usecase | Core auth user entity | Частично пересекается с DTO/result | Оставить |
| `domain.AuthIdentity` | domain | Сейчас напрямую не создается | Не используется как struct; `ProviderPassword` используется | Schema readability for identities | Может быть dead code | Оставить пока; удаление обсудить отдельно |
| `domain.Session` | domain | usecase, repository mapper | usecase/session store | Core active session entity | Частично пересекается с Redis payload | Оставить |
| `repository.redisSessionPayload` | repository private | `SessionRepoRedis` | JSON marshal/unmarshal | Private Redis storage shape | Дублирует `domain.Session` намеренно | Оставить: защищает Redis contract |

#### Вывод по structs

- Основное дублирование намеренное и полезное: HTTP DTO, usecase commands/results
  и domain entities должны жить отдельно.
- `CookieConfig` полезен как delivery-level projection of app config.
- `AuthIdentity` — единственный спорный struct: он не используется напрямую, но
  отражает таблицу `auth_identities`. Удаление не нужно для текущего refactor.
- Handler не нужно дробить на `json.go`/`cookies.go`: helper functions относятся
  к HTTP behavior конкретного handler и пока не мешают читаемости.
- Route registration не является ответственностью handler methods и может быть
  перенесена в app/composition root без изменения route behavior.

## Recommended target structure

Цель: не менять business behavior, а сделать auth module чуть яснее по files и
naming, двигаясь ближе к Yabloko там, где это не ломает WorkNet contracts.

```txt
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
    sql/
      auth/
        create_password_user.sql
        find_user_by_email.sql
        find_user_by_id.sql
        password_hash_by_user_id.sql
        touch_password_identity.sql
    session_repo_redis.go
  delivery/
    errors.go
    http/
      auth_handler.go
      auth_handler_test.go
    middlewares/
      middlewares.go
      context.go
    transport/
      auth_json.go
internal/app/
  app.go
```

Целевые names:

- `authhttp.AuthHandler`;
- `authhttp.NewAuthHandler`;
- `authhttp.AuthUseCaseInterface`;
- `usecase.authUseCase`;
- `usecase.NewAuthUseCase`;
- `registerAuthRoutes` остается private helper в `internal/app` до отдельного
  routing migration step.

Future routing target после отдельного согласования:

- добавить `chi`;
- заменить `http.ServeMux` на `chi.Router` в composition root;
- сохранить exact public routes и текущий method/error/cookie behavior;
- не делать это в naming step.

`cmd/api/main.go` остается. `internal/app/app.go` остается composition root и
держит dependency wiring, server lifecycle и route registration.

Если `app.go` начнет разрастаться, допустимый следующий вариант —
`internal/app/routes.go` с unexported helper. На текущем этапе лучше начать с
private helper в `app.go`, чтобы не плодить файлы.

## Ответственность файлов target structure

- `domain/user.go`: `UserID`, `UserStatus`, `User`, `IsActive`.
- `domain/identity.go`: password provider и future auth identity model.
- `domain/session.go`: domain session entity.
- `domain/errors.go`: stable domain errors для usecase/repository/delivery mapping.
- `usecase/auth_usecase.go`: unexported implementation `authUseCase`,
  exported constructor `NewAuthUseCase`, сценарии `Register`, `Login`,
  `Logout`, `ValidateSession` и private validation/session helpers.
- `usecase/interfaces.go`: ports конкретного auth usecase/module, которые он
  потребляет:
  `AuthRepository`, `SessionStore`, `PasswordHasher`, `SessionTokenGenerator`,
  `Clock`.
- `usecase/models.go`: application/usecase-level commands/results/current user:
  `RegisterCommand`, `LoginCommand`, `LogoutCommand`, `ValidateSessionCommand`,
  `AuthResult`, `CurrentUser`.
- `repository/auth_repo_postgres.go`: PostgreSQL adapter for users/password
  identities; после embedded SQL step содержит Go adapter logic и `go:embed`
  bindings, но не inline query bodies.
- `repository/sql/auth/*.sql`: SQL queries for auth Postgres adapter, перенесенные
  без изменения смысла.
- `repository/session_repo_redis.go`: Redis adapter for active sessions.
- `delivery/errors.go`: mapping domain/usecase errors to WorkNet HTTP error
  contract.
- `delivery/http/auth_handler.go`: `AuthUseCaseInterface`, `AuthHandler`,
  `NewAuthHandler`, handler methods, strict JSON decode helper and session cookie
  helpers.
- `delivery/transport/auth_json.go`: request/response DTO and mapping.
- `delivery/middlewares/middlewares.go`: auth session validation middleware.
- `delivery/middlewares/context.go`: auth context helpers.
- `cmd/api/main.go`: entrypoint and graceful shutdown orchestration.
- `internal/app/app.go`: composition root, resource lifecycle and auth route
  registration helper.

## Что оставить как сейчас

- Public routes and behavior:
  `POST /api/v1/auth/register`, `POST /api/v1/auth/login`,
  `POST /api/v1/auth/logout`, `GET /api/v1/auth/me`.
- Cookie name/config/attributes: `session_id`, `HttpOnly`, `SameSite=Lax`,
  `Secure` from config, TTL from config.
- Redis session contract and key prefix `auth:session:`.
- PostgreSQL schema and migrations.
- `internal/app` as composition root.
- `cmd/api/main.go` as thin entrypoint.
- `usecase/interfaces.go` as auth usecase/module ports home.
- Separate `AuthRepository` and `SessionStore`.
- Separate usecase models and delivery DTO.
- No usecase import of `delivery/transport`.
- WorkNet error response format.
- No `chi` in naming or SQL steps.

## Что переименовать

Strict Yabloko naming cleanup:

- In `delivery/http/auth_handler.go`:
  - `Handler` -> `AuthHandler`;
  - `NewHandler` -> `NewAuthHandler`;
  - `AuthUsecase` -> `AuthUseCaseInterface`;
  - handler field stays `uc`.
- In `usecase/auth_usecase.go`:
  - `AuthUsecase` -> `authUseCase`;
  - `NewAuthUsecase` -> `NewAuthUseCase`;
  - private helper `newAuthUsecase` can become `newAuthUseCase` or be removed if
    constructor remains simple enough.
- In `delivery/middlewares/middlewares.go`:
  - keep `SessionValidator`;
  - optional later: `Middleware` -> `AuthMiddleware`.
- In `internal/app/app.go`:
  - update calls to `NewAuthUseCase` and `NewAuthHandler`;
  - local variable `authUsecase` can stay.
- In tests:
  - update concrete type references and constructors;
  - local variables `uc` stay.

## Что перенести

- Inline SQL bodies from `repository/auth_repo_postgres.go` to
  `repository/sql/auth/*.sql` in a separate embedded SQL step.
- Route registration from current stdlib helper to `chi` route grouping only in
  a separate routing migration step after dependency approval.

Не переносить:

- `decodeJSON`, `maxJSONBodyBytes`, `setSessionCookie`, `clearSessionCookie`
  из `auth_handler.go` сейчас;
- interfaces из `usecase/interfaces.go` в repository;
- usecase results в delivery/transport;
- repo/usecase construction в delivery router factory;
- naming, SQL and routing changes in one implement step.

## Что удалить

В первом безопасном refactor step ничего не удалять.

Отдельно можно обсудить позже:

- удалить unused `domain.AuthIdentity` struct, оставив `ProviderPassword`, если
  команда хочет strict dead-code cleanup. Сейчас лучше оставить ради schema
  readability.

## Что добавить

- In embedded SQL step:
  - `internal/modules/auth/repository/sql/auth/create_password_user.sql`;
  - `internal/modules/auth/repository/sql/auth/find_user_by_email.sql`;
  - `internal/modules/auth/repository/sql/auth/find_user_by_id.sql`;
  - `internal/modules/auth/repository/sql/auth/password_hash_by_user_id.sql`;
  - `internal/modules/auth/repository/sql/auth/touch_password_identity.sql`.
- In chi routing step:
  - dependency `github.com/go-chi/chi/v5`;
  - route tests around exact current public routes and middleware wrapping.

Тесты должны продолжать защищать:

- method checks;
- invalid JSON;
- cookie set/clear;
- missing cookie;
- route wiring;
- route protection for `logout` and `me`.

## Safe refactor plan

### Step A. Strict Yabloko naming only

Файлы:

- `internal/app/app.go`;
- `internal/modules/auth/delivery/http/auth_handler.go`;
- `internal/modules/auth/delivery/http/auth_handler_test.go`;
- `internal/modules/auth/delivery/middlewares/middlewares.go`;
- `internal/modules/auth/usecase/auth_usecase.go`;
- `internal/modules/auth/usecase/auth_usecase_test.go`.

Изменения:

- `authhttp.Handler` -> `authhttp.AuthHandler`.
- `authhttp.NewHandler` -> `authhttp.NewAuthHandler`.
- `authhttp.AuthUsecase` -> `authhttp.AuthUseCaseInterface`.
- Handler field оставить `uc`.
- `usecase.AuthUsecase` -> `usecase.authUseCase`.
- `usecase.NewAuthUsecase` -> `usecase.NewAuthUseCase`.
- Обновить app/tests/handlers/middleware compile references.
- Не менять `usecase/interfaces.go` placement.
- Не менять method signatures сценариев и бизнес-логику.
- Не менять routes, cookie behavior, JSON decode behavior, DB, Redis, Docker,
  `go.mod/go.sum`.

Риск: низкий. Это compile-time naming refactor.

Проверка:

- `gofmt`;
- `goimports`;
- `go test ./...`;
- `go vet ./...`;
- `docker compose config`.

Как проверить, что behavior не изменился:

- handler tests продолжают проверять register/login/logout/me;
- usecase tests продолжают проверять register/login/validate session;
- `git diff` не содержит изменений в SQL strings, cookie attrs, endpoint paths,
  JSON decode helper или error mapping.

### Step B. Embedded SQL only

Файлы:

- `internal/modules/auth/repository/auth_repo_postgres.go`;
- `internal/modules/auth/repository/sql/auth/create_password_user.sql`;
- `internal/modules/auth/repository/sql/auth/find_user_by_email.sql`;
- `internal/modules/auth/repository/sql/auth/find_user_by_id.sql`;
- `internal/modules/auth/repository/sql/auth/password_hash_by_user_id.sql`;
- `internal/modules/auth/repository/sql/auth/touch_password_identity.sql`.

Изменения:

- Добавить embedded SQL files по стилю Yabloko.
- Использовать `//go:embed` для SQL strings.
- Не менять запросы по смыслу:
  - transaction в `CreatePasswordUser` сохраняется;
  - insert into `users` сохраняется;
  - insert into `auth_identities` with `provider='password'` сохраняется;
  - selects и update `last_used_at` сохраняются.
- Не менять DB schema/migrations.
- Не менять Redis/session repository.
- Не менять public API.

Риск: средний-низкий. Основной риск — случайно изменить whitespace/placeholder
или порядок SQL statements так, что Postgres behavior поменяется.

Проверка:

- `gofmt`;
- `go test ./...`;
- `go vet ./...`;
- `docker compose config`;
- optional manual compare of old inline SQL vs embedded SQL content;
- при наличии test DB позже добавить/запустить repository integration smoke.

Как проверить, что behavior не изменился:

- `CreatePasswordUser` still creates user and password identity in one
  transaction;
- `FindByEmail`, `FindByID`, `PasswordHashByUserID`,
  `TouchPasswordIdentity` return same domain/errors;
- no migration files changed.

### Step C. Routing migration to chi

Файлы:

- `go.mod`;
- `go.sum`;
- `internal/app/app.go`;
- likely new or updated route tests, например `internal/app/app_test.go` или
  focused auth route wiring tests.

Минимальный diff для сохранения public API behavior:

- Добавить dependency `github.com/go-chi/chi/v5`.
- В `internal/app/app.go` создать `chi.NewRouter()` вместо `http.NewServeMux`.
- Оставить `health` behavior через handler method check.
- Сохранить exact auth paths:
  - `POST /api/v1/auth/register`;
  - `POST /api/v1/auth/login`;
  - `POST /api/v1/auth/logout`;
  - `GET /api/v1/auth/me`.
- Чтобы сохранить текущий `405`/error behavior handler methods, route
  registration должна направлять все methods на handler methods, а не заменять
  их chi method-specific handlers без проверки response contract.
- Сохранить middleware wrapping for `logout` and `me`, например через
  `middleware.RequireAuth(http.HandlerFunc(handler.Logout))` или эквивалентный
  `chi.With(middleware.RequireAuth)` после проверки tests.
- Не менять handler methods, cookie behavior, JSON decode behavior, DB, Redis.

Риск: средний. `chi` может изменить behavior вокруг method mismatch, trailing
slash, 404/405 и middleware ordering, если использовать method-specific routes
без тестов.

Проверка:

- route tests for exact current public routes;
- tests for wrong method preserving current error/status behavior;
- tests that `logout` and `me` remain protected;
- `go test ./...`;
- `go vet ./...`;
- `docker compose config`;
- quick HTTP smoke for register/login/logout/me after separate approval.

Как проверить, что behavior не изменился:

- response codes and error shapes for current auth tests stay the same;
- cookie attrs stay the same;
- Redis session validation path stays the same;
- `go.mod/go.sum` diff contains only chi dependency changes required by the
  routing step.

## Вопросы для Павла, Георгия и Сергея

1. Подтвердить, что Step A использует именно `authUseCase` + `NewAuthUseCase`.
   - Почему требует решения: `authUseCase` unexported и ближе к Yabloko, но
     exported `AuthUseCase` был бы проще искать из других packages.
   - Рекомендованный вариант: `authUseCase` + `NewAuthUseCase`.
   - Последствия: меньше наружной API surface внутри `internal`, но type нельзя
     будет назвать из других packages напрямую.

2. Подтвердить, что embedded SQL files в Step B называются по operation names,
   а не повторяют дословно Yabloko names.
   - Почему требует решения: WorkNet repo methods отличаются от Yabloko
     (`CreatePasswordUser`, `PasswordHashByUserID`, `TouchPasswordIdentity`).
   - Рекомендованный вариант: names match WorkNet operations.
   - Последствия: легче сопоставлять SQL file и repository method.

3. Перед Step C отдельно подтвердить `chi` dependency и exact route-test scope.
   - Почему требует решения: `go.mod/go.sum` меняются, а router может повлиять
     на 404/405/trailing slash behavior.
   - Рекомендованный вариант: делать Step C только после Step A и Step B.
   - Последствия: routing risk изолирован от naming и repository cleanup.

## Итоговая рекомендация

Сейчас не стоит делать крупный смешанный рефактор. WorkNet уже находится близко
к Yabloko-style по auth module structure, но пользователь принял более strict
naming и embedded SQL как будущие направления.

Рекомендованная последовательность:

1. Step A: strict Yabloko naming only.
2. Step B: embedded SQL only.
3. Step C: routing migration to `chi` only after separate dependency approval.

`auth_handler.go` не дробить на `json.go` и `cookies.go`. Не переносить usecase
models в `delivery/transport`. Не переносить repository/usecase ports из
`usecase/interfaces.go`, потому что это ports конкретного auth usecase/module,
а не общая dump-zone для interfaces всего проекта.
