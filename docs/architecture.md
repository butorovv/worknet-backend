# Backend Architecture

## Назначение

Этот документ описывает целевую backend-архитектуру WorkNet. Он не утверждает,
что вся архитектура уже реализована в коде: текущий backend содержит scaffold
под будущую реализацию, а этот документ задает направление для разработки.

`../architecture-review.md` использован только как источник практик старого Go
проекта. Старая сервисная архитектура не является reference для WorkNet и не
переносится как самостоятельная модель.

## Источники

- `../architecture-review.md` — источник практик и анти-практик.
- `../worknet-backend` — главный backend repo.
- Пользовательская задача от 2026-06-04 — источник целевых архитектурных
  решений для этого документа.

## Целевой стиль

WorkNet backend строится как modular monolith с lightweight clean/hexagonal
границами.

Целевые свойства:

- один backend process;
- бизнес-границы выражены через модули в `internal/modules`;
- `internal/app` является composition root;
- `internal/platform` содержит technical helpers;
- domain не зависит от HTTP, Postgres, Redis, S3 и других infrastructure
  details;
- handlers работают с HTTP DTO;
- usecase работает с input/command структурами, а не с HTTP DTO;
- storage, cache и внешние интеграции подключаются через ports/adapters;
- транзакции задаются на уровне usecase, если сценарий требует
  бизнес-атомарности.

## Текущий scaffold

На момент подготовки документа backend содержит scaffold:

```txt
cmd/
configs/
db/
deployments/
docs/
internal/
  app/
  modules/
  platform/
tests/
```

В `internal/modules` уже есть директории модулей:

```txt
activity/
artifacts/
auth/
contacts/
contributions/
integrations/
media/
profiles/
requests/
responses/
resumes/
spaces/
```

Наличие директории не означает, что API, DTO, database schema или usecase уже
подтверждены реализацией. До появления кода модульные обязанности фиксируются
как целевые.

## Dependency Direction

Целевая зависимость направлена внутрь:

```txt
HTTP handler
  -> usecase input/command
  -> usecase
  -> domain
  -> ports

repository / cache / media storage adapters
  -> implement ports
```

Правила:

- `domain` не импортирует `delivery`, `repository`, `platform/httpx`,
  Postgres, Redis или S3 packages.
- `usecase` не импортирует HTTP DTO.
- `delivery/http` мапит HTTP DTO в usecase input/command.
- `repository` и technical adapters не протекают в business API модуля.
- `internal/app` связывает конкретные implementations.

## Composition Root: `internal/app`

`internal/app` отвечает за wiring приложения:

- загрузка config;
- создание `pgxpool`;
- создание Redis client;
- создание HTTP server и router;
- сборка middleware chain;
- создание module repositories/adapters;
- создание usecases;
- регистрация handlers/routes;
- graceful shutdown и закрытие ресурсов.

Репозитории и usecases не создаются внутри router-фабрик. Delivery слой
получает готовые зависимости.

## Modules: `internal/modules`

`internal/modules` содержит бизнес-модули и module-local adapters. Базовая
структура модуля:

```txt
internal/modules/<module>/
  domain/
  usecase/
  repository/
  delivery/
    http/
```

Целевые обязанности:

- `domain` — entities, value objects, domain errors и domain rules без
  infrastructure details.
- `usecase` — application scenarios, input/command structures, ports,
  transaction boundaries.
- `repository` — Postgres adapters и storage mapping.
- `delivery/http` — route registration, handlers, HTTP DTO, HTTP error mapping,
  mapping HTTP DTO <-> usecase input/output.

Модуль не должен напрямую менять данные другого модуля через его repository.
Если нужен cross-module сценарий, он оформляется через usecase-level port или
application service, а не через обход границ.

## Platform: `internal/platform`

`internal/platform` содержит только технические helpers, которые не являются
business logic:

- logging и request id;
- HTTP response helpers;
- validation helpers;
- password hashing;
- config helpers;
- shared middleware primitives;
- wrappers для infrastructure concerns.

`internal/platform` не должен становиться свалкой business helpers. Если helper
знает о бизнес-смысле профиля, сессии, пространства или заявки, он принадлежит
соответствующему module.

## Auth And Sessions

Целевая auth-модель WorkNet:

- cookie-based sessions;
- JWT не используется;
- session lifecycle находится внутри `auth`;
- session store реализуется через adapter, например Redis-backed repository;
- Redis client создается в `internal/app`;
- session adapter передается в `auth` через port;
- handlers работают только с cookie и HTTP DTO;
- usecase работает с commands, session identifiers и domain/application
  структурами.

`auth` отвечает за:

- регистрацию и вход, если эти сценарии подтверждены API;
- создание session;
- валидацию session;
- logout/session revocation;
- выдачу текущего пользователя на уровне application boundary.

Конкретные endpoint'ы, DTO и error format должны фиксироваться отдельно после
подтверждения API contract.

## Profiles, Avatars And Media

`profiles` владеет бизнес-логикой профиля и аватарки:

- правила, кто может менять avatar;
- валидация бизнес-ограничений avatar;
- связь avatar с profile;
- сценарии обновления profile data.

`media` является technical module:

- хранение media object;
- выдача media object;
- metadata, storage keys, size/content-type checks как technical concerns;
- adapter к локальному storage, S3-compatible storage или другому backend.

`profiles` не должен знать детали S3, filesystem или storage client. Он
работает через media port/usecase и сохраняет только бизнес-связь avatar с
profile.

## Data Access

Postgres используется через `pgxpool`.

Целевые правила:

- `pgxpool` создается в `internal/app`;
- repository получает pool или transaction handle через constructor/adapter;
- ORM не является целевой зависимостью по умолчанию;
- SQL может храниться рядом с repository, если это удобно для читаемости и
  тестирования;
- database schema и migrations фиксируются отдельно в `db`.

Domain не знает о таблицах, SQL-запросах, `pgx`, `pgxpool` или Redis keys.

## Transactions

Транзакции задаются на уровне usecase, когда сценарий должен быть атомарным с
точки зрения бизнеса.

Целевой подход:

- usecase определяет границу транзакции;
- transaction manager или unit-of-work port может быть объявлен на usecase
  boundary;
- repository выполняет операции внутри переданного transaction context/handle;
- repository не принимает самостоятельное решение о бизнес-атомарности;
- rollback/commit не должны быть спрятаны в handler.

Если операция является простой single-statement storage operation, отдельная
usecase transaction boundary может не требоваться.

## DTO And Input Policy

Handlers работают с HTTP DTO:

- request body;
- query/path parameters;
- cookies;
- HTTP response DTO;
- HTTP-specific validation и status mapping.

Usecase работает с input/command структурами:

- `CreateSessionCommand`;
- `UpdateProfileCommand`;
- `UploadAvatarCommand`;
- `CreateSpaceCommand`;
- другие application-level inputs.

Правила:

- HTTP DTO не передаются в usecase напрямую.
- Usecase output не обязан совпадать с HTTP response DTO.
- Mapping HTTP DTO -> command находится в delivery/application boundary.
- Domain structures не используются как request body по умолчанию.

## Error Handling

Целевой подход:

- domain/usecase возвращают domain или application errors;
- delivery/http мапит errors в HTTP status и response DTO;
- единый HTTP error format должен быть зафиксирован отдельным API contract;
- middleware и handlers должны возвращать совместимый error format;
- internal errors не раскрываются клиенту.

Пока API error format не подтвержден, этот документ фиксирует только boundary:
ошибка не должна заставлять domain знать о HTTP.

## Practices From The Old Project

Переносим как практики:

- понятное направление слоев `handler -> usecase -> repository`;
- ручной dependency injection;
- interfaces/ports на стороне usecase;
- HTTP DTO отдельно от domain;
- `pgx` / `pgxpool` для Postgres;
- SQL-first repository style, если он подтверждается конкретной реализацией;
- request id и logger в context;
- единые HTTP response helpers;
- explicit transactions для бизнес-атомарных сценариев;
- тестируемость usecase через mocks.

## Practices We Do Not Carry Over

Не переносим:

- service-oriented monorepo с несколькими отдельными HTTP services;
- дублирование middleware и module logic между services;
- JWT auth и shared-secret local JWT validation;
- session logic вне `auth`;
- dependency wiring внутри delivery/router factories;
- usecase dependency on HTTP transport DTO;
- external API DTO внутри domain;
- domain dependency on HTTP, Postgres, Redis, S3 or provider clients;
- decoding HTTP body directly into domain structures;
- repository-owned transaction boundary для бизнес-сценариев;
- старую архитектуру как standalone WorkNet reference.

## Documentation Rules

Этот документ должен обновляться, когда:

- меняется целевая module boundary;
- меняется composition root;
- меняется auth/session model;
- меняется transaction policy;
- появляется подтвержденный API/database contract, который влияет на backend
  architecture.

При изменении кода source of truth остается:

```txt
1. actual code
2. git diff
3. docs/architecture.md
4. Brain wiki
5. user instructions
6. assumptions
```
