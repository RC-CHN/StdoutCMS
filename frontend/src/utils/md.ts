function inline(html: string): string {
  return html
    .replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
    .replace(/\*(.+?)\*/g, '<em>$1</em>')
    .replace(/`([^`]+)`/g, '<code>$1</code>')
    .replace(/!\[([^\]]*)\]\(([^)]+)\)/g, '<figure class="md-image"><div class="md-image-bar">> $1</div><img src="$2" alt="$1"></figure>')
    .replace(/\[(.+?)\]\((.+?)\)/g, '<a href="$2">$1</a>')
}

export function parseMarkdown(src: string): string {
  const lines = src.split('\n')
  const out: string[] = []
  let i = 0

  while (i < lines.length) {
    const line = lines[i]

    // 空行
    if (!line.trim()) {
      i++
      continue
    }

    // 代码块
    if (line.trim().startsWith('```')) {
      const lang = line.trim().slice(3).trim() || 'code'
      const codeLines: string[] = []
      i++
      while (i < lines.length && !lines[i].trim().startsWith('```')) {
        codeLines.push(lines[i])
        i++
      }
      i++ // skip ```
      out.push(`<pre><div class="code-lang">${escapeHtml(lang)}</div><code>${escapeHtml(codeLines.join('\n'))}</code></pre>`)
      continue
    }

    // h2
    if (line.startsWith('## ')) {
      out.push(`<h2>${inline(line.slice(3))}</h2>`)
      i++
      continue
    }

    // h3
    if (line.startsWith('### ')) {
      out.push(`<h3>${inline(line.slice(4))}</h3>`)
      i++
      continue
    }

    // blockquote
    if (line.startsWith('> ')) {
      const q = [line.slice(2)]
      i++
      while (i < lines.length && lines[i].startsWith('> ')) {
        q.push(lines[i].slice(2))
        i++
      }
      out.push(`<blockquote>${inline(q.join('\n'))}</blockquote>`)
      continue
    }

    // ul
    if (line.startsWith('- ')) {
      const items: string[] = []
      while (i < lines.length && lines[i].startsWith('- ')) {
        items.push(inline(lines[i].slice(2)))
        i++
      }
      out.push(`<ul>${items.map((it) => `<li>${it}</li>`).join('')}</ul>`)
      continue
    }

    // paragraph
    const p = [line]
    i++
    while (i < lines.length && lines[i].trim() && !/^#{1,3} |^> |^- |^```/.test(lines[i])) {
      p.push(lines[i])
      i++
    }
    out.push(`<p>${inline(p.join(' '))}</p>`)
  }

  return out.join('\n')
}

function escapeHtml(s: string): string {
  return s
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
}
