<template>
  <div class="prompt-line">
    guest@k8s-node ~ % <span style="color: var(--muted);">cat about.md</span>
  </div>

  <article class="article-content">
    <h1 class="about-title">> about_me.txt</h1>

    <p>Hi，我是一名云原生基础设施工程师，日常在 YAML 和 shell 之间横跳。</p>

    <h2>技能树</h2>
    <ul>
      <li>Kubernetes / Docker - 集群管理和容器化部署</li>
      <li>Go / TypeScript - 后端服务和前端开发</li>
      <li>CI/CD - GitHub Actions, ArgoCD</li>
      <li>可观测性 - Prometheus, Grafana, Loki</li>
    </ul>

    <h2>关于这个博客</h2>
    <p>这个博客本身就是一次设计实验。我想探索一种剥离了复杂性的数字表达——没有追踪、没有分析、没有平滑过渡。只有文本和对比。就像早期 web 那样纯粹。</p>

    <blockquote>
      "Simplicity is the ultimate sophistication." — Leonardo da Vinci
    </blockquote>

    <h2>联系方式</h2>
    <p>
      GitHub: <a href="https://github.com">@your_handle</a><br/>
      Email: <code>root@example.com</code><br/>
      GPG: <code>0xDEADBEEF1234</code>
    </p>
  </article>

  <div class="eof-marker">EOF</div>
</template>

<style scoped>
.prompt-line {
  margin-bottom: 2rem;
  font-weight: bold;
}

.about-title {
  font-size: 1.8rem;
  margin-bottom: 2rem;
  color: var(--fg);
}

.article-content {
  font-size: 1.05rem;
}

.article-content :deep(h2) {
  margin-top: 2.5rem;
  margin-bottom: 1rem;
  color: var(--fg);
}
.article-content :deep(h2::before) {
  content: ">> ";
  color: var(--muted);
}

.article-content :deep(p) { margin-bottom: 1.5rem; }
.article-content :deep(ul) { margin-bottom: 1.5rem; padding-left: 2rem; list-style-type: square; }
.article-content :deep(li) { margin-bottom: 0.5rem; }
.article-content :deep(code) { background: var(--fg); color: var(--bg); padding: 2px 6px; font-weight: bold; }
.article-content :deep(blockquote) { border-left: 4px solid var(--border); padding: 1rem 1.5rem; margin: 2rem 0; color: var(--muted); background: rgba(128,128,128,0.05); font-style: italic; }
.article-content :deep(a) { color: var(--fg); text-decoration: underline; text-decoration-style: dashed; }
.article-content :deep(a:hover) { background-color: var(--fg); color: var(--bg); text-decoration: none; }

.eof-marker {
  text-align: center;
  margin: 4rem 0 2rem 0;
  font-weight: bold;
  color: var(--muted);
  letter-spacing: 5px;
}
.eof-marker::before { content: "--- [ "; }
.eof-marker::after { content: " ] ---"; }
</style>
