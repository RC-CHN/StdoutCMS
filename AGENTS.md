# AGENTS.md

## Project overview

**STDOUT_CMS_ELF** is a terminal-themed / brutalist personal blog CMS.
Single-user blog engine: author writes markdown posts in an admin panel,
visitors read them through a retro CLI-styled UI.

- **Name in code**: `stdoutcms` (Go module), `STDOUT_CMS_ELF` (user-facing)
- **Author**: Ruochen Pan (`blog`)
- **License**: MIT

## Tech stack

| Layer | Technology |
|-------|-----------|
| Frontend | Vue 3 + TypeScript + Vite, no component library (hand-rolled CSS) |
| Backend | Go 1.26, [Gin](https://github.com/gin-gonic/gin), [pgx v5](https://github.com/jackc/pgx) |
| Database | PostgreSQL 15 with `pg_trgm` extension for full-text search |
| Cache/Sessions | Redis 7 |
| Object storage | MinIO (S3-compatible) |
| Reverse proxy | Nginx (built into the frontend container) |
| Orchestration | Docker Compose (dev), K8s manifests in `./k8s/` (prod) |

### Go dependencies

```
github.com/gin-gonic/gin          HTTP framework
github.com/jackc/pgx/v5           PostgreSQL driver (pgxpool)
github.com/redis/go-redis/v9      Redis client
github.com/minio/minio-go/v7      S3/MinIO client
github.com/google/uuid            UUID generation
golang.org/x/crypto               bcrypt for password hashing
```

## How to run

### Docker Compose (recommended)

```bash
cp .env.example .env
# Edit .env: change SESSION_SECRET, ADMIN_USERNAME, ADMIN_PASSWORD
docker compose up --build
```

The app is at `http://localhost:8080`. Admin panel at `/admin`.

### Dev mode (manual)

```bash
# Terminal 1 — backend
cd backend
export $(grep -v '^#' ../.env | xargs)  # or source the env vars manually
go run ./cmd/server

# Terminal 2 — frontend
cd frontend
npm install
npm run dev         # Vite on :5173, proxies /api to :8080
```

### Running tests

```bash
cd backend
go test ./...       # all tests
go test ./internal/scheduler  # scheduler only
go test ./internal/task       # orphan image tests
```

```bash
cd frontend
npm test              # vitest — markdown parser suite (src/utils/md.test.ts)
```

Frontend coverage currently targets the markdown parser; Vue components
and composables are still untested.

## Project structure

```
StdoutCMS/
├── AGENTS.md              ← this file
├── README.md              ← human-facing readme
├── docker-compose.yml     ← full stack: postgres, redis, minio, backend, frontend
├── .env.example           ← template for environment variables
├── docs/
│   └── architecture.md    ← detailed architecture doc (API design, auth flow, etc.)
├── backend/
│   ├── go.mod / go.sum
│   ├── Dockerfile
│   ├── cmd/server/main.go        ← entrypoint: wiring, migrations, scheduler start
│   ├── internal/
│   │   ├── config/config.go       ← env var loading (mustEnv for required, getEnv for optional)
│   │   ├── store/
│   │   │   ├── models.go          ← Post, Project, Admin, About structs
│   │   │   ├── postgres.go        ← all PostgreSQL queries (pgxpool)
│   │   │   └── redis.go           ← session, cache, chat history, daily quota
│   │   ├── handler/
│   │   │   ├── handler.go         ← core: Version, slugRegex, Health, Meta, SeedAdmin
│   │   │   ├── post.go            ← public + admin post CRUD (ListPosts, CreatePost, etc.)
│   │   │   ├── project.go         ← public + admin project CRUD
│   │   │   ├── auth.go            ← Login, Logout
│   │   │   ├── upload.go          ← UploadImage (multipart, type/size validation)
│   │   │   ├── about.go           ← GetAbout, GetAboutAdmin, UpdateAbout
│   │   │   ├── chat.go            ← AI chat streaming (SSE, OpenAI-compatible)
│   │   │   └── handler_test.go    ← tests for Health, Meta, generateToken
│   │   ├── middleware/
│   │   │   ├── auth.go            ← Bearer token + cookie session validation
│   │   │   ├── cors.go            ← same-origin only, whitelist localhost:5173
│   │   │   └── logger.go          ← structured request logging via slog
│   │   ├── storage/
│   │   │   └── minio.go           ← S3/MinIO upload, delete, list, GenerateKey
│   │   ├── scheduler/
│   │   │   ├── scheduler.go       ← lightweight periodic task runner (goroutine per task)
│   │   │   └── scheduler_test.go
│   │   └── task/
│   │       ├── cleanup.go         ← orphan image cleanup (scans MinIO, matches post content)
│   │       └── cleanup_test.go
│   └── migrations/
│       ├── 001_init.sql           ← posts, projects, admins tables, pg_trgm indices
│       └── 002_about.sql          ← standalone about table, data migration from posts
├── frontend/
│   ├── package.json
│   ├── vite.config.ts
│   ├── nginx.conf                 ← production nginx config (reverse proxy /api → backend)
│   ├── Dockerfile
│   ├── index.html
│   ├── public/
│   │   └── win-95-98/             ← retro cursor files, favicon, pixel font
│   └── src/
│       ├── main.ts                ← Vue app entry
│       ├── App.vue                ← root shell: header, sidebar, footer, chat widget
│       ├── style.css              ← global styles, CSS custom properties, dark mode
│       ├── api/
│       │   ├── client.ts          ← fetch wrapper (base URL, 401 handling, credentials)
│       │   ├── posts.ts           ← post API calls + PostPayload type
│       │   ├── projects.ts        ← project API calls + ProjectPayload type
│       │   ├── about.ts           ← about API calls + AboutPayload type
│       │   ├── auth.ts            ← login/logout
│       │   ├── upload.ts          ← image upload (FormData)
│       │   ├── meta.ts            ← /meta endpoint
│       │   └── chat.ts            ← SSE streaming chat + generateMeta
│       ├── router/
│       │   └── index.ts           ← routes, lazy-loaded views, auth guard
│       ├── composables/
│       │   ├── useAuth.ts         ← login state (token in memory, sessionStorage flag)
│       │   ├── useChat.ts         ← chat message state, send via SSE
│       │   ├── useDraft.ts        ← localStorage auto-save draft for post editor
│       │   └── useImagePaste.ts   ← clipboard image paste → upload → replace placeholder
│       ├── utils/
│       │   └── md.ts              ← custom markdown parser (hand-written, no library)
│       ├── components/
│       │   ├── TerminalHeader.vue  ← typewriter effect, theme toggle, login/logout
│       │   ├── TerminalFooter.vue  ← marquee with meta info (post count, uptime)
│       │   ├── FileListing.vue     ← `ls -lh` style directory listing (root + articles)
│       │   ├── AdminNav.vue        ← admin tab bar [posts] [projects] [about]
│       │   ├── PostCard.vue        ← blog post card on home page
│       │   ├── ArticleRenderer.vue ← markdown → HTML renderer + image lightbox
│       │   ├── ChatWidget.vue      ← draggable floating chat window (SSE-enabled)
│       │   └── TerminalFeedback.vue ← [ OK ] / [FAIL] status messages
│       ├── views/
│       │   ├── HomeView.vue        ← paginated post list
│       │   ├── ArticleView.vue     ← single post with ArticleRenderer
│       │   ├── AboutView.vue       ← about page (rendered from about.md content)
│       │   ├── ProjectsView.vue    ← project cards
│       │   ├── LoginView.vue       ← login form
│       │   ├── ErrorView.vue       ← Unix signal-themed error pages
│       │   └── admin/
│       │       ├── PostListView.vue   ← admin post table (ls -la style)
│       │       ├── PostEditView.vue   ← split-view markdown editor + AI meta generation
│       │       ├── ProjectListView.vue ← admin project table
│       │       ├── ProjectEditView.vue ← project editor
│       │       └── AboutEditView.vue   ← about page editor
│       └── mocks/
│           ├── posts.ts            ← sample blog posts for development
│           └── projects.ts         ← sample projects for development
├── k8s/
│   └── ...                        ← Kubernetes manifests for production deployment
└── tools/
    └── ...                        ← helper scripts
```

## Module responsibility map

### Backend

| Module | Does | Does NOT do |
|--------|------|-------------|
| `config` | Reads env vars, builds Config struct | Business logic, validation |
| `store/postgres` | All SQL queries via pgxpool, migrations | HTTP handling, caching |
| `store/redis` | Session CRUD, cache get/set/delete, chat history, daily quota | Auth decisions |
| `handler` | HTTP handlers: parse requests, call store, return JSON | Direct SQL, business rules |
| `middleware` | Auth check, CORS headers, request logging | Route registration |
| `storage` | MinIO upload/delete/list, key generation | URL construction (CDN base from config) |
| `scheduler` | Periodic task loop (goroutine per task) | Task definitions |
| `task` | Orphan image cleanup logic | Schedule timing |

### Frontend

| Layer | Does | Does NOT do |
|-------|------|-------------|
| `api/` | Typed fetch wrappers, SSE stream parsing | UI rendering, state |
| `composables/` | Reusable reactive state + side effects | Direct API calls (uses api/) |
| `utils/md.ts` | Markdown → HTML compilation | Rendering (that's ArticleRenderer) |
| `components/` | Presentational + interactive pieces | API calls directly (except ChatWidget) |
| `views/` | Page-level composition, API data fetching | Reusable logic (uses composables) |
| `router/` | Route definitions, lazy loading, auth guard | Auth state (uses useAuth) |

## Key conventions

### Go

- All store methods accept `context.Context` as first parameter.
  Handlers pass `c.Request.Context()`, main.go uses `context.Background()`.
- Handler functions are **closure factories**: `func Xxx(pg, rd) gin.HandlerFunc`.
  The returned `gin.HandlerFunc` captures dependencies, no global state.
- Cache TTL conventions: post list 5min, post detail 30min, project list 10min,
  about page 1h, chat session 1h.
- Cache invalidation: creating/updating/deleting a resource invalidates the
  corresponding cache keys via `rd.InvalidatePost()`, `rd.InvalidateProjects()`,
  `rd.InvalidateAbout()`.
- Slug validation: regex `^[a-z0-9]+(-[a-z0-9]+)*$` (lowercase, hyphens, no
  consecutive hyphens, no leading/trailing hyphen). Applied in CreatePost.
- Image keys: `images/{uuid}.{ext}`, extension derived from MIME type via
  `mime.ExtensionsByType`.
- Migrations: `001_init.sql` is checked for table existence before running.
  `002_about.sql` is unconditional but idempotent (INSERT ... WHERE NOT EXISTS).

### TypeScript/Vue

- API types mirror Go structs exactly: `PostPayload ↔ store.Post`,
  `ProjectPayload ↔ store.Project`, `AboutPayload ↔ store.About`.
- Auth state: token stored in memory (`useAuth`), presence flag in
  `sessionStorage('blog_logged_in')` for UI restore after refresh.
  Actual validation happens on first API call (401 → redirect to /login).
- Drafts: `useDraft()` persists editor state to `localStorage('blog_editor_draft')`,
  auto-saves on change (1s debounce), restored on mount.
- Image paste: `useImagePaste` inserts `![image](uploading-xxx)` placeholder,
  uploads via `uploadImage()`, then replaces placeholder with real URL.
- Markdown parser: hand-written in `utils/md.ts`. Supports h1–h6 (rendered as
  h2/h3), code fences (``` and ~~~), blockquotes, unordered/ordered lists
  (one nesting level), bold, italic, strikethrough, inline code, links,
  images/media embeds, horizontal rules, backslash escapes and bare-URL
  autolinks. No external markdown library.
  Inline parsing is placeholder-based: text is escaped first, each construct
  (code spans, media, links, autolinks) is replaced by a `\x00N\x00` token,
  emphasis runs last, then tokens are restored. Link/media URLs are
  sanitized (javascript:/data: schemes blocked).
- Theme: `dark-mode` class on `body`, preference stored in `localStorage('blog_theme_pref')`.
  Falls back to `prefers-color-scheme: dark`.

## Database

### Tables

```sql
posts (
  id SERIAL PK, slug VARCHAR(255) UNIQUE, title VARCHAR(500),
  content TEXT, excerpt TEXT, tags TEXT[], author VARCHAR(100),
  word_count INT, read_time VARCHAR(20), published BOOLEAN DEFAULT false,
  created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ
)

projects (
  id SERIAL PK, name VARCHAR(255), description TEXT,
  lang VARCHAR(100), status VARCHAR(50), url VARCHAR(500),
  sort_order INT, created_at TIMESTAMPTZ
)

admins (
  id SERIAL PK, username VARCHAR(100) UNIQUE,
  password_hash VARCHAR(255), created_at TIMESTAMPTZ
)

about (
  id SERIAL PK, title VARCHAR(500), content TEXT,
  updated_at TIMESTAMPTZ
)
-- Only one row expected (id=1). Updated via INSERT ... ON CONFLICT.
```

### Indices

- `posts`: slug, published (partial), created_at DESC, tags (GIN), content (GIN trigram)
- `projects`: sort_order
- `admins`: username (unique constraint)

## API overview

### Public (no auth)

| Method | Path | Cache TTL |
|--------|------|-----------|
| GET | /health | none |
| GET | /api/v1/meta | none |
| GET | /api/v1/posts?page=1&size=10 | 5 min |
| GET | /api/v1/posts/:slug | 30 min |
| GET | /api/v1/projects | 10 min |
| GET | /api/v1/projects/:id | 10 min |
| GET | /api/v1/about | 1 hour |
| POST | /api/v1/ai/chat | none (only if LLM configured) |

### Admin (requires session)

All under `/api/v1/admin/`. Auth middleware checks `Authorization: Bearer <token>`
or `blog_session_token` cookie.

| Method | Path | Notes |
|--------|------|-------|
| POST | /login | Sets HttpOnly cookie, returns token in body |
| DELETE | /logout | Clears cookie + Redis session |
| GET/POST | /posts | List (paginated) / Create |
| GET/PUT/DELETE | /posts/:slug | Read (incl. drafts) / Update / Delete |
| POST | /upload | Multipart, max 10MB, image/ only |
| GET/POST | /projects | List / Create |
| PUT/DELETE | /projects/:id | Update / Delete |
| GET/PUT | /about | Read / Update (upsert) |
| POST | /ai/chat | Admin chat (no rate limit, only if LLM configured) |
| POST | /ai/generate-meta | AI slug/tags/excerpt generation |

### AI features

LLM integration is **off by default**. Set `LLM_ENDPOINT`, `LLM_API_KEY`,
`LLM_MODEL` in the environment to enable. When disabled:
- Backend: no AI routes registered
- Frontend: no chat widget shown (`aiEnabled` from /meta)
- `config.LLM.IsEnabled()` guards all AI code paths

### Auth flow

1. `POST /api/v1/admin/login` with username+password
2. Backend verifies bcrypt hash, creates random 64-char hex token
3. Token stored in Redis `session:{token}` with TTL from `SESSION_MAX_AGE`
4. Response sets `blog_session_token` cookie (HttpOnly, SameSite=Strict)
   AND returns `{token}` in JSON body
5. Frontend stores flag in `sessionStorage('blog_logged_in')` for UI state
6. Subsequent requests: cookie auto-sent, or frontend sets Bearer header
7. Middleware checks Redis for valid session

## AI chat implementation notes

- **SSE streaming**: OpenAI-compatible streaming endpoint at
  `{LLM_ENDPOINT}/chat/completions` with `stream: true`
- **Thinking phase**: supports `reasoning_content` in delta (DeepSeek-style).
  Frontend shows "thinking..." indicator, skips displaying reasoning tokens.
- **Article context**: when `ctx` field in request matches a post slug,
  the post content is injected into the conversation as a system-level
  `<article>` block. Context change detection avoids re-injecting.
- **History**: stored in Redis List (`chat:session:{sid}`), limited to 20
  rounds (40 messages). Session TTL 1 hour (no sliding renewal).
- **Daily quota**: public chat uses `chat:quota:daily` counter in Redis,
  auto-expires at midnight. Admin chat has no limit.
- **Metadata generation**: non-streaming call to LLM, XML-tagged output
  (`<slug>`, `<tags>`, `<excerpt>`), parsed with regex.

## Build and deployment

### Docker setup

- **backend**: multi-stage Go build → Alpine binary with migrations
- **frontend**: npm build → Nginx serving static files + reverse proxy
- **minio-init**: one-shot container to create `blog-assets` bucket with
  public download policy
- Health checks on all services ensure correct startup order

### Environment variables

See `.env.example` for full list. Key ones:

```
DATABASE_URL    — postgres://... (required)
REDIS_URL       — redis:6379 (required)
SESSION_SECRET  — change in production! (required)
ADMIN_USERNAME / ADMIN_PASSWORD — default: root/root
STORAGE_*       — MinIO connection (endpoint, bucket, keys)
LLM_*           — optional, leave empty to disable AI
```

## Common development tasks

```bash
# Add a new API endpoint
# 1. Add store method (postgres.go or redis.go) — ctx first param
# 2. Add handler factory in the appropriate handler/*.go file
# 3. Register route in cmd/server/main.go
# 4. Add typed API call in frontend/src/api/
# 5. Call it from a view or composable

# Run Go tests
cd backend && go test ./...

# Check test coverage
cd backend && go test -cover ./...

# Lint Go code
cd backend && go vet ./...

# Type-check frontend
cd frontend && npx vue-tsc --noEmit

# Build frontend
cd frontend && npm run build
```

## Known quirks / debt

1. **No slug validation on UpdatePost** — slug format is only checked on create,
   not update. If someone edits a post through the API directly, a bad slug
   could persist.
2. **Chat session has no sliding TTL** — 1-hour hard timeout from creation,
   no renewal on activity.
3. **Single admin user** — `admins` table supports multiple rows, but seeding
   only creates one, and there's no admin management UI.
4. **Thin frontend test coverage** — only the markdown parser is tested
   (vitest); Vue components and composables are untested.
5. **Markdown parser edge cases** — hand-rolled parser; deep nesting beyond
   one list level, tables, and reference-style links are unsupported.
6. **Store tests missing** — PostgreSQL and Redis are tested only indirectly
   via handler/scheduler/task tests. No isolated DB tests with test fixtures.
7. **No structured logging on frontend** — errors are mostly `console.error`
   or shown as terminal-style messages.
