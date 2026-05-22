export interface Project {
  name: string
  description: string
  lang: string
  status: string
  url: string
}

export const projects: Project[] = [
  {
    name: 'k8s-cluster-kit',
    description: '一键部署轻量 K8s 学习集群的 Ansible playbook 集合，适合实验和本地开发。',
    lang: 'YAML / Bash',
    status: 'ACTIVE',
    url: 'https://github.com',
  },
  {
    name: 'terminal-blog',
    description: '就是这个博客本身。Vue 3 SSG + Go 后端，全无状态部署。',
    lang: 'Vue / Go',
    status: 'BUILDING',
    url: 'https://github.com',
  },
  {
    name: 'log-parser-tui',
    description: '一个终端 UI 日志分析器，支持过滤、高亮和实时 tail。',
    lang: 'Go',
    status: 'MAINTAINING',
    url: 'https://github.com',
  },
  {
    name: 'dotfiles',
    description: '个人 dotfiles 配置仓库，包含 zsh、tmux、neovim 等工具的配置文件和安装脚本。',
    lang: 'Shell / Lua',
    status: 'MAINTAINING',
    url: 'https://github.com',
  },
  {
    name: 'go-metrics',
    description: '轻量级 Go metrics 库，支持 Prometheus 和自定义 reporter，零外部依赖。',
    lang: 'Go',
    status: 'ACTIVE',
    url: 'https://github.com',
  },
]
