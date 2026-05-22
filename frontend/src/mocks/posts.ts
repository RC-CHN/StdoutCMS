export interface Post {
  slug: string
  title: string
  date: string
  tags: string[]
  excerpt: string
  content: string
  author: string
  wordCount: number
  readTime: string
}

export const posts: Post[] = [
  {
    slug: 'k8s-stateless-deployment',
    title: 'Kubernetes 无状态应用部署实践：基于纯 CLI 的工作流',
    date: '2023-10-25',
    tags: ['K8s', 'Infra'],
    excerpt:
      '剥离了数据库和静态资源的存储后，服务变得异常轻量。本文记录了如何编写 Deployment 和 Service 文件，以及如何通过纯粹的 CLI 工具流来管理个人博客的发布周期...',
    content: `在决定自己动手提这个博客之前，我考虑过很多方案。由于本身有一台 K8s 集群在跑其他服务，自然想到把博客也塞进去。为了降低维护成本，我决定采用<strong>纯无状态</strong>的设计方案：数据库外挂、图片丢给 S3 兼容存储，前端只渲染 HTML 模板。

## 部署架构概览

核心思路非常简单：构建一个只包含业务逻辑和前端模板的 Docker 镜像，通过 <code>Deployment</code> 控制副本数，用 <code>Service</code> 暴露端口，最后通过 <code>Ingress</code> 绑定域名并配置 TLS 证书。

> "The less state you manage, the better you sleep at night." —— 某不知名运维工程师

### 编写 YAML 文件

对于无状态应用，只需要最基础的声明。以下是 <code>deployment.yaml</code> 的精简版本：

���bash
apiVersion: apps/v1
kind: Deployment
metadata:
  name: terminal-blog
  labels:
    app: blog
spec:
  replicas: 2
  selector:
    matchLabels:
      app: blog
  template:
    metadata:
      labels:
        app: blog
    spec:
      containers:
      - name: blog-app
        image: registry.example.com/blog:v1.0
        ports:
        - containerPort: 8080
���

### 自动化与持续集成

由于前端我使用了简单的黑白终端风格，没有复杂的 Webpack 编译流，我写了一个简单的 Bash 脚本来替代笨重的 CI/CD 工具：

- 执行 <code>git pull</code> 拉取最新 markdown 文章
- 执行 <code>docker build</code> 打包镜像
- 执行 <code>kubectl rollout restart</code> 平滑重启 Pod

你可以查看我的 <a href="https://github.com">GitHub 仓库</a> 获取完整构建脚本。其实在千禧年，很多早期站长就是用简单的 FTP 和脚本完成发布流的，这种粗猖（Brutalist）的方式至今仍然有效。
`,
    author: 'root',
    wordCount: 1024,
    readTime: '3m 12s',
  },
  {
    slug: 'y2k-high-contrast',
    title: '为什么高对比度黑白也是一种 Y2K 表达',
    date: '2023-10-20',
    tags: ['Design', 'Y2K'],
    excerpt:
      '千禧年不仅仅是霓虹粉和镭射光。早期 MacOS 和 Windows 的黑白界面、硬朗的像素线条、不带圆角的粗野主义设计（Brutalism），同样构成了那个时代数字审美的底层代码...',
    content: `千禧年不仅仅是霓虹粉和镭射光。早期 MacOS 和 Windows 的黑白界面、硬朗的像素线条、不带圆角的粗野主义设计（Brutalism），同样构成了那个时代数字审美的底层代码。

## 什么是 Y2K 黑白流

当大家都在追捧千禧年的霓虹彩虹时，我们往往忽略了早期计算机界面的原始美学。

> "Constraints breed creativity."

### 粗野主义的核心特征

- 高对比度配色
- 粗糙的边框和硬边角
- 不装饰的纯粹表达

这种风格在当下的设计师圈子里正在回潮。
`,
    author: 'root',
    wordCount: 856,
    readTime: '2m 45s',
  },
  {
    slug: 'go-concurrency-patterns',
    title: 'Go 并发模式：从 goroutine 到 channel 编排',
    date: '2023-10-15',
    tags: ['Go', 'Concurrency'],
    excerpt:
      'Go 的并发模型看似简单——goroutine + channel，但真正写出优雅的并发代码需要理解几种核心模式。本文总结了 pipeline、fan-out/fan-in、or-done 等实用技巧...',
    content: `Go 的并发模型看似简单，但要真正写出优雅的并发代码，需要理解几种核心模式。

## Pipeline 模式

Pipeline 是最基础的并发模式——数据从一个 stage 流向另一个 stage，每个 stage 由一组 goroutine 处理。

\`\`\`go
func gen(nums ...int) <-chan int {
    out := make(chan int)
    go func() {
        for _, n := range nums {
            out <- n
        }
        close(out)
    }()
    return out
}
\`\`\`

> "Do not communicate by sharing memory; instead, share memory by communicating."

### Fan-out / Fan-in

当某个 stage 的计算量特别大时，可以用多个 goroutine 并行处理同一个 channel。

- 多个 goroutine 从同一个 channel 读取
- 各自处理后将结果发送到各自的输出 channel
- 最后合并到一个 channel
`,
    author: 'root',
    wordCount: 712,
    readTime: '2m 10s',
  },
  {
    slug: 'vim-as-markdown-editor',
    title: '把 Vim 调教成 Markdown 写作利器',
    date: '2023-10-08',
    tags: ['Tools', 'Editor'],
    excerpt:
      '换了无数 Markdown 编辑器，最后还是在终端里用 Vim 最顺手。本文记录了我的 Vim 写作工作流，包括实时预览、拼写检查、词数统计，以及怎么让中文输入体验不那么糟糕...',
    content: `换了无数 Markdown 编辑器，最后还是在终端里用 Vim 最顺手。

## 为什么是 Vim

不是因为信仰，而是因为实用性。Vim 处理纯文本的效率无可匹敌，而 Markdown 恰好就是纯文本。

### 核心插件配置

- **vim-markdown** — 折叠、目录跳转
- **goyo.vim** — 无干扰写作模式
- **vim-pencil** — 软换行、拼写检查

> "Writing is thinking. The tool should disappear."

### 中文输入优化

在 Vim 里写中文最大的痛点是输入法切换。我采用了以下策略：

- 使用 <code>fcitx5</code> 的 Vim 插件自动管理输入法状态
- 用 <code>:w</code> 自动触发格式化和词数统计
`,
    author: 'root',
    wordCount: 598,
    readTime: '1m 50s',
  },
  {
    slug: 'self-hosted-monitoring',
    title: '轻量级自托管监控：Prometheus + Grafana 最简实践',
    date: '2023-09-28',
    tags: ['Infra', 'Observability'],
    excerpt:
      '不用 SaaS、不用 Agent。在单节点上用 Docker Compose 跑一套 Prometheus + Grafana + node_exporter，30 分钟搞定基础监控，顺便聊聊告警规则怎么写最有效...',
    content: `不用 SaaS、不用 Agent。在单节点上 30 分钟搞定基础监控。

## 架构概览

最简单的组合：Prometheus 采集 → Grafana 展示 → Alertmanager 通知。

\`\`\`yaml
version: '3'
services:
  prometheus:
    image: prom/prometheus
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
    ports:
      - "9090:9090"
  grafana:
    image: grafana/grafana
    ports:
      - "3000:3000"
\`\`\`

> "You can't improve what you don't measure."

### 告警规则

好的告警应该是可行动的。不要一看到 CPU 飙升就发通知：

- 按持续时间分级：5 分钟 → warning，15 分钟 → critical
- 区分已知波动和异常模式
- 每个告警附带 runbook 链接
`,
    author: 'root',
    wordCount: 674,
    readTime: '2m 05s',
  },
]

export function getPostBySlug(slug: string): Post | undefined {
  return posts.find((p) => p.slug === slug)
}
