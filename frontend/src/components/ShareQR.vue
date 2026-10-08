<script setup lang="ts">
import { ref, onMounted } from 'vue'
import QRCode from 'qrcode'

const props = defineProps<{ slug: string; title: string; tags?: string[]; readTime?: string; wordCount?: number; createdAt?: string }>()
const emit = defineEmits<{ close: [] }>()
const dialog = ref<HTMLDialogElement | null>(null)
const card = ref<HTMLElement | null>(null)
const canvas = ref<HTMLCanvasElement | null>(null)
const url = location.origin + '/article/' + props.slug
const status = ref('')
const saving = ref(false)
onMounted(async () => {
  dialog.value?.showModal()
  try {
    if (canvas.value) await QRCode.toCanvas(canvas.value, url, { width: 180, margin: 2, color: { dark: '#111111', light: '#ffffff' } })
  } catch { status.value = 'Could not create QR code. You can still copy the link.' }
})
function close() { dialog.value?.close(); emit('close') }
async function copy() {
  try { await navigator.clipboard.writeText(url); status.value = 'Link copied.' }
  catch { status.value = 'Copy failed. Select the link above to copy it manually.' }
}
async function save() {
  if (!card.value || saving.value) return
  saving.value = true
  try {
    const { default: html2canvas } = await import('html2canvas')
    const image = await html2canvas(card.value, { scale: 2, backgroundColor: getComputedStyle(card.value).backgroundColor })
    const blob = await new Promise<Blob | null>(resolve => image.toBlob(resolve, 'image/jpeg', 0.95))
    if (!blob) throw new Error('Export failed')
    const objectURL = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = objectURL
    link.download = props.slug + '-card.jpg'
    link.click()
    setTimeout(() => URL.revokeObjectURL(objectURL), 1000)
    status.value = 'Card saved.'
  } catch { status.value = 'Could not save the card. Try again.' }
  finally { saving.value = false }
}
</script>

<template>
  <Teleport to="body">
    <dialog ref="dialog" class="qr-dialog" aria-labelledby="share-title" @cancel.prevent="close" @click="event => { if (event.target === dialog) close() }">
      <div class="dialog-bar"><span>SHARE ARTICLE</span><button aria-label="Close share card" @click="close">×</button></div>
      <div ref="card" class="qr-card">
        <p class="qr-brand">STDOUT_CMS_ELF / ARTICLE</p>
        <h2 id="share-title">{{ title }}</h2>
        <p class="qr-meta">{{ createdAt?.slice(0, 10) }} · {{ readTime }}<span v-if="tags?.length"> · {{ tags.join(' / ') }}</span></p>
        <canvas ref="canvas" aria-label="Article QR code" />
        <a :href="url">{{ url }}</a>
      </div>
      <div class="qr-actions"><button @click="copy">COPY LINK</button><button :disabled="saving" @click="save">{{ saving ? 'SAVING…' : 'SAVE CARD' }}</button></div>
      <p class="qr-status" role="status">{{ status }}</p>
    </dialog>
  </Teleport>
</template>

<style scoped>
.qr-dialog { margin: auto; width: min(560px, calc(100% - 2rem)); max-height: calc(100dvh - 2rem); overflow-y: auto; background: var(--bg); color: var(--fg); border: 2px solid var(--border); padding: 0; }
.qr-dialog::backdrop { background: rgba(0, 0, 0, 0.7); }
.dialog-bar { display: flex; justify-content: space-between; align-items: center; border-bottom: 1px solid var(--border); padding: 0.5rem 1rem; font-size: 0.8rem; }
.dialog-bar button { min-width: 40px; min-height: 40px; }
.qr-card { padding: 1.5rem; background: var(--bg); text-align: center; overflow-wrap: anywhere; }
.qr-card h2 { font-size: 1.5rem; line-height: 1.5; margin-bottom: 1rem; }
.qr-brand, .qr-meta { color: var(--muted); font-size: 0.75rem; margin-bottom: 1rem; }
.qr-card canvas { display: block; margin: 1rem auto; }
.qr-card a { color: var(--fg); font-size: 0.8rem; }
.qr-actions { padding: 0 1rem; display: flex; justify-content: center; gap: 0.5rem; }
.qr-status { margin: 0.75rem 1rem; color: var(--muted); font-size: 0.8rem; }
</style>
