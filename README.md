# STDOUT_CMS_ELF

> A terminal-style / brutalist personal blog CMS.

```
root@k8s-node ~ $ ./STDOUT_CMS_ELF --help
STATELESS *** ALL GREEN *** 0 TRACKERS *** 0 ANALYTICS
```

---

## Overview

A minimal, terminal-themed blog engine built for people who like their web
content raw. No frameworks, no bloat — just markdown-to-HTML with brutalist
aesthetics.

- **Frontend**: Vue 3 + TypeScript + Vite
- **Backend**: Go + Gin + PostgreSQL + Redis + MinIO
- **Deploy**: Docker Compose (dev) or K8s manifests (prod)

---

## Quick Start

```bash
cp .env.example .env
docker compose up --build
```

Open `http://localhost:8080`. Admin panel at `/admin` (default: `root` / `root`).

> Change `ADMIN_USERNAME`, `ADMIN_PASSWORD`, and `SESSION_SECRET` in `.env` for
> production.

---

## Features

- **Terminal UI** — `$ ls -lh` sidebar, typewriter header, marquee footer
- **Markdown posts** — slug-based URLs, tags, word count, read time
- **Project showcase** — lang / status / url fields
- **About page** — standalone singleton page
- **Admin panel** — split-view markdown editor with live preview
- **Writing recovery** — local backups per article, explicit restore, and separate save/publish/unpublish actions
- **Reading navigation** — article contents, pagination and scroll restoration, mobile sharing
- **Mobile editing** — single-pane edit/preview, accessible post actions, retryable uploads
- **Image paste** — paste images directly into the editor, uploaded to MinIO
- **Dark mode** — `[ INVERT_COLORS ]` toggle, persists to localStorage
- **AI chat** (optional) — SSE streaming chat widget, article-aware context
- **AI metadata** (optional) — auto-generate slug / tags / excerpt from content

---

## AI Features

LLM features are **off by default**. To enable:

```env
LLM_ENDPOINT=https://api.openai.com/v1
LLM_API_KEY=sk-your-key-here
LLM_MODEL=gpt-4o
LLM_DAILY_LIMIT=50
```

If left empty, the chat widget and AI endpoints are completely disabled —
no extra routes registered, no UI shown.

---

## API

### Public

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/posts?page=1&size=10` | Paginated post list |
| GET | `/api/v1/posts/:slug` | Single post |
| GET | `/api/v1/projects` | Project list |
| GET | `/api/v1/projects/:id` | Single project |
| GET | `/api/v1/about` | About page |
| GET | `/api/v1/meta` | System info |
| POST | `/api/v1/ai/chat` | AI chat (if enabled) |

### Admin (requires session)

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/admin/login` | Login → sets cookie + returns token |
| DELETE | `/api/v1/admin/logout` | Logout → clears session |
| GET/POST/PUT/DELETE | `/api/v1/admin/posts` | Post CRUD |
| GET/POST/PUT/DELETE | `/api/v1/admin/projects` | Project CRUD |
| GET/PUT | `/api/v1/admin/about` | About CRUD |
| POST | `/api/v1/admin/upload` | Image upload |
| POST | `/api/v1/admin/ai/chat` | Admin AI chat (no rate limit) |
| POST | `/api/v1/admin/ai/generate-meta` | AI metadata generation |

---

## Architecture

```
Browser
  │
  ├── / → Vue 3 SPA (static)
  │
  └── /api/* → Go API (:8080)
                │
                ├── PostgreSQL 15  (posts, projects, admins, about)
                ├── Redis 7        (sessions, cache, chat history)
                └── MinIO          (image storage, S3 API)
```

Full architecture doc: [docs/architecture.md](docs/architecture.md)

---

## Development

```bash
# Backend
cd backend
go run ./cmd/server

# Frontend
cd frontend
npm install
npm run dev        # Vite dev server on :5173, proxies /api to :8080
```

Run backend tests with `cd backend && go test ./...`. PostgreSQL integration
tests require `TEST_DATABASE_URL` pointing to a test database; they use temporary
tables and are skipped when the variable is unset. CI runs them against PostgreSQL 15.

```bash
# From backend/, with a local test database running:
TEST_DATABASE_URL='postgres://user:password@localhost:5432/stdoutcms_test?sslmode=disable' go test ./...
```

Frontend checks: `cd frontend && npm test && npm run build`. Tests cover the
Markdown parser, API errors, local draft recovery, and upload failure/retry.
Local backups are stored in the current browser; use **Save draft** or
**Save changes** to save to the server. Saving a published article keeps it
published; **Unpublish** is a separate action with confirmation.

---

## Versioning and releases

The root [`VERSION`](VERSION) file is the single application version source,
using `X.Y.Z` without a `v` prefix. Both container images use this version as
their tag; the backend embeds it and exposes it through `/api/v1/meta`.
The private frontend package has no independent version.

For a versioned local backend binary, run `sh tools/build-backend.sh` from
the repository root. Plain `go run ./cmd/server` still reports `dev`.
Standalone Docker builds now use the repository root as their context:
`docker build -f backend/Dockerfile .`. Docker Compose handles this automatically.

To release from a clean, up-to-date `main` branch:

```bash
git pull --ff-only
git fetch --tags
# Commit all features, fixes, and optional docs/releases/vX.Y.Z.md notes first.
sh tools/release.sh X.Y.Z
git push --atomic origin main vX.Y.Z
```

Replace `X.Y.Z` with the next stable version. The script updates only `VERSION`,
creates `chore: release vX.Y.Z`, and points an annotated `vX.Y.Z` tag at that
commit. It rejects dirty worktrees, existing local tags, and non-increasing
versions. Historical tags are unchanged.

The Release workflow checks the tagged commit's subject, changed files, and
version before running CI. Only after the checks and both GHCR image builds
succeed does it publish the GitHub Release, using committed release notes when
available. Invalid release tags cannot publish through this workflow.
Release tooling tests run with
`python3 -B -m unittest discover -s tools -p 'test_release.py'`.

---

## License

WTFPL — DO WHAT THE FUCK YOU WANT TO PUBLIC LICENSE, Version 2.
See [LICENSE](LICENSE).
