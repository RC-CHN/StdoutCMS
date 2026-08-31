import { describe, it, expect } from 'vitest'
import { parseMarkdown } from './md'

// Each case: [name, input markdown, expected HTML]
const exact: Array<[string, string, string]> = [
  // --- inline nesting fixes ---
  ['bold inside link', '[**bold** link](https://x.com)',
    '<p><a href="https://x.com"><strong>bold</strong> link</a></p>'],
  ['link inside bold', '**[text](https://x.com)**',
    '<p><strong><a href="https://x.com">text</a></strong></p>'],
  ['code span protects markdown', '`**not bold**` and *em*',
    '<p><code>**not bold**</code> and <em>em</em></p>'],
  ['literal math asterisks', '2 * 3 * 4 = 24',
    '<p>2 * 3 * 4 = 24</p>'],
  ['bold and italic', '**bold** and *italic*',
    '<p><strong>bold</strong> and <em>italic</em></p>'],

  // --- XSS / URL handling ---
  ['javascript: link blocked', '[click](javascript:alert(1))',
    '<p><a href="#">click</a></p>'],
  ['strikethrough', '~~gone~~ keep', '<p><del>gone</del> keep</p>'],
  ['horizontal rule', 'above\n\n---\n\nbelow', '<p>above</p>\n<hr>\n<p>below</p>'],
  ['autolink strips trailing punct', 'visit https://example.com.',
    '<p>visit <a href="https://example.com">https://example.com</a>.</p>'],
  ['no autolink inside link syntax', '[label](https://x.com)',
    '<p><a href="https://x.com">label</a></p>'],

  // --- escapes ---
  ['backslash escape', '\\*literal\\*', '<p>*literal*</p>'],

  // --- block parsing ---
  ['h1 maps to h2', '# Title', '<h2 id="title">Title</h2>'],
  ['h3 stays h3', '### Sub', '<h3 id="sub">Sub</h3>'],
  ['blockquote soft break joins', '> line one\n> line two',
    '<blockquote>line one line two</blockquote>'],
  ['blockquote hard break', '> line one  \n> line two',
    '<blockquote>line one<br>line two</blockquote>'],
  ['ordered list', '1. first\n2. second',
    '<ol><li>first</li><li>second</li></ol>'],
  ['nested list', '- a\n  - a1\n  - a2\n- b',
    '<ul><li>a<ul><li>a1</li><li>a2</li></ul></li><li>b</li></ul>'],
  ['paragraph stops at list', 'text\n- item',
    '<p>text</p>\n<ul><li>item</li></ul>'],

  // --- regression guards (pre-rewrite behavior must not break) ---
  ['plain paragraph', 'hello world', '<p>hello world</p>'],
  ['hard break', 'line one  \nline two', '<p>line one<br>line two</p>'],
  ['code fence escapes', '```js\nif (a < b) { }\n```',
    '<pre><div class="code-lang">js</div><code>if (a &lt; b) { }</code></pre>'],
  ['crlf normalized', 'a\r\nb', '<p>a b</p>'],

  // --- compatibility with existing special blocks ---
  ['leading indent preserved', '    indented text',
    '<p>&nbsp;&nbsp;&nbsp;&nbsp;indented text</p>'],
  ['fence with tildes', '~~~\ncode\n~~~',
    '<pre><div class="code-lang">code</div><code>code</code></pre>'],
  ['hr inside paragraph splits', 'text\n---\nmore', '<p>text</p>\n<hr>\n<p>more</p>'],
  ['dash list not hr', '- a\n- b', '<ul><li>a</li><li>b</li></ul>'],
  ['blockquote with emphasis', '> **bold** quote',
    '<blockquote><strong>bold</strong> quote</blockquote>'],

  // --- edge cases ---
  ['empty input', '', ''],
  ['multiple paragraphs', 'a\n\nb', '<p>a</p>\n<p>b</p>'],
  ['html tags escaped in text', '<script>alert(1)</script>',
    '<p>&lt;script&gt;alert(1)&lt;/script&gt;</p>'],
  ['star bullet list', '* a\n* b', '<ul><li>a</li><li>b</li></ul>'],
  ['list item with inline format', '- **bold** item',
    '<ul><li><strong>bold</strong> item</li></ul>'],
  ['blockquote without space', '>quote', '<blockquote>quote</blockquote>'],
  ['heading with inline code', '## `ls` command', '<h2 id="ls-command"><code>ls</code> command</h2>'],
  ['unclosed fence consumes to eof', '```\ncode',
    '<pre><div class="code-lang">code</div><code>code</code></pre>'],
  ['unmatched bold stays literal', '**broken', '<p>**broken</p>'],
  ['strikethrough with space stays literal', 'a ~~ b', '<p>a ~~ b</p>'],
  ['code span inside link label', '[`code`](https://x.com)',
    '<p><a href="https://x.com"><code>code</code></a></p>'],
  ['ordered nested in unordered', '- a\n  1. x\n- b',
    '<ul><li>a<ol><li>x</li></ol></li><li>b</li></ul>'],

  // --- tables (GFM) ---
  ['table basic', '| a | b |\n|---|---|\n| 1 | 2 |',
    '<table class="md-table"><thead><tr><th>a</th><th>b</th></tr></thead>'
    + '<tbody><tr><td>1</td><td>2</td></tr></tbody></table>'],
  ['table without border pipes', 'a | b\n--- | ---\n1 | 2',
    '<table class="md-table"><thead><tr><th>a</th><th>b</th></tr></thead>'
    + '<tbody><tr><td>1</td><td>2</td></tr></tbody></table>'],
  ['table alignment', '| l | c | r |\n|:--|:-:|--:|\n| 1 | 2 | 3 |',
    '<table class="md-table"><thead><tr><th>l</th>'
    + '<th style="text-align:center">c</th><th style="text-align:right">r</th></tr></thead>'
    + '<tbody><tr><td>1</td><td style="text-align:center">2</td>'
    + '<td style="text-align:right">3</td></tr></tbody></table>'],
  ['table cell inline formatting', '| **k** |\n|---|\n| `v` |',
    '<table class="md-table"><thead><tr><th><strong>k</strong></th></tr></thead>'
    + '<tbody><tr><td><code>v</code></td></tr></tbody></table>'],
  ['table interrupts paragraph', 'intro\n| a |\n|---|\n| 1 |',
    '<p>intro</p>\n<table class="md-table"><thead><tr><th>a</th></tr></thead>'
    + '<tbody><tr><td>1</td></tr></tbody></table>'],
  ['table padded cells', '| a | b |\n|---|---|\n| 1 |',
    '<table class="md-table"><thead><tr><th>a</th><th>b</th></tr></thead>'
    + '<tbody><tr><td>1</td><td></td></tr></tbody></table>'],
  ['column mismatch is not a table', '| a | b |\n|---|\n| 1 | 2 |',
    '<p>| a | b | |---| | 1 | 2 |</p>'],

  // --- task lists (GFM) ---
  ['task list unchecked', '- [ ] todo',
    '<ul><li class="md-task"><input type="checkbox" disabled> todo</li></ul>'],
  ['task list checked mixed', '- [x] done\n- [ ] later',
    '<ul><li class="md-task"><input type="checkbox" disabled checked> done</li>'
    + '<li class="md-task"><input type="checkbox" disabled> later</li></ul>'],
  ['bracket link is not a task', '- [link](https://x.com)',
    '<ul><li><a href="https://x.com">link</a></li></ul>'],

  // --- deep list nesting ---
  ['two-level nested list', '- a\n  - a1\n    - a2\n- b',
    '<ul><li>a<ul><li>a1<ul><li>a2</li></ul></li></ul></li><li>b</li></ul>'],
  ['indent jump collapses one level', '- a\n      - deep\n- b',
    '<ul><li>a<ul><li>deep</li></ul></li><li>b</li></ul>'],

  // --- heading anchors ---
  ['duplicate headings numbered', '# Same\n\n# Same',
    '<h2 id="same">Same</h2>\n<h2 id="same-2">Same</h2>'],
]

// Substring checks for cases where full-figure HTML is noisy to assert.
const contains: Array<[string, string, string]> = [
  ['parens in url kept', '[w](https://en.wikipedia.org/wiki/Go_(language))',
    'href="https://en.wikipedia.org/wiki/Go_(language)"'],
  ['query string single-escaped', '[a](https://x.com/?a=1&b=2)',
    'href="https://x.com/?a=1&amp;b=2"'],
  ['media ext with query string', '![a](https://x.com/y.mp3?dl=1)', 'md-audio'],
  ['mailto allowed', '[m](mailto:a@b.c)', 'href="mailto:a@b.c"'],
  ['autolink bare url', 'see https://example.com/a for details',
    '<a href="https://example.com/a">https://example.com/a</a>'],
  ['autolink before code span', 'https://x.com`code`',
    '<a href="https://x.com">https://x.com</a><code>code</code>'],
  ['image renders', '![pic](https://cdn.x.com/a.png)',
    '<img src="https://cdn.x.com/a.png" alt="pic">'],
  ['audio renders', '![song](https://cdn.x.com/a.mp3)',
    '<audio src="https://cdn.x.com/a.mp3"'],
  ['editor paste placeholder survives', '![image](uploading-abc123)', 'uploading-abc123'],
  ['explicit media prefix', '![audio:my song](https://cdn.x.com/f.bin)', 'md-audio'],
  ['explicit file prefix', '![file:report.pdf](https://cdn.x.com/x)', 'md-file'],
  ['autolink with query single-escaped', 'go https://x.com/?a=1&b=2 now',
    'href="https://x.com/?a=1&amp;b=2"'],
  ['image with empty alt', '![](https://cdn.x.com/a.png)',
    '<img src="https://cdn.x.com/a.png" alt="">'],
  ['cjk heading id', '## 中文标题', '<h2 id="中文标题">中文标题</h2>'],
  ['escaped pipe in table cell', '| a \\| b |\n|---|\n| 1 |', '<th>a | b</th>'],
]

describe('parseMarkdown', () => {
  describe('exact output', () => {
    it.each(exact)('%s', (_name, input, expected) => {
      expect(parseMarkdown(input)).toBe(expected)
    })
  })

  describe('output contains', () => {
    it.each(contains)('%s', (_name, input, needle) => {
      expect(parseMarkdown(input)).toContain(needle)
    })
  })

  it('sanitizes javascript: media URLs (href or src depending on type)', () => {
    const html = parseMarkdown('![x](javascript:alert(1))')
    expect(html).not.toContain('javascript:')
    expect(html).toContain('="#"')
  })

  it('mobile flag switches media renderer', () => {
    expect(parseMarkdown('![a](https://x.com/b.mp3)', true)).toContain('md-mobile')
    expect(parseMarkdown('![a](https://x.com/b.mp3)')).not.toContain('md-mobile')
  })

  it('explicit image prefix + javascript: src is neutralized', () => {
    const html = parseMarkdown('![image:x](javascript:alert(1))')
    expect(html).not.toContain('javascript:')
    expect(html).toContain('src="#"')
  })

  it('mobile video renders native controls', () => {
    expect(parseMarkdown('![v](https://x.com/v.mp4)', true))
      .toContain('<video src="https://x.com/v.mp4" preload="metadata" controls playsinline>')
  })
})
