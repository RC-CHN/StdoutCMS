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

---

## License

MIT
