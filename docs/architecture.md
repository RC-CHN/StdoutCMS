# SYS_BLOG.EXE — 架构设计

## 概述

一个终端风 / Brutalist 风格的个人博客系统。前端 Vue3 SPA，后端 Go API，数据层 PostgreSQL + Redis + MinIO。

---

## 系统拓扑

```
┌─────────────────────────────────────────────────────────┐
│                     Client Browser                       │
│  ┌──────────────┐  ┌──────────────────────────────────┐ │
│  │  Vue3 SPA    │  │  /admin 编辑器 (需 session)      │ │
│  │  静态资源    │  │  增删改文章/项目/图片             │ │
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
DELETE /api/v1/admin/logout        登出

POST   /api/v1/admin/posts         新建文章
PUT    /api/v1/admin/posts/:slug   编辑文章
DELETE /api/v1/admin/posts/:slug   删除文章

POST   /api/v1/admin/projects      新建项目
PUT    /api/v1/admin/projects/:id  编辑项目
DELETE /api/v1/admin/projects/:id  删除项目

POST   /api/v1/admin/upload        上传图片 → MinIO
DELETE /api/v1/admin/images/:id    删除图片
```

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
│   │   ├── api/               # API 调用封装
│   │   │   ├── posts.ts
│   │   │   ├── projects.ts
│   │   │   └── auth.ts
│   │   ├── components/        # 复用组件
│   │   ├── views/             # 页面
│   │   │   ├── HomeView.vue
│   │   │   ├── ArticleView.vue
│   │   │   ├── AboutView.vue
│   │   │   ├── ProjectsView.vue
│   │   │   └── AdminView.vue  # 🆕 管理后台
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
1. [ ] `docker-compose.yml` — 起 Postgres + Redis + MinIO
2. [ ] Go 项目骨架 + `/health` 端点
3. [ ] PostgreSQL store — migration + CRUD
4. [ ] API handlers — posts / projects 增删改查
5. [ ] Redis cache — API 读缓存 + admin session
6. [ ] MinIO integration — 图片上传/删除
7. [ ] 前端数据层 — 把 mock 换成 API 调用
8. [ ] 编辑器页面 — `/admin` 管理后台
