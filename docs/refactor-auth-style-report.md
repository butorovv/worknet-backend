# Auth Style Refactor Report

- Author: Сергей
- Mode: backend
- Date: 2026-06-04
- Status: completed

## Изученные проекты

- `2026_1_ASAP`, branch `dev`.
  - Единый backend: `handlers/services/repository/domain/dto/middleware/utils`.
  - Хорошие практики: интерфейсы объявлены у потребителя, Redis-backed sessions, отдельные DTO, repository models/mappers, простые constructors.
  - Минусы для WorkNet MVP: handlers местами толстые, response format отличается, composition root крупный.
- `2025_2_Yabloko`, branch `main`.
  - Сервисная структура: `delivery/usecase/repository/domain`.
  - Хорошие практики: понятное разделение слоев, прямые constructors, минимальные repository abstractions, stdlib `http.ServeMux`.
  - Минусы для WorkNet auth: auth завязан на JWT, а не opaque Redis sessions; часть DTO находится слишком близко к usecase.

## Style Reference

Основной structural reference: `2025_2_Yabloko`, потому что его `delivery/usecase/repository/domain` ближе к текущей MVP-структуре `worknet-backend` и проще поддерживается без enterprise-overengineering.

Из `2026_1_ASAP` взяты практики для auth/session: cookie-based session style, Redis session semantics, интерфейсы у потребителя, простые service/repository constructors и pragmatic tests.

## Примененные правила

- Сохраняем `internal/modules/auth/{domain,usecase,repository,delivery/http}`.
- Не добавляем `ports/adapters` folders.
- Интерфейсы остаются там, где потребляются: delivery зависит от `AuthService`, usecase зависит от repository/session/hash/token interfaces.
- Handlers остаются тонкими: HTTP DTO -> usecase command -> HTTP response.
- Business logic остается в usecase.
- Repository не импортирует usecase и занимается только storage.
- Platform helpers не расширяются без необходимости.
- API error format остается `{ "error": { "code": "...", "message": "..." } }`.

## Измененные файлы

- `internal/app/app.go`
- `internal/modules/auth/usecase/interfaces.go`
- `internal/modules/auth/usecase/ports.go`
- `internal/modules/auth/usecase/service.go`
- `internal/modules/auth/usecase/commands.go`
- `internal/modules/auth/usecase/service_test.go`
- `internal/modules/auth/repository/postgres.go`
- `internal/modules/auth/repository/redis_sessions.go`
- `internal/modules/auth/delivery/http/handler_test.go`
- `docs/refactor-auth-style-report.md`

## Что упрощено

- Убран `Dependencies` struct из `auth/usecase`.
- `UserRepository` и `IdentityRepository` объединены в один MVP-level `AuthRepository`.
- `ports.go` заменен на `interfaces.go`, чтобы naming был ближе к старым проектам.
- Убран `CreatePasswordUserParams`, repository method теперь принимает простые аргументы.
- `PostgresRepository` больше не импортирует package `usecase`.
- `RedisSessionStore.Delete` теперь возвращает `domain.ErrSessionNotFound`, если Redis key не удален.
- Добавлен route smoke test для `RegisterRoutes`.

## Что оставлено без изменений

- Public auth API: `register`, `login`, `logout`, `me`.
- Cookie-based sessions и cookie name/config.
- PostgreSQL tables `users` и `auth_identities`.
- Redis-backed active sessions.
- Existing migrations.
- Graceful shutdown.
- Dockerfile и `docker-compose.yaml`.
- Response DTO и error format.
- `internal/platform` helpers.

## Проверки

- `gofmt -w ...` - passed.
- `goimports -w ...` - passed.
- `go test ./...` - passed.
- `go vet ./...` - passed.
- `docker compose config` - passed.

`docker compose up` не запускался: для этой задачи достаточно было проверить, что compose-конфиг синтаксически собирается и не сломан изменениями.
