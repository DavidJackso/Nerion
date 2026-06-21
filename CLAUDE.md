# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
make run       # go run ./cmd
make build     # binary → ./bin/server
make migrate   # apply SQL migrations via goose
make test      # go test ./...
make lint      # go vet ./...
make tidy      # go mod tidy
make rename MODULE=github.com/you/myapp  # sed-replace module path everywhere
```

Run a single test:
```bash
go test ./internal/service/... -run TestName
```

## Config

Requires `config.yaml` (copy from `config.yaml.example`) or env vars:

| Env var | Required |
|---|---|
| `APP_DB_DSN` | yes |
| `APP_JWT_SECRET` | yes |
| `APP_JWT_TTL` | no (default `24h`) |
| `APP_HTTP_ADDR` | no (default `:8080`) |
| `APP_LOG_LEVEL` | no (default `info`) |
| `APP_LOG_FORMAT` | no (default `json`) |
| `APP_STORAGE_S3_BUCKET` | no — if set, S3 is used; otherwise local |
| `APP_STORAGE_S3_ENDPOINT` | no |
| `APP_STORAGE_S3_ACCESS_KEY` | no |
| `APP_STORAGE_S3_SECRET_KEY` | no |
| `APP_STORAGE_S3_REGION` | no (default `ru-central1`) |
| `APP_STORAGE_UPLOAD_DIR` | no (default `./uploads`, local mode only) |
| `APP_STORAGE_PRESIGN_TTL` | no (default `1h`) |

Storage selection: if `APP_STORAGE_S3_BUCKET` is non-empty, `S3Adapter` (minio) is used; otherwise `LocalAdapter` writes to `upload_dir` and returns `/files/<key>` URLs (dev only).

Env vars override `config.yaml`.

## Architecture

Clean layered architecture — dependency direction: transport → service → repository → DB. Interfaces in `internal/domain/` decouple layers.

```
cmd/main.go              — entry: loads config+logger, creates App, runs until signal
cmd/migrate/main.go      — standalone migration runner (goose embedded SQL)
internal/app/main.go     — wires pool → repo → service → HTTP server, graceful shutdown
internal/config/         — viper: file + env loading
internal/domain/         — interfaces only (repositories + services per domain)
internal/entity/         — plain structs (User, Session, Space, SpaceMember, TableMeta, FieldMeta, APIKey, List, PDFTemplate, PDFJob, AuditEntry)
internal/repository/     — pgx/v5 pool queries implementing domain interfaces
internal/service/        — business logic implementing domain service interfaces
internal/transport/http/ — chi router, handlers, request/response contracts
internal/middleware/      — Auth (JWT Bearer) + RequireRole + LoadSpace + RequireSpaceRole + APIKeyAuth + SecurityHeaders + request logger
internal/jwtauth/        — HS256 sign/parse (Claims: sub, role, exp)
internal/adapter/email/  — stub EmailSender (swap with real SMTP impl)
internal/adapter/storage/ — LocalAdapter + S3Adapter (minio); selected at startup by config
migrations/              — goose SQL files, embedded via embed.go
pkg/apierrors/           — typed APIError sentinel vars (ErrNotFound, ErrForbidden, …)
```

**Error handling:** services return `*apierrors.APIError` sentinels; `server.writeError` unwraps them with `errors.As` and maps to HTTP status. Unexpected errors log via slog and return 500.

## Auth flow

Two-token auth: short-lived JWT access token + opaque refresh token (stored as SHA-256 hash in `sessions` table).

- `POST /auth/register` → creates user, sends email verification link (24h expiry)
- `POST /auth/login` → verifies bcrypt password, issues JWT + refresh token (30-day session)
- `POST /auth/refresh` → rotates refresh token (revoke old, create new session), issues new JWT
- `POST /auth/logout` → revokes refresh token session
- `POST /auth/verify-email` → marks user `email_verified=true`
- `POST /auth/password/reset-request` / `POST /auth/password/reset` → 60-min token, invalidates all sessions on success
- `GET /auth/me` → returns current user from JWT claims

Protected routes: `middleware.Auth(jm)` parses Bearer token → `*jwtauth.Claims` in context. Retrieve with `middleware.ClaimsFrom(ctx)`. `middleware.RequireRole("admin")` gates admin-only routes.

Tokens stored as raw hex strings; only SHA-256 hashes persisted in DB.

## Spaces

Multi-tenant workspace model. Each space has an owner and members with roles (`admin` | `member`).

- `LoadSpace` middleware resolves `{slug}` URL param → `*entity.Space` in context (retrieve with `middleware.SpaceFrom(ctx)`)
- `RequireSpaceRole(minRole, memberRepo)` checks caller's membership — use after `Auth` + `LoadSpace`
- Space creation auto-creates a Postgres schema (`CREATE SCHEMA`) with the space slug; deletion drops it
- `SpaceService.Delete` requires caller to pass `confirmName` matching space name (guard against accidents)

## Schema (Tables & Fields)

Tables and fields are metadata describing the structure of data within a space. Each table has a slug unique within its space; each field has a slug unique within its table.

Field types: `text`, `longtext`, `number`, `date`, `datetime`, `boolean`, `enum`, `email`, `phone`, `url`, `file`, `relation`.

`SchemaService.UpdateFields` does a full replace via `FieldRepository.Upsert` + `Delete` on removed fields.

## Records

Each space gets its own Postgres schema (named by slug). Table data lives there as actual SQL tables created from field metadata via the `DDLExecutor`. Record operations use dynamic `map[string]any` keyed by field slugs; the repository constructs queries from the resolved `[]*entity.FieldMeta` slice.

`RecordRepository` methods all take `spaceSlug`, `tableSlug`, and `fields` — the service layer resolves these before calling the repo. `entity.ListParams` carries pagination (`limit`/`offset`), sorting, and full-text search.

## Public REST API (API Keys)

Auto-generated CRUD at `/api/{space}/{table}` authenticated by `X-Api-Key` header.

- Keys have scope `read` (GET only) or `write` (full CRUD)
- `APIKeyAuth` middleware hashes the raw key with SHA-256, looks it up in DB; updates `last_used_at` and writes an audit entry asynchronously (goroutine) to avoid latency
- `GET /api/{space}/openapi.json` → auto-generates OpenAPI 3.0 spec from live table/field metadata
- API status: `GET /spaces/{slug}/api/status` (JWT auth) returns `online`/`offline`/`setup` + active key count + table slugs

Manage keys: `POST /spaces/{slug}/api-keys`, `GET /spaces/{slug}/api-keys`, `DELETE /spaces/{slug}/api-keys/{id}`.

## Lists

Shareable public views on table data. A list pins a table, optional field filter, and publish state.

- `POST/GET/PATCH /spaces/{slug}/lists` — manage lists (JWT auth, space member required)
- `GET /lists/{space}/{list}` — public data endpoint (no auth), returns filtered records

## PDF

Template-based PDF generation pipeline:

1. Upload PDF template (`POST /spaces/{slug}/pdf/templates`) — stored via `StorageAdapter`
2. Define field mappings (`PUT /spaces/{slug}/pdf/templates/{id}/mappings`) — maps PDF form fields → table columns
3. Generate job (`POST /spaces/{slug}/pdf/generate`) — fills template for given record IDs, saves output
4. Poll job status (`GET /spaces/{slug}/pdf/jobs/{id}`) or list archive (`GET /spaces/{slug}/pdf/jobs`)

Jobs are tracked in `pdf_jobs` table; generated files stored via `StorageAdapter`.

## File uploads

`StorageAdapter` (injected into `Server` and `PDFService`) provides `Upload`, `PresignedURL`, `Delete`. Two impls selected at startup:

- `LocalAdapter` — writes to `upload_dir`, presigned URL is just `/files/<key>` (dev)
- `S3Adapter` — minio client; real presigned URLs with configurable TTL

Files routed at `POST /spaces/{slug}/files/upload`, served at `GET /spaces/{slug}/files/{key...}`.

## Audit log

`AuditRepository.Log` is fire-and-forget (called in goroutines). Tracks API key requests and other events with `space_id`, `action`, `entity_type`, `entity_id`, and JSON `meta`.

Query: `GET /spaces/{slug}/audit` (JWT, space admin).

## DB schema (migration 002)

Key tables beyond `users`: `sessions`, `email_verifications`, `password_resets`, `spaces`, `space_members`, `table_meta`, `field_meta`, `api_keys`, `lists`, `pdf_templates`, `pdf_mappings`, `pdf_jobs`, `audit_log`.

## Adding a new domain object

1. Entity struct → `internal/entity/`
2. Repository + service interfaces → `internal/domain/`
3. Repository impl → `internal/repository/`
4. Service impl → `internal/service/`
5. Wire in `internal/app/main.go`
6. Handler + route registration → `internal/transport/http/`

## graphify

This project has a graphify knowledge graph at graphify-out/.

Rules:
- Before answering architecture or codebase questions, read graphify-out/GRAPH_REPORT.md for god nodes and community structure
- If graphify-out/wiki/index.md exists, navigate it instead of reading raw files
- For cross-module "how does X relate to Y" questions, prefer `graphify query "<question>"`, `graphify path "<A>" "<B>"`, or `graphify explain "<concept>"` over grep — these traverse the graph's EXTRACTED + INFERRED edges instead of scanning files
- After modifying code files in this session, run `graphify update .` to keep the graph current (AST-only, no API cost)
