<script setup lang="ts">
import { computed, onMounted, watch, nextTick } from 'vue'
import { parseMarkdown } from '../utils/md'

const props = defineProps<{
  content: string
}>()

const html = computed(() => parseMarkdown(props.content, true))

// images: open in new tab on mobile (native browser zoom)
function onClick(e: MouseEvent) {
  const target = e.target as HTMLElement
  const img = target.closest('.md-mobile img') as HTMLImageElement | null
  if (img) {
    window.open(img.src, '_blank')
  }
}

// ---- custom audio player (brutalist, touch-friendly) ----

function fmtTime(s: number): string {
  if (!isFinite(s)) return '--:--'
  const m = Math.floor(s / 60)
  const sec = Math.floor(s % 60)
  return `${m}:${sec.toString().padStart(2, '0')}`
}

function initPlayers(root: HTMLElement) {
  root.querySelectorAll<HTMLElement>('.md-audio.md-mobile[data-player]').forEach(player => {
    if ((player as any).__inited) return
    ;(player as any).__inited = true

    const src = player.querySelector<HTMLMediaElement>('.md-player-src')
    const btn = player.querySelector<HTMLElement>('[data-play]')
    const track = player.querySelector<HTMLElement>('[data-seek]')
    const bar = track?.querySelector<HTMLElement>('.md-player-progress')
    const timeEl = player.querySelector<HTMLElement>('.md-player-time')
    const status = player.querySelector<HTMLElement>('.md-status')
    const info = player.querySelector<HTMLElement>('.md-audio-info')

    if (!src || !btn || !track || !bar || !timeEl) return

    let duration = NaN

    src.addEventListener('loadedmetadata', () => {
      duration = src.duration
      timeEl.textContent = `0:00 / ${fmtTime(duration)}`
      if (info) {
        const ext = (src.src.split('.').pop() || '').toUpperCase()
        info.textContent = `${ext} · ${fmtTime(duration)}`
      }
    })

    src.addEventListener('timeupdate', () => {
      if (!isNaN(duration) && duration > 0) {
        bar.style.width = (src.currentTime / duration * 100) + '%'
      }
      timeEl.textContent = `${fmtTime(src.currentTime)} / ${fmtTime(duration)}`
    })

    src.addEventListener('ended', () => {
      btn.textContent = 'PLAY'
      bar.style.width = '0%'
      if (status) status.textContent = 'STATUS: DONE'
    })

    src.addEventListener('play', () => {
      btn.textContent = 'PAUSE'
      if (status) status.textContent = 'STATUS: PLAYING'
    })

    src.addEventListener('pause', () => {
      btn.textContent = 'PLAY'
      if (status) status.textContent = 'STATUS: PAUSED'
    })

    btn.addEventListener('click', () => {
      if (src.paused) src.play()
      else src.pause()
    })

    // ---- seek (touch + click) ----
    function seekFromX(clientX: number) {
      if (!src || isNaN(duration)) return
      const rect = track!.getBoundingClientRect()
      src.currentTime = Math.max(0, Math.min(1, (clientX - rect.left) / rect.width)) * duration
    }

    track!.addEventListener('click', (e) => {
      seekFromX((e as MouseEvent).clientX)
    })

    track!.addEventListener('touchstart', (e) => {
      seekFromX(e.touches[0].clientX)

      const onMove = (e: TouchEvent) => {
        e.preventDefault()
        seekFromX(e.touches[0].clientX)
      }
      const onEnd = () => {
        document.removeEventListener('touchmove', onMove)
        document.removeEventListener('touchend', onEnd)
        document.removeEventListener('touchcancel', onEnd)
      }
      document.addEventListener('touchmove', onMove, { passive: false })
      document.addEventListener('touchend', onEnd)
      document.addEventListener('touchcancel', onEnd)
    }, { passive: true })
  })
}

watch(html, () => nextTick(() => {
  const el = document.querySelector('.article-content') as HTMLElement | null
  if (el) initPlayers(el)
}))

onMounted(() => {
  nextTick(() => {
    const el = document.querySelector('.article-content') as HTMLElement | null
    if (el) initPlayers(el)
  })
})
</script>

<template>
  <div class="article-content" v-html="html" @click="onClick" />
</template>

<style scoped>
.article-content {
  font-size: 1rem;
}

.article-content :deep(h2) {
  margin-top: 2rem;
  margin-bottom: 0.8rem;
  color: var(--fg);
}
.article-content :deep(h2::before) {
  content: ">> ";
  color: var(--accent);
}

.article-content :deep(h3) {
  margin-top: 2rem;
  margin-bottom: 0.8rem;
  color: var(--fg);
}
.article-content :deep(h3::before) {
  content: "# ";
  color: var(--accent);
}

.article-content :deep(p) {
  margin-bottom: 1.2rem;
}

.article-content :deep(a) {
  color: var(--fg);
  text-decoration: underline;
  text-decoration-style: dashed;
  text-underline-offset: 3px;
}

.article-content :deep(ul) {
  margin-bottom: 1.2rem;
  padding-left: 1.5rem;
  list-style-type: square;
}

.article-content :deep(li) {
  margin-bottom: 0.4rem;
}

.article-content :deep(blockquote) {
  border-left: 3px solid var(--accent);
  padding: 0.8rem 1rem;
  margin: 1.5rem 0;
  color: var(--muted);
  background: rgba(128, 128, 128, 0.05);
}

.article-content :deep(code) {
  background: var(--fg);
  color: var(--bg);
  padding: 2px 5px;
  font-size: 0.85em;
  font-weight: bold;
}

.article-content :deep(pre) {
  background: var(--bg);
  border: 2px solid var(--border);
  padding: 0.8rem;
  overflow-x: auto;
  margin: 1.5rem 0;
  font-size: 0.8rem;
}

.article-content :deep(.code-lang) {
  border-bottom: 2px dashed var(--border);
  padding-bottom: 4px;
  margin-bottom: 8px;
  font-size: 0.75rem;
  color: var(--muted);
  text-transform: uppercase;
}

.article-content :deep(pre code) {
  background: transparent;
  color: var(--fg);
  padding: 0;
  font-weight: normal;
}

.article-content :deep(strong) {
  color: var(--fg);
}

/* ---- mobile media ---- */

.article-content :deep(.md-media.md-mobile) {
  display: block;
  margin: 1.5rem 0;
  border: 2px solid var(--border);
  background: var(--bg);
  max-width: 100%;
}

.article-content :deep(.md-media-meta) {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.4rem 0.6rem;
  font-size: 0.72rem;
  font-weight: bold;
  background: var(--fg);
  color: var(--bg);
  text-transform: uppercase;
  font-family: var(--font-main);
}

/* image: full-width, no crop, no grayscale */
.article-content :deep(.md-image.md-mobile) {
  border: 2px solid var(--border);
}

.article-content :deep(.md-image.md-mobile img) {
  display: block;
  width: 100%;
  height: auto;
  cursor: url('/win-95-98/hand.cur'), pointer;
}

.article-content :deep(.md-image-caption) {
  padding: 0.4rem 0.6rem;
  font-size: 0.75rem;
  border-top: 2px solid var(--border);
  color: var(--muted);
  font-family: var(--font-main);
}

/* audio: brutalist custom player */
.article-content :deep(.md-audio.md-mobile .md-player-src) {
  display: none;
}

.article-content :deep(.md-audio.md-mobile .md-meta-info) {
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
}

.article-content :deep(.md-audio-name) {
  color: var(--bg);
  text-transform: none;
}

.article-content :deep(.md-audio-info) {
  color: var(--bg);
  opacity: 0.7;
  font-size: 0.65rem;
  text-transform: uppercase;
}

.article-content :deep(.md-status) {
  color: var(--bg);
  opacity: 0.7;
  font-size: 0.65rem;
  flex-shrink: 0;
}

.article-content :deep(.md-player-controls) {
  display: flex;
  align-items: stretch;
  border-top: 2px solid var(--border);
  overflow: hidden;
}

.article-content :deep(.md-player-btn) {
  appearance: none;
  flex: 0 0 auto;
  background: var(--bg);
  color: var(--fg);
  border: none;
  border-right: 2px solid var(--border);
  font-family: var(--font-main);
  font-weight: bold;
  font-size: 0.75rem;
  padding: 0.6rem 0.7rem;
  cursor: pointer;
  text-transform: uppercase;
  min-width: auto;
  -webkit-tap-highlight-color: transparent;
}

.article-content :deep(.md-player-btn:active) {
  background: var(--fg);
  color: var(--bg);
}

.article-content :deep(.md-player-track) {
  flex: 1 1 0;
  min-width: 0;
  position: relative;
  cursor: ew-resize;
  background: var(--bg);
  border-right: 2px solid var(--border);
  min-height: 44px;
}

.article-content :deep(.md-player-progress) {
  height: 100%;
  width: 0%;
  background: var(--fg);
  border-right: 2px solid var(--border);
}

.article-content :deep(.md-player-time) {
  flex: 0 0 auto;
  padding: 0.6rem 0.35rem;
  font-weight: bold;
  font-size: 0.6rem;
  font-family: var(--font-main);
  min-width: auto;
  text-align: center;
  display: flex;
  align-items: center;
  justify-content: center;
}

/* video: native controls, full width */
.article-content :deep(.md-video.md-mobile video) {
  display: block;
  width: 100%;
  height: auto;
  max-height: 50vh;
}

/* file: compact download bar */
.article-content :deep(.md-file.md-mobile) {
  border: 2px solid var(--border);
}

.article-content :deep(.md-file-link) {
  display: block;
  text-decoration: none;
  color: var(--fg);
}

.article-content :deep(.md-file-link:active) {
  background: var(--fg);
  color: var(--bg);
}

.article-content :deep(.md-image.md-mobile img) {
  cursor: pointer;
}
</style>
