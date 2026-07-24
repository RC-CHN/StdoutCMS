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

// unescapeHtml reverses escapeHtml. inline() escapes the source text before
// matching markdown constructs, so captured groups (URLs, alt text) are
// already escaped; this restores the raw value so it can be sanitized and
// re-escaped exactly once downstream (avoids "&amp;amp;" double-escaping).
function unescapeHtml(s: string): string {
  return s.replace(/&(amp|lt|gt|quot);/g, (_, entity: string) => {
    switch (entity) {
      case 'amp': return '&'
      case 'lt': return '<'
      case 'gt': return '>'
      default: return '"'
    }
  })
}

// sanitizeUrl blocks script-bearing schemes (javascript:, data:, vbscript:)
// in link/media targets. http(s), mailto, anchors and relative URLs pass.
function sanitizeUrl(url: string): string {
  const u = url.trim()
  if (/^[a-z][a-z0-9+.-]*:/i.test(u) && !/^(https?|mailto):/i.test(u)) return '#'
  return u
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

  // auto-detect by file extension (query string / fragment stripped first,
  // otherwise "a.mp3?dl=1" misdetects as a generic file)
  const path = src.split(/[?#]/)[0]
  const ext = path.includes('.') ? path.split('.').pop()!.toLowerCase() : ''
  const inferred = EXT_MAP[ext] ?? 'file'
  return { type: inferred, label: alt }
}

// renderMediaMobile outputs touch-friendly HTML: native <audio>/<video> controls,
// full-width images, compact file download bar.
function renderMediaMobile(type: MediaType, label: string, src: string): string {
  const safeSrc = escapeAttr(src)
  const safeLabel = escapeHtml(label)
  const filename = src.split('/').pop() || 'unknown'
  const display = safeLabel || filename

  switch (type) {
    case 'audio':
      return `<figure class="md-media md-audio md-mobile" data-player>`
        + `<audio src="${safeSrc}" preload="metadata" class="md-player-src"></audio>`
        + `<div class="md-media-meta">`
        + `<div class="md-meta-info">`
        + `<span class="md-audio-name">&#x266B; ${display}</span>`
        + `<span class="md-audio-info">--</span>`
        + `</div>`
        + `<span class="md-status">STATUS: IDLE</span>`
        + `</div>`
        + `<div class="md-player-controls">`
        + `<button class="md-player-btn" data-play>PLAY</button>`
        + `<div class="md-player-track" data-seek><div class="md-player-progress"></div></div>`
        + `<div class="md-player-time">--:-- / --:--</div>`
        + `</div></figure>`
    case 'video':
      return `<figure class="md-media md-video md-mobile">`
        + `<video src="${safeSrc}" preload="metadata" controls playsinline></video>`
        + `</figure>`
    case 'file':
      return `<figure class="md-media md-file md-mobile">`
        + `<a href="${safeSrc}" download class="md-file-link">`
        + `<div class="md-media-meta"><span>FILE: ${display}</span><span>&#x2193;</span></div>`
        + `</a></figure>`
    default: // image — full width, no thumbnail/crop
      return `<figure class="md-media md-image md-mobile">`
        + `<img src="${safeSrc}" alt="${safeLabel}">`
        + (safeLabel ? `<figcaption class="md-image-caption">${safeLabel}</figcaption>` : '')
        + `</figure>`
  }
}

function renderMedia(type: MediaType, label: string, src: string, mobile = false): string {
  if (mobile) return renderMediaMobile(type, label, src)

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

// emphasis parses bold/italic/strikethrough in pre-escaped text. Content
// must start with a non-space and may not contain the delimiter char, so
// "2 * 3 * 4" and "a ~~ b" stay literal.
function emphasis(s: string): string {
  return s
    .replace(/\*\*([^*\s][^*]*)\*\*/g, '<strong>$1</strong>')
    .replace(/\*([^*\s][^*]*)\*/g, '<em>$1</em>')
    .replace(/~~([^~\s][^~]*)~~/g, '<del>$1</del>')
}

// URL body allowing one level of balanced parentheses, e.g. "(...)" inside
// wikipedia-style links. Shared by the image/link patterns below.
const URL_BODY = String.raw`(?:[^()\s]|\([^()\s]*\))+`

// inline transforms inline markdown in raw text.
// Text is escaped first, then each construct is replaced by a placeholder
// token (\x00N\x00) holding its rendered HTML. Later rules therefore cannot
// mangle already-generated markup, which fixes the classic nesting bugs:
// bold/italic inside link labels, markdown inside code spans, and stray
// asterisks inside generated HTML. Placeholders are restored at the end.
function inline(raw: string, mobile = false): string {
  const stash: string[] = []
  const hold = (html: string): string => `\x00${stash.push(html) - 1}\x00`

  let text = escapeHtml(raw)

  // inline code spans — highest precedence, content stays fully literal
  text = text.replace(/`([^`\n]+)`/g, (_, code: string) => hold(`<code>${code}</code>`))

  // backslash escapes (code spans already stashed, so they are unaffected)
  text = text.replace(/\\([\\`*_[\]()>#+.!-])/g, (_, ch: string) => hold(escapeHtml(ch)))

  // images / media embeds
  text = text.replace(new RegExp(String.raw`!\[([^\]]*)\]\((${URL_BODY})\)`, 'g'),
    (_, alt: string, rawUrl: string) => {
      const url = sanitizeUrl(unescapeHtml(rawUrl))
      const { type, label } = detectMedia(unescapeHtml(alt), url)
      return hold(renderMedia(type, label, url, mobile))
    })

  // links — label is already-escaped text, emphasis still applies inside it
  text = text.replace(new RegExp(String.raw`\[([^\]]+)\]\((${URL_BODY})\)`, 'g'),
    (_, label: string, rawHref: string) => {
      const href = escapeAttr(sanitizeUrl(unescapeHtml(rawHref)))
      return hold(`<a href="${href}">${emphasis(label)}</a>`)
    })

  // autolink bare http(s) URLs — explicit links/images are already stashed,
  // so anything left is plain text (useful for chat messages). \x00 is
  // excluded so a URL directly followed by a placeholder token (e.g. an
  // inline code span) does not swallow the token.
  text = text.replace(/https?:\/\/[^\s\x00]+/g, (m: string) => {
    // trailing punctuation is almost never part of the intended URL
    const trail = m.match(/[.,;:!?)\]]*$/)![0]
    const url = m.slice(0, m.length - trail.length)
    const href = escapeAttr(sanitizeUrl(unescapeHtml(url)))
    return hold(`<a href="${href}">${url}</a>`) + trail
  })

  return emphasis(text)
    .replace(/\x00(\d+)\x00/g, (_, n: string) => stash[Number(n)])
}

// preserveIndent converts leading spaces (2+) to &nbsp; so they survive
// HTML whitespace collapse. Applied to flow content after inline().
function preserveIndent(html: string): string {
  return html.replace(/^( {2,})/gm, (_, spaces) => '&nbsp;'.repeat(spaces.length))
}

// ---- block-level parsing ----

const RE_FENCE = /^(`{3,}|~{3,})(.*)$/
const RE_HEADING = /^(#{1,6})\s+(.*)$/
const RE_LIST_ITEM = /^(\s*)([-*]|\d{1,9}\.)\s+(.*)$/
const RE_HR = /^\s*(?:-{3,}|\*{3,}|_{3,})\s*$/
// matches the start of any block construct; used to terminate paragraphs
const RE_BLOCK_START = /^(#{1,6}\s|>|`{3,}|~{3,}|\s*([-*]|\d{1,9}\.)\s|\s*(-{3,}|\*{3,}|_{3,})\s*$)/

interface ListItem {
  sub: boolean     // indented 2+ spaces → nested list item
  ordered: boolean // "1." marker vs "-"/"*"
  html: string
}

// renderList emits <ul>/<ol> with one level of nesting. Nested lists are
// placed inside the preceding <li>, keeping the output valid HTML.
function renderList(items: ListItem[]): string {
  const tagOf = (ordered: boolean): string => (ordered ? 'ol' : 'ul')
  const rootTag = tagOf(items[0].ordered)
  let html = `<${rootTag}>`
  let subTag: string | null = null

  for (const it of items) {
    if (it.sub && subTag === null && html.endsWith('</li>')) {
      subTag = tagOf(it.ordered)
      html = html.slice(0, -'</li>'.length) + `<${subTag}>`
    } else if (!it.sub && subTag !== null) {
      html += `</${subTag}></li>`
      subTag = null
    }
    html += `<li>${it.html}</li>`
  }
  if (subTag !== null) html += `</${subTag}></li>`
  return html + `</${rootTag}>`
}

// renderFlow joins wrapped lines (soft break → space, two trailing spaces →
// hard break), runs inline parsing, then converts hard breaks to <br>.
function renderFlow(ls: string[], mobile: boolean): string {
  let raw = ''
  for (let j = 0; j < ls.length; j++) {
    const hard = ls[j].endsWith('  ')
    const text = hard ? ls[j].replace(/\s+$/, '') : ls[j]
    if (j > 0) raw += ls[j - 1].endsWith('  ') ? '\n' : ' '
    raw += text
  }
  return preserveIndent(inline(raw, mobile)).replace(/\n/g, '<br>')
}

export function parseMarkdown(src: string, mobile = false): string {
  const lines = src.replace(/\r\n?/g, '\n').split('\n')
  const out: string[] = []
  let i = 0

  while (i < lines.length) {
    const line = lines[i]

    // empty line
    if (!line.trim()) {
      i++
      continue
    }

    // fenced code block (``` or ~~~) — content is escaped, not markdown-processed
    const fence = line.trim().match(RE_FENCE)
    if (fence) {
      const marker = fence[1]
      const lang = fence[2].trim() || 'code'
      const codeLines: string[] = []
      i++
      while (i < lines.length && !lines[i].trim().startsWith(marker)) {
        codeLines.push(lines[i])
        i++
      }
      i++ // skip closing fence
      out.push(`<pre><div class="code-lang">${escapeHtml(lang)}</div><code>${escapeHtml(codeLines.join('\n'))}</code></pre>`)
      continue
    }

    // headings: #/## → h2, ### and deeper → h3 (the theme styles only h2/h3)
    const h = line.match(RE_HEADING)
    if (h) {
      const tag = h[1].length <= 2 ? 'h2' : 'h3'
      out.push(`<${tag}>${inline(h[2], mobile)}</${tag}>`)
      i++
      continue
    }

    // horizontal rule: --- / *** / ___ (3+ chars)
    if (RE_HR.test(line)) {
      out.push('<hr>')
      i++
      continue
    }

    // blockquote — consecutive lines joined like a paragraph (<br> aware)
    if (line.startsWith('>')) {
      const q: string[] = []
      while (i < lines.length && lines[i].startsWith('>')) {
        q.push(lines[i].replace(/^>\s?/, ''))
        i++
      }
      out.push(`<blockquote>${renderFlow(q, mobile)}</blockquote>`)
      continue
    }

    // lists: "- "/"* " bullets and "1." ordered, one level of nesting
    if (RE_LIST_ITEM.test(line)) {
      const items: ListItem[] = []
      while (i < lines.length) {
        const m = lines[i].match(RE_LIST_ITEM)
        if (!m) break
        items.push({
          sub: m[1].length >= 2,
          ordered: m[2] !== '-' && m[2] !== '*',
          html: inline(m[3], mobile),
        })
        i++
      }
      out.push(renderList(items))
      continue
    }

    // paragraph — supports hard breaks (2 trailing spaces → <br>)
    const p = [line]
    i++
    while (i < lines.length && lines[i].trim() && !RE_BLOCK_START.test(lines[i])) {
      p.push(lines[i])
      i++
    }
    out.push(`<p>${renderFlow(p, mobile)}</p>`)
  }

  return out.join('\n')
}
