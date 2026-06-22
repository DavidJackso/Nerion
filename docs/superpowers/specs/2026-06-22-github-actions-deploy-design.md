# GitHub Actions Auto-Deploy Design

**Date:** 2026-06-22  
**Scope:** CI/CD pipeline — test → build Docker image → push GHCR → deploy to VPS via SSH

---

## Overview

Push to `main` triggers a 3-job pipeline: run tests, build and push Docker image to GHCR, then SSH to VPS and restart containers via docker compose. `docker-compose.yml` lives in the repo and is copied to the server on every deploy to stay in sync with code.

---

## File Structure

```
.github/
  workflows/
    deploy.yml
Dockerfile
docker-compose.yml
nginx/
  nerion.conf
```

---

## Pipeline Jobs

### 1. test
- Runner: `ubuntu-latest`
- Steps: `go test ./...` + `go vet ./...`
- Failure stops the pipeline — nothing deploys

### 2. build-push
- Needs: `test`
- Login to GHCR via built-in `GITHUB_TOKEN` (no extra secret)
- Build multi-stage image for `linux/amd64`
- Push two tags:
  - `ghcr.io/<owner>/<repo>:latest`
  - `ghcr.io/<owner>/<repo>:<git-sha>`

### 3. deploy
- Needs: `build-push`
- SSH to VPS using `SSH_KEY`, `SSH_HOST`, `SSH_USER` repo secrets
- `scp docker-compose.yml nginx/nerion.conf` → server deploy directory
- `docker compose pull && docker compose up -d`
- Verify: `docker compose ps` confirms containers running

---

## GitHub Repo Secrets

| Secret | Value |
|---|---|
| `SSH_KEY` | ed25519 private key for VPS access |
| `SSH_HOST` | VPS IP or domain |
| `SSH_USER` | SSH user on VPS |

`GITHUB_TOKEN` is automatic — no additional GHCR secret needed.

---

## Dockerfile

Multi-stage build:

- **Stage 1 `builder`:** `golang:1.25-alpine` — compiles binary with `CGO_ENABLED=0`
- **Stage 2 `runtime`:** `alpine:3.20` — copies binary + `migrations/` only

Final image ~15-20 MB. Binary runs as non-root user.

---

## docker-compose.yml

Three services:

| Service | Image | Notes |
|---|---|---|
| `app` | `ghcr.io/<owner>/<repo>:latest` | reads `.env` via `env_file` |
| `postgres` | `postgres:16-alpine` | named volume for data persistence |
| `nginx` | `nginx:alpine` | ports 80/443, proxy_pass → app:8080 |

`.env` file lives only on the server at the deploy directory — never committed. Contains `APP_DB_DSN`, `APP_JWT_SECRET`, and other runtime secrets.

---

## nginx

`nginx/nerion.conf` — reverse proxy config:
- HTTP → HTTPS redirect
- `proxy_pass http://app:8080`
- certbot SSL volumes mounted into nginx container

---

## Server Setup (one-time, manual)

1. Install Docker + docker compose plugin
2. Create deploy directory (e.g. `/opt/nerion`)
3. Place `.env` with all required secrets
4. Add SSH public key to `~/.ssh/authorized_keys`
5. Allow VPS to pull from GHCR (public package or `docker login ghcr.io`)

---

## Rollback

Re-run any previous workflow run → redeploys that commit's `:<sha>` tag.  
Or manually on server: `docker compose pull ghcr.io/<owner>/<repo>:<sha> && docker compose up -d`.
