<script setup lang="ts">
import { ref, computed, watch, nextTick, onMounted, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import { useChat } from '../composables/useChat'
import { parseMarkdown } from '../utils/md'

const route = useRoute()
const { messages, input, loading, thinking, error, canSend, send, clear } = useChat()

const isOpen = ref(false)
const scrollRef = ref<HTMLDivElement>()
const inputRef = ref<HTMLTextAreaElement>()

// 文章页自动获取 ctx
const ctx = computed(() => {
  if (route.name === 'article') return route.params.slug as string
  return undefined
})

function handleSend() {
  if (!canSend.value) return
  const q = input.value
  input.value = ''
  send(q, ctx.value)
}

function handleKeydown(e: KeyboardEvent) {
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault()
    handleSend()
  }
}

// 自动滚动到底部
watch([messages, loading], () => {
  nextTick(() => {
    scrollRef.value?.scrollTo({ top: scrollRef.value.scrollHeight, behavior: 'smooth' })
  })
}, { deep: true })

// ---- 拖动 ----
const pos = ref({ x: 0, y: 0 })
const isDragging = ref(false)
const dragOffset = ref({ x: 0, y: 0 })
const dragged = ref(false)

function clampPos(x: number, y: number) {
  const w = isOpen.value ? 420 : 48
  const h = isOpen.value ? 540 : 48
  pos.value.x = Math.max(0, Math.min(window.innerWidth - w, x))
  pos.value.y = Math.max(0, Math.min(window.innerHeight - h, y))
}

function initPos() {
  const w = isOpen.value ? 420 : 48
  const h = isOpen.value ? 540 : 48
  clampPos(window.innerWidth - w - 24, window.innerHeight - h - 24)
}

function onDragStart(e: MouseEvent) {
  isDragging.value = true
  dragged.value = false
  dragOffset.value.x = e.clientX - pos.value.x
  dragOffset.value.y = e.clientY - pos.value.y
}

function onWindowMouseMove(e: MouseEvent) {
  if (!isDragging.value) return
  const nx = e.clientX - dragOffset.value.x
  const ny = e.clientY - dragOffset.value.y
  if (Math.abs(nx - pos.value.x) > 3 || Math.abs(ny - pos.value.y) > 3) {
    dragged.value = true
  }
  clampPos(nx, ny)
}

function onWindowMouseUp() {
  isDragging.value = false
}

function onWindowResize() {
  clampPos(pos.value.x, pos.value.y)
}

function onTriggerClick() {
  if (dragged.value) return
  toggle()
}
function onTriggerMouseDown(e: MouseEvent) { onDragStart(e) }
function onHeaderMouseDown(e: MouseEvent) {
  if ((e.target as HTMLElement).closest('.chat-action')) return
  onDragStart(e)
}

// 切换展开/收起时重新确保位置不越界
function toggle() {
  isOpen.value = !isOpen.value
  nextTick(() => {
    clampPos(pos.value.x, pos.value.y)
    if (isOpen.value) {
      inputRef.value?.focus()
    }
  })
}

onMounted(() => {
  initPos()
  window.addEventListener('mousemove', onWindowMouseMove)
  window.addEventListener('mouseup', onWindowMouseUp)
  window.addEventListener('resize', onWindowResize)
})

onUnmounted(() => {
  window.removeEventListener('mousemove', onWindowMouseMove)
  window.removeEventListener('mouseup', onWindowMouseUp)
  window.removeEventListener('resize', onWindowResize)
})
</script>

<template>
  <!-- 收起状态按钮（可拖动） -->
  <button
    v-if="!isOpen"
    class="chat-trigger"
    :class="{ dragging: isDragging }"
    :style="{ left: pos.x + 'px', top: pos.y + 'px' }"
    @mousedown="onTriggerMouseDown"
    @click="onTriggerClick"
    title="AI Assistant"
  >
    [&gt;_]
  </button>

  <!-- 聊天窗口 -->
  <div
    v-else
    class="chat-window"
    :style="{ left: pos.x + 'px', top: pos.y + 'px' }"
  >
    <!-- 标题栏 — 可拖动 -->
    <div
      class="chat-header"
      :class="{ dragging: isDragging }"
      @mousedown="onHeaderMouseDown"
    >
      <span>[ AI_ASSISTANT ]</span>
      <div class="chat-header-actions">
        <button class="chat-action" @click="clear" title="clear session">[CLR]</button>
        <button class="chat-action" @click="toggle" title="close">[×]</button>
      </div>
    </div>

    <!-- 消息区 -->
    <div ref="scrollRef" class="chat-messages">
      <div v-if="messages.length === 0" class="chat-empty">
        [STDOUT_CMS_ELF] ready.<br>
        Ask me anything<span v-if="ctx"> about this article</span>.
      </div>

      <div
        v-for="(msg, i) in messages"
        :key="i"
        :class="['chat-msg', msg.role]"
      >
        <div class="chat-msg-prefix">
          {{ msg.role === 'user' ? '>' : '[STDOUT_CMS_ELF]' }}
        </div>
        <!-- 最后一个 assistant 空内容 + 加载中：思考 / 光标 -->
        <div
          v-if="msg.role === 'assistant' && !msg.content && loading && i === messages.length - 1"
          class="chat-msg-body"
        >
          <span v-if="thinking" class="thinking-indicator">thinking...</span>
          <span v-else class="typing-cursor">█</span>
        </div>
        <div v-else class="chat-msg-body" v-html="parseMarkdown(msg.content)" />
      </div>

      <!-- 错误 -->
      <div v-if="error" class="chat-error">
        [ERROR: {{ error }}]
      </div>
    </div>

    <!-- 输入区 -->
    <div class="chat-input-area">
      <span class="chat-prompt">guest@k8s-node ~ $</span>
      <textarea
        ref="inputRef"
        v-model="input"
        rows="1"
        class="chat-input"
        placeholder="type your question..."
        @keydown="handleKeydown"
      />
      <button
        class="chat-send btn"
        :disabled="!canSend"
        @click="handleSend"
      >
        [ENTER]
      </button>
    </div>
  </div>
</template>

<style scoped>
/* ---- 触发按钮 ---- */
.chat-trigger {
  position: fixed;
  width: 48px;
  height: 48px;
  border: 2px solid var(--border);
  background: var(--bg);
  box-shadow: var(--shadow);
  font-family: var(--font-main);
  font-size: 0.9rem;
  font-weight: bold;
  color: var(--fg);
  cursor: url('/win-95-98/relocate.cur'), move;
  z-index: 100;
  display: flex;
  align-items: center;
  justify-content: center;
}
.chat-trigger:hover {
  background: var(--fg);
  color: var(--bg);
}
.chat-trigger.dragging {
  background: var(--fg);
  color: var(--bg);
}

/* ---- 聊天窗口 ---- */
.chat-window {
  position: fixed;
  width: 420px;
  max-width: calc(100vw - 48px);
  height: 540px;
  max-height: calc(100vh - 48px);
  border: 2px solid var(--border);
  background: var(--bg);
  box-shadow: var(--shadow);
  display: flex;
  flex-direction: column;
  z-index: 100;
}

/* ---- 标题栏 ---- */
.chat-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 14px;
  border-bottom: 2px solid var(--border);
  background: var(--fg);
  color: var(--bg);
  font-weight: bold;
  font-size: 0.8rem;
  flex-shrink: 0;
  cursor: url('/win-95-98/relocate.cur'), move;
  user-select: none;
}
.chat-header.dragging {
  opacity: 0.9;
}
.chat-header-actions {
  display: flex;
  gap: 8px;
}
.chat-action {
  background: none;
  border: none;
  color: var(--bg);
  font-family: var(--font-main);
  font-size: 0.75rem;
  cursor: url('/win-95-98/hand.cur'), pointer;
  padding: 0;
}
.chat-action:hover {
  opacity: 0.7;
}

/* ---- 消息区 ---- */
.chat-messages {
  flex: 1;
  overflow-y: auto;
  padding: 14px;
  font-size: 0.88rem;
  line-height: 1.65;
}

.chat-empty {
  color: var(--muted);
  font-size: 0.82rem;
  text-align: center;
  margin-top: 2rem;
}

.chat-msg {
  margin-bottom: 1.2rem;
}
.chat-msg.user {
  text-align: right;
}
.chat-msg.user .chat-msg-body {
  display: inline-block;
  text-align: left;
}
.chat-msg-prefix {
  font-size: 0.72rem;
  color: var(--muted);
  margin-bottom: 3px;
  user-select: none;
}
.chat-msg.user .chat-msg-prefix {
  color: #3b82f6;
}
.chat-msg-body :deep(p) {
  margin: 0.4rem 0;
}
.chat-msg-body :deep(p:first-child) {
  margin-top: 0;
}
.chat-msg-body :deep(pre) {
  background: rgba(128, 128, 128, 0.08);
  padding: 10px;
  border: 1px dashed var(--border);
  overflow-x: auto;
  font-size: 0.78rem;
  margin: 6px 0;
  text-align: left;
}
.chat-msg-body :deep(code) {
  background: rgba(128, 128, 128, 0.08);
  padding: 1px 4px;
  font-size: 0.85em;
}
.chat-msg-body :deep(a) {
  color: var(--fg);
  text-decoration: underline;
}
.chat-msg-body :deep(blockquote) {
  border-left: 2px solid var(--border);
  padding-left: 10px;
  margin: 6px 0;
  color: var(--muted);
}

.chat-error {
  color: #ef4444;
  font-size: 0.82rem;
  padding: 8px 10px;
  border: 1px dashed #ef4444;
  margin-top: 8px;
}

/* ---- 输入区 ---- */
.chat-input-area {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 12px;
  border-top: 2px solid var(--border);
  flex-shrink: 0;
  background: var(--bg);
}
.chat-prompt {
  color: var(--muted);
  font-size: 0.8rem;
  white-space: nowrap;
  user-select: none;
  line-height: 1.5;
}
.chat-input {
  flex: 1;
  background: transparent;
  border: none;
  outline: none;
  color: var(--fg);
  font-family: var(--font-main);
  font-size: 0.88rem;
  resize: none;
  min-height: 20px;
  max-height: 120px;
  field-sizing: content;
  line-height: 1.5;
  padding: 0;
  margin: 0;
}
.chat-input::placeholder {
  color: var(--muted);
  opacity: 0.5;
}

/* ---- 发送按钮 ---- */
.chat-send {
  font-family: var(--font-main);
  font-size: 0.75rem;
  padding: 4px 10px;
  white-space: nowrap;
  cursor: url('/win-95-98/hand.cur'), pointer;
  background: var(--bg);
  color: var(--fg);
  border: 2px solid var(--border);
  box-shadow: 2px 2px 0px var(--border);
  transition: all 0.1s;
}
.chat-send:hover:not(:disabled) {
  background: var(--fg);
  color: var(--bg);
  box-shadow: 3px 3px 0px var(--border);
  transform: translate(-1px, -1px);
}
.chat-send:disabled {
  opacity: 0.35;
  cursor: default;
  box-shadow: none;
  transform: none;
}

/* ---- 打字机光标 ---- */
.typing-cursor {
  animation: blink 1s step-end infinite;
}
@keyframes blink {
  50% { opacity: 0; }
}

/* ---- 思考中 ---- */
.thinking-indicator {
  color: var(--muted);
  font-style: italic;
  animation: think-pulse 1.5s ease-in-out infinite;
}
@keyframes think-pulse {
  0%, 100% { opacity: 0.4; }
  50% { opacity: 1; }
}

/* ---- 响应式 ---- */
@media (max-width: 600px) {
  .chat-window {
    width: calc(100vw - 32px);
    height: 65vh;
  }
  .chat-trigger {
    width: 42px;
    height: 42px;
  }
}
</style>
