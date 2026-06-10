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

// ---- media type detection ----

type MediaType = 'image' | 'audio' | 'video' | 'file'

const EXT_MAP: Record<string, MediaType> = {
  // images
  jpg: 'image', jpeg: 'image', png: 'image', gif: 'image',
  webp: 'image', svg: 'image', bmp: 'image', ico: 'image',
  // audio
  mp3: 'audio', wav: 'audio', ogg: 'audio', flac: 'audio',
  aac: 'audio', m4a: 'audio', wma: 'audio', opus: 'audio',
  // video
  mp4: 'video', webm: 'video', ogv: 'video', mov: 'video',
  avi: 'video', mkv: 'video',
}

function detectMedia(alt: string, src: string): { type: MediaType; label: string } {
  // explicit prefix: "image:", "audio:", "video:", "file:"
  const m = alt.match(/^(image|audio|video|file):(.*)/)
  if (m) return { type: m[1] as MediaType, label: m[2] || '' }

  // auto-detect by file extension
  const ext = src.split('.').pop()?.toLowerCase() ?? ''
  const inferred = EXT_MAP[ext] ?? 'file'
  return { type: inferred, label: alt }
}

function renderMedia(type: MediaType, label: string, src: string): string {
  const safeSrc = escapeAttr(src)
  const safeLabel = escapeHtml(label)
  const filename = src.split('/').pop() || 'unknown'
  const display = safeLabel || filename

  switch (type) {
    case 'audio':
      return `<figure class="md-media md-audio" data-player>`
        + `<audio src="${safeSrc}" preload="metadata" class="md-player-src"></audio>`
        + `<div class="md-media-meta">`
        + `<div class="md-meta-info">`
        + `<span class="md-audio-name">&#x266B; ${display}</span>`
        + `<span class="md-audio-info">--</span>`
        + `<span class="md-status">STATUS: IDLE</span>`
        + `</div>`
        + `</div>`
        + `<div class="md-player-controls">`
        + `<button class="md-player-btn" data-play>PLAY</button>`
        + `<div class="md-player-track" data-seek><div class="md-player-progress"></div></div>`
        + `<div class="md-player-time">--:-- / --:--</div>`
        + `<button class="md-player-btn" data-mute>VOL: 100%</button>`
        + `</div></figure>`
    case 'video':
      return `<figure class="md-media md-video" data-player data-collapse>`
        + `<div class="md-media-meta">`
        + `<div class="md-meta-info"><span>VIDEO</span><span class="md-res">RES: --</span></div>`
        + `<button class="md-toggle-btn">[+] REVEAL</button>`
        + `</div>`
        + `<div class="md-foldable-content md-video-body">`
        + `<video src="${safeSrc}" preload="metadata" class="md-player-src"></video>`
        + `<div class="md-player-controls">`
        + `<button class="md-player-btn" data-play>PLAY</button>`
        + `<div class="md-player-track" data-seek><div class="md-player-progress"></div></div>`
        + `<div class="md-player-time">--:-- / --:--</div>`
        + `<button class="md-player-btn" data-mute>VOL: 100%</button>`
        + `<button class="md-player-btn" data-fullscreen title="fullscreen">&#x26F6;</button>`
        + `</div></div></figure>`
    case 'file':
      return `<figure class="md-media md-file">`
        + `<a href="${safeSrc}" download class="md-file-link">`
        + `<div class="md-media-meta"><span>FILE</span><span>GET</span></div>`
        + `<div class="md-file-body">`
        + `<h3 class="md-file-name">${display}</h3>`
        + `<div>[ CLICK TO DOWNLOAD ]</div>`
        + `</div></a></figure>`
    default: // image — collapsible: thumbnail → full
      return `<figure class="md-media md-image" data-collapse>`
        + `<div class="md-media-meta">`
        + `<div class="md-meta-info"><span>IMG</span><span class="md-img-dims">DIM: --</span></div>`
        + `<button class="md-toggle-btn">[+] REVEAL</button>`
        + `</div>`
        + `<div class="md-image-body">`
        + `<img src="${safeSrc}" alt="${safeLabel}">`
        + `<div class="md-image-caption">${safeLabel}</div>`
        + `</div></figure>`
  }
}

// inline transforms inline markdown in pre-escaped text.
// Text is escaped BEFORE markdown rules are applied.
function inline(html: string): string {
  const safe = escapeHtml(html)
  return safe
    .replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
    .replace(/\*(.+?)\*/g, '<em>$1</em>')
    .replace(/`([^`]+)`/g, '<code>$1</code>')
    .replace(/!\[([^\]]*)\]\(([^)]+)\)/g, (_, alt: string, src: string) => {
      const { type, label } = detectMedia(alt, src)
      return renderMedia(type, label, src)
    })
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
