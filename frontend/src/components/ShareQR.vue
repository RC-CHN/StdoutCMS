<script setup lang="ts">
import { ref, computed, watch, onMounted, nextTick } from 'vue'
import QRCode from 'qrcode'
import html2canvas from 'html2canvas'

const props = defineProps<{
  slug: string
  title: string
  tags?: string[]
  readTime?: string
  wordCount?: number
  createdAt?: string
}>()

const open = ref(false)
const cardRef = ref<HTMLElement | null>(null)

const url = computed(() => `${window.location.origin}/posts/${props.slug}`)
const pubDate = computed(() => props.createdAt?.slice(0, 10) || '—')
const tagsStr = computed(() => (props.tags || []).join(' / ') || '—')
const read = computed(() => {
  if (props.readTime) return props.readTime
  if (props.wordCount) {
    const charsPerMin = 400
    const secs = Math.round(props.wordCount / charsPerMin * 60)
    if (secs < 60) return `${secs}S`
    return `${Math.round(secs / 60)} MIN`
  }
  return '—'
})

let copyTimer: ReturnType<typeof setTimeout> | null = null

function openCard() { open.value = true }
function closeCard() { open.value = false }

onMounted(() => {
  document.addEventListener('keydown', (e) => { if (e.key === 'Escape') closeCard() })
})

// ---- QR canvas ----
const qrCanvas = ref<HTMLCanvasElement | null>(null)
watch(open, async (val) => {
  if (val) {
    await nextTick()
    // Small delay for canvas to mount
    setTimeout(() => {
      if (qrCanvas.value) {
        QRCode.toCanvas(qrCanvas.value, url.value, {
          width: 150,
          margin: 1,
          color: { dark: document.body.classList.contains('dark-mode') ? '#d4d4d4' : '#111111', light: document.body.classList.contains('dark-mode') ? '#1a1a1c' : '#f4f4f4' },
        })
      }
    }, 50)
  }
})

// ---- copy URL ----
const copyLabel = ref('COPY URL')

async function copyUrl() {
  await navigator.clipboard.writeText(url.value)
  copyLabel.value = 'COPIED'
  if (copyTimer) clearTimeout(copyTimer)
  copyTimer = setTimeout(() => { copyLabel.value = 'COPY URL' }, 2000)
}

// ---- save JPEG ----
const saveLabel = ref('SAVE JPEG')

async function saveJpeg() {
  if (!cardRef.value) return
  saveLabel.value = 'SAVING...'
  const footer = cardRef.value.querySelector('.qr-footer') as HTMLElement | null
  try {
    // hide buttons before capture so they don't appear in the JPEG
    if (footer) footer.style.display = 'none'
    const canvas = await html2canvas(cardRef.value, {
      backgroundColor: document.body.classList.contains('dark-mode') ? '#1a1a1c' : '#f4f4f4',
      scale: 2,
    })
    canvas.toBlob((blob) => {
      if (!blob) { saveLabel.value = 'FAILED'; return }
      const a = document.createElement('a')
      a.href = URL.createObjectURL(blob)
      a.download = `${props.slug}-card.jpg`
      a.click()
      saveLabel.value = 'SAVE JPEG'
    }, 'image/jpeg', 0.95)
  } catch {
    saveLabel.value = 'FAILED'
  } finally {
    if (footer) footer.style.display = ''
  }
}
</script>

<template>
  <button class="btn" @click="openCard">[QR]</button>

  <teleport to="body">
    <div v-if="open" class="qr-backdrop" @click.self="closeCard">
      <div ref="cardRef" class="qr-card">
        <!-- meta bar -->
        <div class="qr-meta-bar">
          <span>SLUG: {{ slug }}</span>
          <span>PUB: {{ pubDate }}</span>
        </div>

        <div class="qr-body">
          <!-- left: info -->
          <div class="qr-info">
            <h1 class="qr-title">{{ title }}</h1>

            <div class="qr-url-box">
              <div class="qr-url-label">TARGET_URL:</div>
              <a :href="url" class="qr-url-text" target="_blank">{{ url }}</a>
            </div>

            <div class="qr-stats">
              <div class="qr-stat">
                <span class="qr-stat-label">TAGS</span>
                <span class="qr-stat-value">{{ tagsStr }}</span>
              </div>
              <div class="qr-stat">
                <span class="qr-stat-label">READ</span>
                <span class="qr-stat-value">{{ read }}</span>
              </div>
              <div class="qr-stat">
                <span class="qr-stat-label">POSTED</span>
                <span class="qr-stat-value">{{ pubDate }}</span>
              </div>
            </div>
          </div>

          <!-- right: QR -->
          <div class="qr-code">
            <canvas ref="qrCanvas" width="150" height="150" />
            <div class="qr-code-label">[ SCAN TO ACCESS ]</div>
          </div>
        </div>

        <!-- footer actions -->
        <div class="qr-footer">
          <button class="qr-btn" @click="copyUrl">{{ copyLabel }}</button>
          <button class="qr-btn" @click="saveJpeg">{{ saveLabel }}</button>
        </div>
      </div>
    </div>
  </teleport>
</template>

<style scoped>
.qr-backdrop {
  position: fixed;
  inset: 0;
  z-index: 10000;
  background: rgba(0, 0, 0, 0.85);
  display: flex;
  align-items: center;
  justify-content: center;
  animation: qr-fade-in 0.15s ease;
}

@keyframes qr-fade-in {
  from { opacity: 0; }
  to { opacity: 1; }
}

.qr-card {
  width: 100%;
  max-width: 600px;
  border: 2px solid var(--border);
  box-shadow: var(--shadow);
  background: var(--bg);
  display: flex;
  flex-direction: column;
  margin: 1rem;
}

/* meta bar */
.qr-meta-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-bottom: 2px solid var(--border);
  padding: 0.5rem 0.6rem;
  font-size: 0.75rem;
  font-weight: bold;
  background: var(--fg);
  color: var(--bg);
  font-family: var(--font-main);
  text-transform: uppercase;
}

/* body */
.qr-body {
  display: flex;
  flex-wrap: wrap;
}

.qr-info {
  flex: 1 1 300px;
  display: flex;
  flex-direction: column;
  border-right: 2px solid var(--border);
}

.qr-title {
  font-family: var(--font-main);
  font-size: 1.4rem;
  font-weight: 900;
  line-height: 1.1;
  text-transform: uppercase;
  margin: 0;
  padding: 1rem;
  border-bottom: 2px solid var(--border);
  word-break: break-word;
}

.qr-url-box {
  padding: 1rem;
  border-bottom: 2px solid var(--border);
  flex-grow: 1;
}

.qr-url-label {
  font-size: 0.7rem;
  font-weight: bold;
  color: var(--muted);
  margin-bottom: 0.25rem;
}

.qr-url-text {
  font-size: 0.8rem;
  font-weight: bold;
  word-break: break-all;
  color: var(--fg);
  text-decoration: none;
}

.qr-url-text:hover {
  background: var(--fg);
  color: var(--bg);
}

/* stats */
.qr-stats {
  display: flex;
}

.qr-stat {
  flex: 1;
  padding: 0.5rem 0.6rem;
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  border-right: 2px solid var(--border);
}

.qr-stat:last-child {
  border-right: none;
}

.qr-stat-label {
  font-size: 0.65rem;
  font-weight: bold;
  text-transform: uppercase;
  color: var(--muted);
  margin-bottom: 0.15rem;
}

.qr-stat-value {
  font-size: 0.85rem;
  font-family: var(--font-main);
  font-weight: 900;
  color: var(--fg);
}

/* QR code */
.qr-code {
  flex: 0 0 180px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 1rem;
  background: var(--bg);
}

.qr-code canvas {
  display: block;
  image-rendering: pixelated;
}

.qr-code-label {
  margin-top: 0.5rem;
  font-size: 0.7rem;
  font-weight: bold;
  color: var(--muted);
  text-transform: uppercase;
}

/* footer */
.qr-footer {
  border-top: 2px solid var(--border);
  display: flex;
}

.qr-btn {
  appearance: none;
  flex: 1;
  background: var(--bg);
  color: var(--fg);
  border: none;
  border-right: 2px solid var(--border);
  font-family: var(--font-main);
  font-weight: bold;
  font-size: 0.8rem;
  padding: 0.7rem 1rem;
  cursor: url('/win-95-98/hand.cur'), pointer;
  text-transform: uppercase;
}

.qr-btn:last-child {
  border-right: none;
}

.qr-btn:hover {
  background: var(--fg);
  color: var(--bg);
}

.qr-btn:active {
  background: var(--bg);
  color: var(--fg);
}

@media (max-width: 500px) {
  .qr-info {
    border-right: none;
    border-bottom: 2px solid var(--border);
  }
  .qr-code {
    flex-direction: row;
    gap: 1rem;
  }
}
</style>
