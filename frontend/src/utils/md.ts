function escapeHtml(s: string): string {
  return s
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
}

// escapeAttr escapes for HTML attribute context (img alt, link titles, etc.)
function escapeAttr(s: string): string {
  return escapeHtml(s).replace(/'/g, '&#39;')
}

// inline transforms inline markdown in pre-escaped text.
// Text is escaped BEFORE markdown rules are applied.
function inline(html: string): string {
  const safe = escapeHtml(html)
  return safe
    .replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
    .replace(/\*(.+?)\*/g, '<em>$1</em>')
    .replace(/`([^`]+)`/g, '<code>$1</code>')
    .replace(/!\[([^\]]*)\]\(([^)]+)\)/g, (_, alt: string, src: string) =>
      `<figure class="md-image"><div class="md-image-bar">&gt; ${escapeHtml(alt)}</div><img src="${escapeAttr(src)}" alt="${escapeAttr(alt)}"></figure>`
    )
    .replace(/\[(.+?)\]\((.+?)\)/g, (_, text: string, href: string) =>
      `<a href="${escapeAttr(href)}">${text}</a>`
    )
}

// preserveIndent converts leading spaces (2+) to &nbsp; so they survive
// HTML whitespace collapse. Applied to paragraph content after inline().
function preserveIndent(html: string): string {
  return html.replace(/^( {2,})/gm, (_, spaces) => '&nbsp;'.repeat(spaces.length))
}

export function parseMarkdown(src: string): string {
  const lines = src.split('\n')
  const out: string[] = []
  let i = 0

  while (i < lines.length) {
    const line = lines[i]

    // empty line
    if (!line.trim()) {
      i++
      continue
    }

    // fenced code block — content is escaped, not markdown-processed
    if (line.trim().startsWith('```')) {
      const lang = line.trim().slice(3).trim() || 'code'
      const codeLines: string[] = []
      i++
      while (i < lines.length && !lines[i].trim().startsWith('```')) {
        codeLines.push(lines[i])
        i++
      }
      i++ // skip closing ```
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

    // paragraph — supports hard breaks (2 trailing spaces → <br>)
    const p = [line]
    i++
    while (i < lines.length && lines[i].trim() && !/^#{1,3} |^> |^- |^```/.test(lines[i])) {
      p.push(lines[i])
      i++
    }

    // Build paragraph text: soft-break lines joined with ' ', hard-break
    // lines (ending with 2 spaces) joined with '\n' then replaced with <br>.
    let raw = ''
    for (let j = 0; j < p.length; j++) {
      let text = p[j]
      const hardBreak = text.endsWith('  ')
      if (hardBreak) text = text.slice(0, -2)
      if (j > 0) {
        raw += (p[j - 1].endsWith('  ') ? '\n' : ' ')
      }
      raw += text
    }
    out.push(`<p>${preserveIndent(inline(raw).replace(/\n/g, '<br>'))}</p>`)
  }

  return out.join('\n')
}
