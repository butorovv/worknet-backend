# Yabloko Backend Style Guide For WorkNet

Дата: 2026-06-04

## Цель

Этот документ фиксирует правила, по которым WorkNet backend приводится к стилю
`2025_2_Yabloko` branch `main`.

Главный reference для структуры, naming и Go-кодстайла: `2025_2_Yabloko`.

`2026_1_ASAP` разрешено использовать только как подтверждение уже выбранной
auth/session механики WorkNet:

- opaque cookie session;
- Redis-backed session store;
- session middleware semantics.

JWT-подход и общий стиль `2026_1_ASAP` не переносится.

## Базовое ограничение WorkNet

`docs/architecture.md` WorkNet имеет приоритет, если Yabloko-паттерн ломает
dependency direction или composition root.

Конфликты, которые нельзя переносить напрямую:

- в Yabloko часть router-фабрик создает repository/usecase внутри delivery;
  в WorkNet зависимости создаются в `internal/app`;
- в Yabloko auth usecase импортирует `delivery/transport`;
  в WorkNet usecase не должен зависеть от HTTP/transport DTO;
- Yabloko auth использует JWT;
  WorkNet сохраняет cookie-based Redis sessions;
- Yabloko не везде сохраняет graceful shutdown;
  WorkNet сохраняет текущий graceful shutdown;
- Yabloko tests местами расходятся с актуальным кодом;
  переносится стиль, а не слепые test fixtures.

## 1. Итоговая структура папок

WorkNet остается modular monolith, но внутри модулей используется Yabloko-like
layout:

```txt
cmd/api/
internal/
  app/
  modules/
    auth/
      domain/
      usecase/
      repository/
        sql/
          auth/
      delivery/
        http/
        transport/
        middlewares/
  platform/
```

Правила:

- `internal/app` остается composition root.
- `internal/modules/<module>/delivery/http` содержит handlers и route
  registration.
- `internal/modules/<module>/delivery/transport` содержит HTTP request/response
  structs и mapping к usecase/domain результатам.
- `internal/modules/<module>/delivery/middlewares` содержит module-local HTTP
  middleware и context helpers.
- `internal/modules/<module>/usecase` содержит сценарии, input/result models и
  интерфейсы repository/session/password/token.
- `internal/modules/<module>/repository` содержит storage adapters.
- SQL можно вынести в `repository/sql/<area>/*.sql` по Yabloko-образцу, если это
  не увеличивает риск.

## 2. Naming для packages

- Package names короткие и соответствуют папке: `domain`, `usecase`,
  `repository`, `transport`, `middlewares`.
- Package `delivery/http` может называться `http`, как в Yabloko, но в местах
  импорта его нужно алиасить, например `authhttp`, чтобы не путать с `net/http`.
- Platform packages остаются техническими: `config`, `httpx`, `password`,
  `token`.
- Не добавлять `ports`, `adapters`, `services` как новые папки без явной пользы.

## 3. Naming для файлов

Yabloko-паттерн:

- delivery handler: `<module>_handler.go`;
- delivery transport DTO: `<module>_json.go`;
- usecase: `<module>_usecase.go`;
- repository Postgres adapter: `<module>_repo_postgres.go`;
- shared repository DB helpers: `db.go`, если появятся;
- middleware: `middlewares.go` или короткий файл по назначению;
- tests рядом с кодом: `<file>_test.go`.

Для Auth целевой naming:

```txt
internal/modules/auth/delivery/http/auth_handler.go
internal/modules/auth/delivery/http/auth_handler_test.go
internal/modules/auth/delivery/transport/auth_json.go
internal/modules/auth/delivery/middlewares/middlewares.go
internal/modules/auth/usecase/auth_usecase.go
internal/modules/auth/usecase/interfaces.go
internal/modules/auth/repository/auth_repo_postgres.go
internal/modules/auth/repository/session_repo_redis.go
```

Если rename дает большой diff без смысловой пользы, его можно делать отдельным
маленьким этапом.

## 4. Domain / Entities

Правила:

- Domain содержит сущности, value objects и domain errors.
- Domain не импортирует delivery, usecase, repository, Postgres, Redis,
  platform/http helpers.
- Struct naming простой: `User`, `Identity`, `Session`.
- Errors лежат в `domain/errors.go`, объявляются как package-level `var` через
  `errors.New`.
- Domain errors должны описывать бизнес-смысл, а не HTTP status.

Не переносим:

- Yabloko JWT claims как domain/transport модель.

## 5. Usecase / Service

Yabloko-паттерн:

- основной тип сценариев называется `<Module>Usecase` или `<module>UseCase`;
- constructor называется `New<Module>Usecase` / `New<Module>UseCase`;
- методы usecase принимают `context.Context`;
- бизнес-валидация и orchestration находятся в usecase;
- handlers остаются тонкими.

WorkNet-правила:

- usecase не импортирует `delivery/transport`;
- usecase принимает command/input structs или простые параметры;
- usecase возвращает usecase result structs/domain entities;
- hashing, token generation, session store подключаются через интерфейсы;
- интерфейсы для repository/session/password/token лежат в `usecase`, потому что
  usecase их использует.

Для Auth:

- сохранить register/login/logout/me semantics;
- сохранить Redis session lifecycle;
- сохранить validation и error mapping;
- не добавлять JWT/refresh token.

## 6. Repository

Yabloko-паттерн:

- concrete repository type: `<Module>RepoPostgres`;
- constructor: `New<Module>RepoPostgres(db)`;
- методы принимают `context.Context`;
- repository возвращает domain entities и domain errors;
- repository не содержит бизнес-валидацию;
- SQL может быть вынесен в embedded `.sql` files.

WorkNet-правила:

- Postgres adapter работает с `pgxpool.Pool`;
- Redis session adapter остается отдельным concrete repository/store;
- repository не импортирует usecase или delivery;
- repository мапит storage errors в domain/usecase-level errors только на
  своей границе.

## 7. Delivery / HTTP

Yabloko-паттерн:

- handler struct содержит usecase interface и response sender/helper;
- constructor `New<Module>Handler(uc)`;
- route registration рядом с handler, например `New<Module>Router` или
  `RegisterRoutes`;
- handler methods проверяют HTTP method;
- decode JSON делается через `json.Decoder`;
- `DisallowUnknownFields` используется для strict request body там, где contract
  уже ожидает строгий JSON;
- errors мапятся в handler-level helper;
- successful response отправляется через общий response helper.

WorkNet-правила:

- public API behavior не менять без необходимости;
- status codes, cookie attributes, response JSON contract сохраняются;
- route registration не создает repository/usecase;
- handler мапит HTTP DTO в usecase commands;
- handler не содержит storage или business orchestration;
- cookie helpers остаются рядом с auth delivery.

## 8. Transport / DTO

Yabloko-паттерн:

- HTTP request/response DTO лежат в `delivery/transport`;
- файл называется `<module>_json.go`;
- DTO names exportable, если используются из handler package:
  `RegisterRequest`, `LoginRequest`, `AuthResponse`, `UserResponse`;
- mapping functions называются `To...Response`.

WorkNet-правила:

- transport DTO не импортируются в usecase;
- transport DTO могут импортировать domain/usecase для response mapping;
- request DTO не должны становиться business command structs.

## 9. Middleware

Yabloko-паттерн:

- middleware лежит в `delivery/middlewares`;
- context helpers лежат рядом с middleware;
- middleware возвращает `func(http.Handler) http.Handler` или equivalent
  handler wrapper.

WorkNet-правила:

- session middleware использует auth usecase/session validation;
- missing/invalid session behavior сохраняется;
- user/session context keys не экспортируют mutable details.

## 10. Errors

Yabloko-паттерн:

- domain errors централизованы в `domain/errors.go`;
- handler сравнивает errors через `errors.Is`;
- HTTP status мапится в delivery.

WorkNet-правила:

- сохранить текущий public error response body;
- не менять business errors ради формального совпадения;
- unknown/internal errors не должны раскрывать infrastructure details.

## 11. Response helpers

Yabloko-паттерн:

- общий response sender/helper;
- success and error responses идут через один слой.

WorkNet-правила:

- текущий `internal/platform/httpx` можно оставить, потому что он уже
  закрепляет public API contract;
- не переносить Yabloko `{ "error": "..." }`, если это меняет WorkNet response
  contract;
- можно привести call sites к более единообразному стилю без смены JSON shape.

## 12. Constructors

Yabloko-паттерн:

- `New<Module>Handler(uc)`;
- `New<Module>Usecase(repo, ...)`;
- `New<Module>RepoPostgres(db)`;
- constructors возвращают pointer на concrete type.

WorkNet-правила:

- constructors не должны скрыто создавать внешние clients;
- Redis/Postgres clients создаются в `internal/app`;
- constructor arguments идут от бизнес-зависимостей к technical dependencies.

## 13. Interfaces

Yabloko-паттерн:

- интерфейс лежит в пакете, который его использует;
- delivery объявляет usecase interface;
- usecase объявляет repository interface.

WorkNet-правила:

- keep interfaces small;
- repository concrete types не объявляют интерфейсы для самих себя;
- не добавлять `ports` папку ради одного интерфейса;
- test fakes используют эти локальные interfaces.

## 14. Tests

Yabloko-паттерн:

- tests лежат рядом с кодом;
- handler tests проверяют HTTP status/body/header/cookie;
- usecase tests используют fake/mock repository and dependencies;
- repository tests возможны отдельно, если есть isolated DB/test harness.

WorkNet-правила:

- не добавлять новые зависимости для тестов;
- для auth delivery покрывать:
  - register/login/logout/me;
  - invalid method;
  - invalid JSON;
  - unknown JSON field;
  - trailing token;
  - cookie set/clear behavior.
- usecase tests должны сохранять business semantics.

## 15. Что не переносим из Yabloko

Не переносим:

- JWT auth и refresh-token API;
- direct dependency from usecase to delivery/transport;
- создание repositories/usecases inside delivery router;
- multi-service runtime layout как обязательную архитектуру;
- изменение error JSON contract WorkNet;
- изменение cookie names/attributes/session TTL semantics;
- weak config parsing patterns;
- неактуальные test fixtures, которые не соответствуют Yabloko production code.

## Практическое правило миграции

Каждый шаг должен быть маленьким:

1. Сначала переносить naming/layout, который не меняет behavior.
2. Затем двигать DTO/middleware по слоям.
3. Затем переименовывать concrete types/constructors.
4. Затем приводить repository SQL style, если тесты стабильно проходят.
5. После каждого логического шага запускать форматирование и tests.
