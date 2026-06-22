# GitHub Actions Auto-Deploy Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Wire CI/CD so every push to `main` runs tests, builds a Docker image, pushes it to GHCR, and redeploys on the VPS via SSH + docker compose.

**Architecture:** Three-job pipeline (test → build-push → deploy). Multi-stage Dockerfile produces a ~15 MB alpine image with both `server` and `migrate` binaries. docker-compose.yml lives in the repo and is copied to the server on each deploy; secrets live in a `.env` file on the server that CI never touches except to update `NERION_IMAGE`.

**Tech Stack:** GitHub Actions, Docker (multi-stage), GHCR, docker compose v2, nginx:alpine, postgres:16-alpine, appleboy/ssh-action, webfactory/ssh-agent

## Global Constraints

- Go module name: `nerion` (as in go.mod)
- Go version: 1.25 (as in go.mod) — use `golang:1.25-alpine` builder image
- Target platform: `linux/amd64`
- App listens on `:8080` inside container
- Compose project name: `nerion` (set explicitly in docker-compose.yml)
- Deploy directory on server: `/opt/nerion`
- GHCR image: `ghcr.io/${{ github.repository }}:latest` and `:<sha>`
- Migrations binary: `./cmd/migrate/main.go` → built as `./migrate`

---

### Task 1: Dockerfile and .dockerignore

**Files:**
- Create: `Dockerfile`
- Create: `.dockerignore`

**Interfaces:**
- Produces:
  - Image entrypoint: `CMD ["./server"]`
  - Migrate binary available at `/app/migrate` in image (compose migrate service overrides CMD)
  - Migrations embedded in binary via `migrations/` directory in image

- [ ] **Step 1: Write .dockerignore**

```
.git
.github
bin/
docs/
graphify-out/
uploads/
config.yaml
*.md
```

- [ ] **Step 2: Write Dockerfile**

```dockerfile
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd && \
    CGO_ENABLED=0 GOOS=linux go build -o migrate ./cmd/migrate

FROM alpine:3.20
RUN addgroup -S nerion && adduser -S nerion -G nerion
WORKDIR /app
COPY --from=builder /app/server .
COPY --from=builder /app/migrate .
COPY --from=builder /app/migrations ./migrations
USER nerion
EXPOSE 8080
CMD ["./server"]
```

- [ ] **Step 3: Build image locally to verify**

Run: `docker build -t nerion-test .`
Expected: build succeeds, both stages complete without error

- [ ] **Step 4: Verify image size**

Run: `docker images nerion-test --format "{{.Size}}"`
Expected: under 50 MB (typically ~15-25 MB)

- [ ] **Step 5: Verify server binary runs**

Run: `docker run --rm nerion-test ./server --help 2>&1 || true`
Expected: exits (no panic), may print usage or exit 1 on missing config — that's fine

- [ ] **Step 6: Verify migrate binary exists in image**

Run: `docker run --rm --entrypoint ls nerion-test /app`
Expected: output contains `server` and `migrate`

- [ ] **Step 7: Clean up test image**

Run: `docker rmi nerion-test`

- [ ] **Step 8: Commit**

```bash
git add Dockerfile .dockerignore
git commit -m "feat: add multi-stage Dockerfile with server and migrate binaries"
```

---

### Task 2: docker-compose.yml and nginx config

**Files:**
- Create: `docker-compose.yml`
- Create: `nginx/nerion.conf`

**Interfaces:**
- Consumes: `${NERION_IMAGE}` env var (set by CI in deploy step, or manually)
- Consumes: `/opt/nerion/.env` on server (never committed)
- Produces:
  - `docker compose up -d` starts app + postgres + nginx
  - `docker compose --profile migrate run --rm migrate` runs migrations and exits
  - nginx proxies `80`→HTTPS redirect, `443`→`http://app:8080`

- [ ] **Step 1: Create nginx directory**

Run: `mkdir -p nginx`

- [ ] **Step 2: Write nginx/nerion.conf**

Replace `nerion.example.com` with your actual domain before deploying to server.

```nginx
server {
    listen 80;
    server_name nerion.example.com;

    location /.well-known/acme-challenge/ {
        root /var/www/certbot;
    }

    location / {
        return 301 https://$host$request_uri;
    }
}

server {
    listen 443 ssl;
    server_name nerion.example.com;

    ssl_certificate /etc/letsencrypt/live/nerion.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/nerion.example.com/privkey.pem;

    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;

    location / {
        proxy_pass http://app:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

- [ ] **Step 3: Write docker-compose.yml**

```yaml
name: nerion

services:
  app:
    image: ${NERION_IMAGE:-ghcr.io/change-me/nerion:latest}
    restart: unless-stopped
    env_file: .env
    depends_on:
      postgres:
        condition: service_healthy
    networks:
      - internal

  migrate:
    image: ${NERION_IMAGE:-ghcr.io/change-me/nerion:latest}
    env_file: .env
    command: ["./migrate"]
    depends_on:
      postgres:
        condition: service_healthy
    networks:
      - internal
    profiles: ["migrate"]

  postgres:
    image: postgres:16-alpine
    restart: unless-stopped
    environment:
      POSTGRES_DB: nerion
      POSTGRES_USER: nerion
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U nerion -d nerion"]
      interval: 5s
      timeout: 5s
      retries: 5
    networks:
      - internal

  nginx:
    image: nginx:alpine
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx/nerion.conf:/etc/nginx/conf.d/default.conf:ro
      - /etc/letsencrypt:/etc/letsencrypt:ro
      - /var/www/certbot:/var/www/certbot:ro
    depends_on:
      - app
    networks:
      - internal

volumes:
  postgres_data:

networks:
  internal:
```

- [ ] **Step 4: Validate compose config**

Run: `docker compose config`
Expected: YAML output with no errors, all services listed

- [ ] **Step 5: Validate nginx config syntax**

Run: `docker run --rm -v $(pwd)/nginx/nerion.conf:/etc/nginx/conf.d/default.conf:ro nginx:alpine nginx -t`
Expected: `nginx: configuration file /etc/nginx/nginx.conf test is successful`

Note: nginx test will warn about missing SSL cert paths — that's expected in dev, only matters on server.

- [ ] **Step 6: Commit**

```bash
git add docker-compose.yml nginx/nerion.conf
git commit -m "feat: add docker-compose with app, postgres, nginx, and migrate profile"
```

---

### Task 3: GitHub Actions workflow

**Files:**
- Create: `.github/workflows/deploy.yml`

**Interfaces:**
- Consumes:
  - Repo secrets: `SSH_KEY` (ed25519 private key), `SSH_HOST` (VPS IP/domain), `SSH_USER` (SSH username)
  - Built-in: `GITHUB_TOKEN` (automatic, for GHCR login)
- Produces:
  - On push to `main`: run tests → build+push image → deploy via SSH
  - Image tags: `ghcr.io/<owner>/<repo>:latest` and `ghcr.io/<owner>/<repo>:<sha>`

- [ ] **Step 1: Create workflows directory**

Run: `mkdir -p .github/workflows`

- [ ] **Step 2: Write .github/workflows/deploy.yml**

```yaml
name: Deploy

on:
  push:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
          cache: true

      - name: Test
        run: go test ./...

      - name: Vet
        run: go vet ./...

  build-push:
    needs: test
    runs-on: ubuntu-latest
    permissions:
      contents: read
      packages: write
    steps:
      - uses: actions/checkout@v4

      - uses: docker/login-action@v3
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

      - uses: docker/setup-buildx-action@v3

      - uses: docker/build-push-action@v6
        with:
          context: .
          push: true
          platforms: linux/amd64
          tags: |
            ghcr.io/${{ github.repository }}:latest
            ghcr.io/${{ github.repository }}:${{ github.sha }}

  deploy:
    needs: build-push
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: webfactory/ssh-agent@v0.9.0
        with:
          ssh-private-key: ${{ secrets.SSH_KEY }}

      - name: Add server to known hosts
        run: ssh-keyscan ${{ secrets.SSH_HOST }} >> ~/.ssh/known_hosts

      - name: Copy compose files to server
        run: |
          ssh ${{ secrets.SSH_USER }}@${{ secrets.SSH_HOST }} "mkdir -p /opt/nerion/nginx"
          scp docker-compose.yml \
            ${{ secrets.SSH_USER }}@${{ secrets.SSH_HOST }}:/opt/nerion/docker-compose.yml
          scp nginx/nerion.conf \
            ${{ secrets.SSH_USER }}@${{ secrets.SSH_HOST }}:/opt/nerion/nginx/nerion.conf

      - name: Deploy
        env:
          IMAGE: ghcr.io/${{ github.repository }}:latest
        run: |
          ssh ${{ secrets.SSH_USER }}@${{ secrets.SSH_HOST }} \
            "cd /opt/nerion && \
             NERION_IMAGE='$IMAGE' docker compose pull && \
             NERION_IMAGE='$IMAGE' docker compose --profile migrate run --rm migrate && \
             NERION_IMAGE='$IMAGE' docker compose up -d --remove-orphans app nginx postgres && \
             docker compose ps"
```

- [ ] **Step 3: Validate workflow YAML syntax**

Run: `python3 -c "import yaml; yaml.safe_load(open('.github/workflows/deploy.yml'))" && echo OK`
Expected: `OK`

- [ ] **Step 4: Commit**

```bash
git add .github/workflows/deploy.yml
git commit -m "feat: add GitHub Actions CI/CD workflow (test → build → deploy)"
```

---

## Server Setup Checklist (one-time, manual — not automated by CI)

Before the first deploy, on the VPS:

```bash
# 1. Install Docker
curl -fsSL https://get.docker.com | sh
# Add your user to docker group
usermod -aG docker $USER

# 2. Create deploy directory
mkdir -p /opt/nerion/nginx

# 3. Create .env (CI never modifies anything except NERION_IMAGE on deploy)
cat > /opt/nerion/.env <<EOF
APP_DB_DSN=postgres://nerion:CHANGE_ME@postgres:5432/nerion
APP_JWT_SECRET=CHANGE_ME_STRONG_SECRET
APP_HTTP_ADDR=:8080
POSTGRES_PASSWORD=CHANGE_ME
EOF

# 4. Set up SSL with certbot (after DNS points to this server)
apt install certbot
certbot certonly --standalone -d nerion.example.com

# 5. Generate ed25519 SSH key for CI (on your local machine, not server)
ssh-keygen -t ed25519 -C "github-actions" -f ~/.ssh/nerion_deploy

# 6. Add public key to server
cat ~/.ssh/nerion_deploy.pub >> ~/.ssh/authorized_keys

# 7. Add private key to GitHub repo secrets
# GitHub → repo → Settings → Secrets → Actions:
#   SSH_KEY   = contents of ~/.ssh/nerion_deploy (private key)
#   SSH_HOST  = your server IP or domain
#   SSH_USER  = your SSH user on server
```

## Rollback

Re-run any previous workflow run in GitHub Actions → redeploys that commit's `:<sha>` image.

Or manually on server:
```bash
cd /opt/nerion
NERION_IMAGE=ghcr.io/<owner>/nerion:<sha> docker compose up -d app
```
