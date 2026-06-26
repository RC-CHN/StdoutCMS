<script setup lang="ts">
import { computed, ref, onMounted, onUnmounted, watch, nextTick } from 'vue'
import { parseMarkdown } from '../utils/md'

const props = defineProps<{
  content: string
}>()

const html = computed(() => parseMarkdown(props.content))

// lightbox
const lbOpen = ref(false)
const lbSrc = ref('')
const lbAlt = ref('')

function openLightbox(img: HTMLImageElement) {
  lbSrc.value = img.src
  lbAlt.value = img.alt || 'image'
  lbOpen.value = true
}

function closeLightbox() {
  lbOpen.value = false
}

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape') closeLightbox()
}

function onClick(e: MouseEvent) {
  const target = e.target as HTMLElement
  const img = target.closest('.md-image img') as HTMLImageElement | null
  if (img) openLightbox(img)
}

onMounted(() => window.addEventListener('keydown', onKeydown))
onUnmounted(() => window.removeEventListener('keydown', onKeydown))

// ---- custom media player ----

function fmtTime(s: number): string {
  if (!isFinite(s)) return '--:--'
  const m = Math.floor(s / 60)
  const sec = Math.floor(s % 60)
  return `${m}:${sec.toString().padStart(2, '0')}`
}

function initPlayers(root: HTMLElement) {
  root.querySelectorAll<HTMLElement>('.md-media[data-player]').forEach(player => {
    if ((player as any).__inited) return
    ;(player as any).__inited = true

    const src = player.querySelector<HTMLMediaElement>('.md-player-src')
    const btn = player.querySelector<HTMLElement>('[data-play]')
    const track = player.querySelector<HTMLElement>('[data-seek]')
    const bar = track?.querySelector<HTMLElement>('.md-player-progress')
    const timeEl = player.querySelector<HTMLElement>('.md-player-time')
    const muteLabel = player.querySelector<HTMLElement>('[data-mute]')

    if (!src || !btn || !track || !bar || !timeEl || !muteLabel) return

    let duration = NaN
    let lastVol = 1

    src.addEventListener('loadedmetadata', () => {
      duration = src.duration
      timeEl.textContent = `0:00 / ${fmtTime(duration)}`
      muteLabel.textContent = `VOL: ${Math.round(src.volume * 100)}%`
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
    })

    src.addEventListener('play', () => { btn.textContent = 'PAUSE' })
    src.addEventListener('pause', () => { btn.textContent = 'PLAY' })

    src.addEventListener('volumechange', () => {
      muteLabel.textContent = src.muted ? 'MUTE' : `VOL: ${Math.round(src.volume * 100)}%`
    })

    btn.addEventListener('click', () => {
      if (src.paused) src.play()
      else src.pause()
    })

    // ---- seek drag (mouse + touch) ----
    let seeking = false

    function seekFromX(clientX: number) {
      if (!src) return
      const rect = track!.getBoundingClientRect()
      src.currentTime = Math.max(0, Math.min(1, (clientX - rect.left) / rect.width)) * duration
    }

    track!.addEventListener('mousedown', (e) => {
      seeking = true
      seekFromX(e.clientX)

      const onMove = (e: MouseEvent) => {
        if (!seeking) return
        seekFromX(e.clientX)
      }
      const onUp = () => { seeking = false; cleanup() }
      const cleanup = () => {
        document.removeEventListener('mousemove', onMove)
        document.removeEventListener('mouseup', onUp)
      }
      document.addEventListener('mousemove', onMove)
      document.addEventListener('mouseup', onUp)
    })

    track!.addEventListener('touchstart', (e) => {
      seeking = true
      seekFromX(e.touches[0].clientX)

      const onMove = (e: TouchEvent) => {
        if (!seeking) return
        e.preventDefault()
        seekFromX(e.touches[0].clientX)
      }
      const onEnd = () => { seeking = false; cleanup() }
      const cleanup = () => {
        document.removeEventListener('touchmove', onMove)
        document.removeEventListener('touchend', onEnd)
        document.removeEventListener('touchcancel', onEnd)
      }
      document.addEventListener('touchmove', onMove, { passive: false })
      document.addEventListener('touchend', onEnd)
      document.addEventListener('touchcancel', onEnd)
    }, { passive: true })

    // click fallback
    track!.addEventListener('click', (e) => {
      if (isNaN(duration)) return
      seekFromX((e as MouseEvent).clientX)
    })

    // ---- mute toggle ----
    muteLabel.addEventListener('click', () => {
      if (src.muted) { src.muted = false; src.volume = lastVol || 0.5 }
      else { lastVol = src.volume; src.muted = true }
    })

    // ---- fullscreen toggle (video only) —— fullscreen the <video> natively
    const fsBtn = player.querySelector<HTMLElement>('[data-fullscreen]')
    if (fsBtn) {
      fsBtn.addEventListener('click', () => {
        if (document.fullscreenElement) {
          document.exitFullscreen()
        } else if (src instanceof HTMLVideoElement) {
          src.requestFullscreen()
        }
      })
    }
  })
}

// ---- collapsible media ----

function initCollapsibles(root: HTMLElement) {
  // toggle buttons: [+] REVEAL / [-] HIDE
  root.querySelectorAll<HTMLElement>('.md-toggle-btn').forEach(btn => {
    if ((btn as any).__toggleInited) return
    ;(btn as any).__toggleInited = true

    btn.addEventListener('click', () => {
      const fig = btn.closest('.md-media') as HTMLElement | null
      if (!fig) return
      fig.classList.toggle('is-expanded')
      btn.textContent = fig.classList.contains('is-expanded') ? '[-] HIDE' : '[+] REVEAL'
    })
  })

  // images: read dimensions and update DIM span
  root.querySelectorAll<HTMLElement>('.md-image img').forEach(img => {
    const imgEl = img as HTMLImageElement
    const dims = img.closest('.md-image')?.querySelector('.md-img-dims')
    if (!dims) return
    const set = () => { dims.textContent = `DIM: ${imgEl.naturalWidth}x${imgEl.naturalHeight}` }
    if (imgEl.complete && imgEl.naturalWidth) set()
    else imgEl.addEventListener('load', set, { once: true })
  })

  // audio: update format + duration + STATUS in meta bar
  root.querySelectorAll<HTMLElement>('.md-audio .md-player-src').forEach(el => {
    const audio = el as HTMLMediaElement
    const info = el.closest('.md-audio')?.querySelector('.md-audio-info')
    const status = el.closest('.md-audio')?.querySelector('.md-status')
    if (!status) return

    audio.addEventListener('loadedmetadata', () => {
      if (info) {
        const ext = (audio.src.split('.').pop() || '').toUpperCase()
        info.textContent = `${ext} · ${fmtTime(audio.duration)}`
      }
    }, { once: true })

    audio.addEventListener('play',  () => { status.textContent = 'STATUS: PLAYING' })
    audio.addEventListener('pause', () => { status.textContent = 'STATUS: PAUSED' })
    audio.addEventListener('ended', () => { status.textContent = 'STATUS: DONE' })
  })

  // video: update RES when metadata loads
  root.querySelectorAll<HTMLElement>('.md-video .md-player-src').forEach(el => {
    const video = el as HTMLVideoElement
    const res = el.closest('.md-video')?.querySelector('.md-res')
    if (!res) return
    video.addEventListener('loadedmetadata', () => {
      res.textContent = `RES: ${video.videoWidth}x${video.videoHeight}`
    }, { once: true })
  })
}

watch(html, () => nextTick(() => {
  const el = document.querySelector('.article-content') as HTMLElement | null
  if (el) { initPlayers(el); initCollapsibles(el) }
}))

onMounted(() => nextTick(() => {
  const el = document.querySelector('.article-content') as HTMLElement | null
  if (el) { initPlayers(el); initCollapsibles(el) }
}))
</script>

<template>
  <div class="article-content" v-html="html" @click="onClick" />

  <teleport to="body">
    <div v-if="lbOpen" class="lb-backdrop" @click.self="closeLightbox">
      <div class="lb-frame">
        <div class="lb-bar">
          <span>> viewing: {{ lbAlt }}</span>
          <button class="lb-close" @click="closeLightbox">[X]</button>
        </div>
        <img :src="lbSrc" :alt="lbAlt" class="lb-image" />
      </div>
    </div>
  </teleport>
</template>

<style scoped>
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

.article-content :deep(h3) {
  margin-top: 2.5rem;
  margin-bottom: 1rem;
  color: var(--fg);
}
.article-content :deep(h3::before) {
  content: "# ";
  color: var(--muted);
}

.article-content :deep(p) {
  margin-bottom: 1.5rem;
}

.article-content :deep(a) {
  color: var(--fg);
  text-decoration: underline;
  text-decoration-style: dashed;
  text-underline-offset: 4px;
}
.article-content :deep(a:hover) {
  background-color: var(--fg);
  color: var(--bg);
  text-decoration: none;
}

.article-content :deep(ul) {
  margin-bottom: 1.5rem;
  padding-left: 2rem;
  list-style-type: square;
}
.article-content :deep(li) {
  margin-bottom: 0.5rem;
}

.article-content :deep(blockquote) {
  border-left: 4px solid var(--border);
  padding: 1rem 1rem 1rem 1.5rem;
  margin: 2rem 0;
  color: var(--muted);
  background: rgba(128, 128, 128, 0.05);
  font-style: italic;
}

.article-content :deep(code) {
  background: var(--fg);
  color: var(--bg);
  padding: 2px 6px;
  font-size: 0.9em;
  font-weight: bold;
}

.article-content :deep(pre) {
  background: var(--bg);
  border: 2px solid var(--border);
  padding: 1rem;
  overflow-x: auto;
  margin: 2rem 0;
  box-shadow: var(--shadow);
  position: relative;
}

.article-content :deep(.code-lang) {
  border-bottom: 2px dashed var(--border);
  padding-bottom: 5px;
  margin-bottom: 10px;
  font-size: 0.8rem;
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

/* ---- brutalist media components ---- */

.article-content :deep(.md-media) {
  display: block;
  margin: 2rem auto;
  max-width: 80%;
  border: 2px solid var(--border);
  box-shadow: var(--shadow);
  background: var(--bg);
}

/* meta bar: inverse fg/bg — brutalist header */
.article-content :deep(.md-media-meta) {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  border-bottom: 2px solid var(--border);
  padding: 0.4rem 0.6rem;
  font-size: 0.75rem;
  font-weight: bold;
  background: var(--fg);
  color: var(--bg);
  text-transform: uppercase;
  font-family: var(--font-main);
}

.article-content :deep(.md-meta-info) {
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
}

/* toggle button: [+] REVEAL / [-] HIDE */
.article-content :deep(.md-toggle-btn) {
  appearance: none;
  flex-shrink: 0;
  margin-left: 0.75rem;
  border: 2px solid var(--bg);
  background: transparent;
  color: var(--bg);
  font-family: var(--font-main);
  font-weight: bold;
  font-size: 0.7rem;
  padding: 2px 6px;
  cursor: url('/win-95-98/hand.cur'), pointer;
}

.article-content :deep(.md-toggle-btn:hover) {
  background: var(--bg);
  color: var(--fg);
}

/* audio: filename + info (inverse bar) */
.article-content :deep(.md-audio-name) {
  color: var(--bg);
  text-transform: none;
}

.article-content :deep(.md-audio-info) {
  color: var(--bg);
  opacity: 0.7;
  font-size: 0.7rem;
  text-transform: uppercase;
}

.article-content :deep(.md-status) {
  color: var(--bg);
  opacity: 0.7;
  font-size: 0.7rem;
}

/* ---- image (collapsible) ---- */

.article-content :deep(.md-image) {
  max-width: 300px;
}

.article-content :deep(.md-image.is-expanded) {
  max-width: 80%;
}

.article-content :deep(.md-image-body) {
  display: block;
}

.article-content :deep(.md-image img) {
  display: block;
  width: 100%;
  height: 150px;
  object-fit: cover;
  filter: grayscale(1) contrast(1.2);
}

.article-content :deep(.md-image.is-expanded img) {
  height: auto;
}

.article-content :deep(.md-image.is-expanded img),
.article-content :deep(.md-image img:hover) {
  filter: none;
}

.article-content :deep(.md-image-caption) {
  display: none;
  padding: 0.5rem 0.6rem;
  font-size: 0.8rem;
  border-top: 2px solid var(--border);
  background: var(--bg);
  color: var(--fg);
  font-family: var(--font-main);
}

.article-content :deep(.md-image.is-expanded .md-image-caption) {
  display: block;
}

/* ---- foldable content (audio / video) ---- */

.article-content :deep(.md-foldable-content) {
  display: none;
}

.article-content :deep(.md-media.is-expanded .md-foldable-content) {
  display: block;
}

/* ---- file download ---- */

.article-content :deep(.md-file-link) {
  display: block;
  text-decoration: none;
  color: var(--fg);
}

.article-content :deep(.md-file-link:hover) {
  background: var(--fg);
  color: var(--bg);
}

.article-content :deep(.md-file-link:hover .md-media-meta) {
  border-bottom-color: var(--bg);
}

.article-content :deep(.md-file-body) {
  padding: 1.5rem;
  font-family: var(--font-main);
}

.article-content :deep(.md-file-name) {
  font-size: 1.8rem;
  font-weight: 900;
  margin: 0 0 0.8rem 0;
  word-break: break-all;
  line-height: 1;
}

/* ---- player controls ---- */

.article-content :deep(.md-player-src) {
  display: none;  /* hidden — controlled via custom UI */
}

.article-content :deep(.md-video .md-player-src) {
  display: block;
  width: 100%;
  height: auto;
  filter: grayscale(1);
}

.article-content :deep(.md-video .md-player-src:hover),
.article-content :deep(.md-video.is-expanded .md-player-src) {
  filter: none;
}

.article-content :deep(.md-player-controls) {
  display: flex;
  border-top: 2px solid var(--border);
}

/* button: hard cut, no transition */
.article-content :deep(.md-player-btn) {
  appearance: none;
  flex-shrink: 0;
  background: var(--bg);
  color: var(--fg);
  border: none;
  border-right: 2px solid var(--border);
  font-family: var(--font-main);
  font-weight: bold;
  font-size: 0.75rem;
  padding: 0.6rem 0.8rem;
  cursor: url('/win-95-98/hand.cur'), pointer;
  text-transform: uppercase;
  text-align: center;
  min-width: 60px;
}

.article-content :deep(.md-player-btn:hover) {
  background: var(--fg);
  color: var(--bg);
}

.article-content :deep(.md-player-btn:active) {
  background: var(--bg);
  color: var(--fg);
}

/* volume button: fixed width so VOL:100% ↔ MUTE doesn't jitter */
.article-content :deep(.md-player-btn[data-mute]) {
  width: 90px;
  overflow: hidden;
  white-space: nowrap;
}

/* fullscreen button: last in row, no right border */
.article-content :deep(.md-player-btn[data-fullscreen]) {
  border-right: none;
  min-width: 44px;
  width: 44px;
}

/* progress / volume track */
.article-content :deep(.md-player-track) {
  flex-grow: 1;
  position: relative;
  cursor: ew-resize;
  background: var(--bg);
  border-right: 2px solid var(--border);
  height: auto;
  min-height: 40px;
}

.article-content :deep(.md-player-progress) {
  height: 100%;
  width: 0%;
  background: var(--fg);
  border-right: 2px solid var(--border);
}

/* time display */
.article-content :deep(.md-player-time) {
  flex-shrink: 0;
  padding: 0.6rem 0.8rem;
  font-weight: bold;
  font-size: 0.75rem;
  font-family: var(--font-main);
  border-left: 2px solid var(--border);
  border-right: 2px solid var(--border);
  min-width: 120px;
  text-align: center;
}

/* ---- lightbox ---- */

.lb-backdrop {
  position: fixed;
  inset: 0;
  z-index: 9999;
  background: rgba(0, 0, 0, 0.85);
  display: flex;
  align-items: center;
  justify-content: center;
  animation: lb-fade-in 0.15s ease;
}

@keyframes lb-fade-in {
  from { opacity: 0; }
  to { opacity: 1; }
}

.lb-frame {
  max-width: 90vw;
  max-height: 90vh;
  border: 2px solid var(--fg);
  box-shadow: 8px 8px 0px var(--fg);
  background: var(--bg);
  display: flex;
  flex-direction: column;
}

.lb-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 6px 12px;
  border-bottom: 2px dashed var(--border);
  font-size: 0.8rem;
  font-weight: bold;
  color: var(--muted);
}

.lb-close {
  background: none;
  border: none;
  color: var(--fg);
  font-family: var(--font-main);
  font-weight: bold;
  cursor: url('/win-95-98/hand.cur'), pointer;
  padding: 2px 6px;
}

.lb-close:hover {
  background: var(--fg);
  color: var(--bg);
}

.lb-image {
  display: block;
  max-width: 90vw;
  max-height: calc(90vh - 40px);
  object-fit: contain;
  filter: none;
}
</style>
