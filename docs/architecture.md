# SYS_BLOG.EXE — 架构设计

## 概述

一个终端风 / Brutalist 风格的个人博客系统。前端 Vue3 SPA，后端 Go API，数据层 PostgreSQL + Redis + MinIO。

---

## 系统拓扑

```
┌─────────────────────────────────────────────────────────┐
│                     Client Browser                       │
│  ┌──────────────┐  ┌──────────────────────────────────┐ │
│  │  Vue3 SPA    │  │  /login  (手动输入，无入口)       │ │
│  │  静态资源    │  │  /admin 编辑器 (需 session)      │ │
│  └──────────────┘  └──────────────────────────────────┘ │
└──────────┬────────────────────┬─────────────────────────┘
           │                    │
     GET / (静态)          /api/v1/*
           │                    │
┌──────────┴────────────────────┴─────────────────────────┐
│                    Nginx / Caddy                         │
│   /          → 前端静态文件                              │
│   /api/*     → Go 后端 :8080                             │
│   /img/*     → MinIO 直出 (或走 API 鉴权)                │
└──────────────────────┬───────────────────────────────────┘
                       │
┌──────────────────────┴───────────────────────────────────┐
│                 Go API Server (:8080)                     │
│                                                          │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐               │
│  │ /admin/* │  │ /posts/* │  │ /upload  │               │
│  │ session  │  │ CRUD     │  │ 图片上传  │               │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘               │
│       │             │              │                     │
│       ▼             ▼              ▼                     │
│  ┌─────────┐  ┌──────────┐  ┌──────────┐                │
│  │ Redis   │  │PostgreSQL│  │  MinIO   │                │
│  │ session │  │ 文章+项目 │  │  图片    │                │
│  │ + cache │  │  全文检索 │  │  S3 API  │                │
│  └─────────┘  └──────────┘  └──────────┘                │
└──────────────────────────────────────────────────────────┘
```

---

## 认证与访问控制

### 前端行为

| 状态 | 表现 |
|------|------|
| 未登录 | 侧边栏无 `admin/` 入口；无其他跳转链接；`/admin/*` 被路由守卫拦截重定向到 `/login` |
| 已登录 | 侧边栏显示 `admin/` 入口；Header 右上角显示 `[ LOGOUT ]`；可正常访问编辑器 |

### 登录流程

1. 用户手动访问 `/login`
2. 提交 username + password → `POST /api/v1/admin/login`
3. 后端验证后返回 `token`，前端存入 **cookie** (SameSite=Strict) + localStorage fallback
4. 前端路由守卫 (`router.beforeEach`) 检查 `meta.requiresAuth`
5. 登出时清除 cookie + localStorage，跳转首页

### Session 存储

- 前端：`document.cookie` 优先，localStorage fallback
- 后端：Redis `session:{token}` → JSON `{ user_id, username, expires_at }`
- 后续 API 请求带 `Authorization: Bearer {token}` header（或 cookie 自动携带）

---

## 组件及职责

| 组件 | 技术选型 | 职责 |
|------|----------|------|
| 前端 | Vue3 + TypeScript + Vite | 博客展示 + 管理后台编辑器 |
| 网关 | Nginx | 静动分离、反向代理、TLS termination |
| 后端 | Go (net/http 或 gin) | REST API，业务逻辑 |
| 数据库 | PostgreSQL 15 | 文章/项目持久化，全文搜索 |
| 缓存 | Redis 7 | admin session + API 热点数据缓存 |
| 对象存储 | MinIO (S3 兼容) | 文章配图及静态文件存储 |

---

## API 设计

### 公开接口 (无需认证)

```
GET   /api/v1/posts           文章列表 (分页: ?page=1&size=10)
GET   /api/v1/posts/:slug     单篇文章全文
GET   /api/v1/projects        项目列表
GET   /api/v1/projects/:id    单个项目详情
```

### 管理接口 (需要 session)

```
POST   /api/v1/admin/login         登录 → 生成 session 存 Redis
DELETE /api/v1/admin/logout        登出 → 清除 Redis session

POST   /api/v1/admin/posts         新建文章
PUT    /api/v1/admin/posts/:slug   编辑文章
DELETE /api/v1/admin/posts/:slug   删除文章

POST   /api/v1/admin/projects      新建项目
PUT    /api/v1/admin/projects/:id  编辑项目
DELETE /api/v1/admin/projects/:id  删除项目

POST   /api/v1/admin/upload        上传图片 → S3/MinIO
DELETE /api/v1/admin/images/:id    删除图片
```

### AI 聊天接口

AI 功能**默认关闭**。未配置 LLM 环境变量时，后端不注册聊天路由，前端不显示聊天入口。

#### 启用方式

```bash
LLM_ENDPOINT=https://api.openai.com/v1   # 必填
LLM_API_KEY=sk-your-key-here             # 必填
LLM_MODEL=gpt-4o                         # 必填
LLM_DAILY_LIMIT=50                       # 可选，默认 50
```

前端通过 `GET /api/v1/meta` 返回的 `{ "ai": true/false }` 判断是否显示相关 UI。

| 端点 | 鉴权 | 限流 |
|------|------|------|
| `POST /api/v1/ai/chat` | 无需 | 50次/天全局限流 |
| `POST /api/v1/admin/ai/chat` | session | 不限 |

#### 请求 / 响应

```
POST /api/v1/ai/chat
  { "q": "用户输入", "ctx": "文章slug", "sid": "chat_abc123" }
  → { "a": "AI 回复内容", "sid": "chat_abc123" }
```

- `q` — 必填，≤500 字符
- `ctx` — 可选文章 slug，后端据此查全文注入 system prompt
- `sid` — 可选，续写已有会话；不传则新建

#### 设计要点

- **非标字段** — `q`/`a`/`sid` 而非 `messages`/`choices`，无法被通用 OpenAI 客户端直连反代
- **有状态会话** — 历史消息存 Redis `chat:session:{sid}`，最大 20 轮，每次对话刷新 TTL 1 小时滑动过期
- **ctx 间接引用** — 前端只传文章 slug，后端自己查全文，攻击者无法灌入任意 system prompt
- **限流** — 公开接口用 Redis `chat:quota:daily` 计数器，每天全局 50 次；admin 接口不限制
- **后端内部翻译** — 将 `q` + 历史 + ctx 组装成 OpenAI 标准 messages 发给 LLM，前端永远看不到原生格式

#### Redis Key

| Key | 类型 | TTL |
|-----|------|-----|
| `chat:session:{sid}` | list (JSON) | 1h 滑动过期 |
| `chat:quota:daily` | string (int) | 当天 23:59 |

---

## 图片上传设计

### 原则

- 编辑器内**不显示任何上传状态指示**（无进度条、无灰色、无 spinning）
- 图片占位符与普通 markdown 文本完全一致
- 上传失败不弹窗提示，用户手动删除重新粘贴
- 追求本地文本编辑器的朴素体验

### 流程

```
用户粘贴 / 拖拽 / 选择图片文件
    ↓
前端捕获文件，在光标处插入占位符文本：
    ![image](uploading-${uuid})
    ↓
异步 POST /api/v1/admin/upload (multipart/form-data)
    ↓
后端：S3 PutObject → 返回 { url: "${CDN_BASE_URL}/images/${uuid}.ext" }
    ↓
前端正则替换占位符：
    ![image](uploading-${uuid})
    → ![image](https://cdn.example.com/images/uuid.png)
```

### 失败处理

- 上传中：占位符就是普通 markdown 文本，预览时显示 broken image
- 上传失败：占位符保持原样 `![image](uploading-${uuid})`，用户手动删掉重来
- 不显示任何 toast / 弹窗 / 状态提示

### 存储配置抽象

```go
type StorageConfig struct {
    Provider   string // "s3" (通用，MinIO 也用 aws-sdk)
    Endpoint   string // "http://localhost:9000" 或 "https://s3.amazonaws.com"
    Bucket     string // "blog-assets"
    Region     string // "us-east-1" (S3) / "" (MinIO 忽略)
    AccessKey  string
    SecretKey  string
    CDNBaseURL string // 对外访问图片的 base URL
    UseSSL     bool
}
```

| 场景 | Endpoint | Region | CDNBaseURL |
|------|----------|--------|------------|
| 本地 MinIO | `http://localhost:9000` | `""` | `http://localhost:9000/blog-assets` |
| 公网 MinIO | `https://minio.example.com` | `""` | `https://minio.example.com/blog-assets` |
| AWS S3 | `https://s3.amazonaws.com` | `us-east-1` | `https://cdn.example.com` |
| Cloudflare R2 | `https://<account>.r2.cloudflarestorage.com` | `auto` | `https://pub-<id>.r2.dev` |

---

## 数据模型

### PostgreSQL — posts

```sql
CREATE TABLE posts (
    id         SERIAL PRIMARY KEY,
    slug       VARCHAR(255) UNIQUE NOT NULL,
    title      VARCHAR(500) NOT NULL,
    content    TEXT NOT NULL,           -- markdown 原文
    excerpt    TEXT,
    tags       TEXT[],                  -- PG 数组
    author     VARCHAR(100) DEFAULT 'root',
    word_count INT DEFAULT 0,
    read_time  VARCHAR(20),
    published  BOOLEAN DEFAULT false,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

-- 全文搜索索引
CREATE INDEX idx_posts_content ON posts USING gin(to_tsvector('simple', content));
```

### PostgreSQL — projects

```sql
CREATE TABLE projects (
    id          SERIAL PRIMARY KEY,
    name        VARCHAR(255) NOT NULL,
    description TEXT,
    lang        VARCHAR(100),
    status      VARCHAR(50),
    url         VARCHAR(500),
    sort_order  INT DEFAULT 0,
    created_at  TIMESTAMPTZ DEFAULT now()
);
```

### Redis

| Key | 类型 | 用途 |
|-----|------|------|
| `session:{token}` | string (JSON) | admin 登录态 |
| `posts:list:{page}` | string (JSON) | 文章列表缓存，TTL 5min |
| `posts:{slug}` | string (JSON) | 单篇文章缓存，TTL 10min |

### MinIO

```
bucket: blog-assets/
├── images/
│   ├── {uuid}.png
│   └── {uuid}.jpg
└── uploads/
```

---

## 项目目录结构

```
StdoutCMS/
├── frontend/                  # Vue3 + TypeScript
│   ├── src/
│   │   ├── api/               # API 调用封装 (待建)
│   │   │   ├── posts.ts
│   │   │   ├── projects.ts
│   │   │   └── auth.ts
│   │   ├── components/        # 复用组件
│   │   ├── composables/       # 🆕 组合式逻辑
│   │   │   ├── useAuth.ts     # 认证状态管理
│   │   │   └── useDraft.ts    # 编辑器草稿自动保存
│   │   ├── views/             # 页面
│   │   │   ├── HomeView.vue
│   │   │   ├── ArticleView.vue
│   │   │   ├── AboutView.vue
│   │   │   ├── ProjectsView.vue
│   │   │   ├── LoginView.vue       # 🆕 登录页
│   │   │   └── admin/
│   │   │       ├── PostListView.vue   # 🆕 文章管理列表
│   │   │       └── PostEditView.vue   # 🆕 Markdown 编辑器
│   │   ├── router/
│   │   ├── mocks/             # 开发用 mock 数据
│   │   └── utils/
│   └── ...
├── backend/
│   ├── cmd/
│   │   └── server/
│   │       └── main.go        # 入口
│   ├── internal/
│   │   ├── handler/           # HTTP handlers
│   │   │   ├── posts.go
│   │   │   ├── projects.go
│   │   │   ├── admin.go
│   │   │   └── upload.go
│   │   ├── store/             # 数据访问层
│   │   │   ├── postgres.go
│   │   │   └── redis.go
│   │   ├── storage/           # 文件存储
│   │   │   └── minio.go
│   │   ├── middleware/
│   │   │   └── auth.go        # session 验证中间件
│   │   └── config/
│   │       └── config.go      # 配置加载
│   ├── migrations/
│   │   └── 001_init.sql
│   ├── go.mod
│   └── Dockerfile
├── docs/
│   └── architecture.md        # 本文档
├── docker-compose.yml
├── Makefile
└── .env.example
```

---

## 开发计划

0. [x] 前端脚手架 + 设计稿落地
1. [x] `/admin` 管理后台 + Markdown 编辑器 + 视图切换 + 自动保存
2. [x] 登录/认证系统 — `/login` 页面 + 路由守卫 + session 管理
3. [ ] `docker-compose.yml` — 起 Postgres + Redis + MinIO
4. [ ] Go 项目骨架 + `/health` 端点
5. [ ] PostgreSQL store — migration + CRUD
6. [ ] API handlers — posts / projects 增删改查
7. [ ] Redis cache — API 读缓存 + admin session
8. [ ] MinIO integration — 图片上传/删除
9. [ ] 前端数据层 — 把 mock 换成 API 调用
