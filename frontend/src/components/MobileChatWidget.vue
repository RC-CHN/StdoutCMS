<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue'
import { useRoute } from 'vue-router'
import { useChat } from '../composables/useChat'
import { parseMarkdown } from '../utils/md'

const route = useRoute()
const { messages, input, loading, thinking, error, canSend, send, clear } = useChat()

const isOpen = ref(false)
const scrollRef = ref<HTMLDivElement>()
const inputRef = ref<HTMLTextAreaElement>()

const ctx = computed(() => {
  if (route.name === 'article') return route.params.slug as string
  return undefined
})

function toggle() {
  isOpen.value = !isOpen.value
  if (isOpen.value) {
    nextTick(() => inputRef.value?.focus())
  }
}

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

watch([messages, loading], () => {
  nextTick(() => {
    scrollRef.value?.scrollTo({ top: scrollRef.value.scrollHeight, behavior: 'smooth' })
  })
}, { deep: true })
</script>

<template>
  <div class="mobile-chat">
    <!-- collapsed: tap to expand -->
    <button v-if="!isOpen" class="mobile-chat-trigger" @click="toggle">
      <span>[ &gt;_ ASK_AI ]</span>
      <span v-if="ctx" class="mobile-chat-ctx">about this article</span>
    </button>

    <!-- expanded: full chat panel -->
    <div v-else class="mobile-chat-panel">
      <div class="mobile-chat-header">
        <span>[ AI_ASSISTANT ]</span>
        <div class="mobile-chat-actions">
          <button class="mobile-chat-action" @click="clear">CLR</button>
          <button class="mobile-chat-action" @click="toggle">CLOSE</button>
        </div>
      </div>

      <div ref="scrollRef" class="mobile-chat-messages">
        <div v-if="messages.length === 0" class="mobile-chat-empty">
          [STDOUT_CMS_ELF] ready.<br>
          Ask me anything<span v-if="ctx"> about this article</span>.
        </div>

        <div
          v-for="(msg, i) in messages"
          :key="i"
          :class="['mobile-chat-msg', msg.role]"
        >
          <div class="mobile-chat-prefix">
            {{ msg.role === 'user' ? '>' : '[AI]' }}
          </div>
          <div
            v-if="msg.role === 'assistant' && !msg.content && loading && i === messages.length - 1"
            class="mobile-chat-body"
          >
            <span v-if="thinking" class="mobile-thinking">thinking...</span>
            <span v-else class="mobile-cursor">█</span>
          </div>
          <div v-else class="mobile-chat-body" v-html="parseMarkdown(msg.content, true)" />
        </div>

        <div v-if="error" class="mobile-chat-error">
          [ERROR: {{ error }}]
        </div>
      </div>

      <div class="mobile-chat-input-area">
        <textarea
          ref="inputRef"
          v-model="input"
          rows="1"
          class="mobile-chat-input"
          placeholder="type your question..."
          @keydown="handleKeydown"
        />
        <button
          class="mobile-chat-send"
          :disabled="!canSend"
          @click="handleSend"
        >
          [ ENTER ]
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.mobile-chat {
  margin-top: 2rem;
  width: 100%;
}

/* ---- collapsed trigger ---- */
.mobile-chat-trigger {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  padding: 0.8rem 1rem;
  border: 2px solid var(--border);
  background: var(--bg);
  color: var(--fg);
  font-family: var(--font-main);
  font-weight: bold;
  font-size: 0.85rem;
  cursor: pointer;
  -webkit-tap-highlight-color: transparent;
}

.mobile-chat-trigger:active {
  background: var(--fg);
  color: var(--bg);
}

.mobile-chat-ctx {
  font-size: 0.7rem;
  font-weight: normal;
  color: var(--muted);
}

/* ---- expanded panel ---- */
.mobile-chat-panel {
  border: 2px solid var(--border);
  background: var(--bg);
  display: flex;
  flex-direction: column;
}

/* ---- header ---- */
.mobile-chat-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.6rem 0.8rem;
  border-bottom: 2px solid var(--border);
  background: var(--fg);
  color: var(--bg);
  font-weight: bold;
  font-size: 0.75rem;
  font-family: var(--font-main);
  flex-shrink: 0;
}

.mobile-chat-actions {
  display: flex;
  gap: 12px;
}

.mobile-chat-action {
  background: none;
  border: none;
  color: var(--bg);
  font-family: var(--font-main);
  font-size: 0.7rem;
  font-weight: bold;
  cursor: pointer;
  padding: 2px 4px;
  -webkit-tap-highlight-color: transparent;
}

.mobile-chat-action:active {
  opacity: 0.6;
}

/* ---- messages ---- */
.mobile-chat-messages {
  overflow-y: auto;
  padding: 0.8rem;
  font-size: 0.85rem;
  line-height: 1.6;
  max-height: 50vh;
  -webkit-overflow-scrolling: touch;
}

.mobile-chat-empty {
  color: var(--muted);
  font-size: 0.8rem;
  text-align: center;
  margin-top: 1.5rem;
}

.mobile-chat-msg {
  margin-bottom: 1rem;
}

.mobile-chat-msg.user {
  text-align: right;
}

.mobile-chat-msg.user .mobile-chat-body {
  display: inline-block;
  text-align: left;
  background: rgba(128, 128, 128, 0.08);
  padding: 0.4rem 0.6rem;
  border: 1px solid var(--border);
  max-width: 85%;
}

.mobile-chat-prefix {
  font-size: 0.68rem;
  color: var(--muted);
  margin-bottom: 2px;
}

.mobile-chat-msg.user .mobile-chat-prefix {
  color: #3b82f6;
}

.mobile-chat-body :deep(p) {
  margin: 0.3rem 0;
}

.mobile-chat-body :deep(p:first-child) {
  margin-top: 0;
}

.mobile-chat-body :deep(pre) {
  background: rgba(128, 128, 128, 0.08);
  padding: 8px;
  border: 1px dashed var(--border);
  overflow-x: auto;
  font-size: 0.75rem;
  margin: 4px 0;
}

.mobile-chat-body :deep(code) {
  background: rgba(128, 128, 128, 0.08);
  padding: 1px 3px;
  font-size: 0.85em;
}

.mobile-chat-body :deep(a) {
  color: var(--fg);
  text-decoration: underline;
}

.mobile-chat-body :deep(blockquote) {
  border-left: 2px solid var(--border);
  padding-left: 8px;
  margin: 4px 0;
  color: var(--muted);
}

.mobile-chat-error {
  color: #ef4444;
  font-size: 0.78rem;
  padding: 6px 8px;
  border: 1px dashed #ef4444;
  margin-top: 6px;
}

/* ---- input area ---- */
.mobile-chat-input-area {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 0.6rem 0.8rem;
  border-top: 2px solid var(--border);
  flex-shrink: 0;
  background: var(--bg);
}

.mobile-chat-input {
  width: 100%;
  box-sizing: border-box;
  background: transparent;
  border: 1px solid var(--border);
  outline: none;
  color: var(--fg);
  font-family: var(--font-main);
  font-size: 0.9rem;
  resize: none;
  min-height: 44px;
  max-height: 120px;
  field-sizing: content;
  line-height: 1.5;
  padding: 0.55rem 0.7rem;
  -webkit-appearance: none;
  border-radius: 0;
}

.mobile-chat-input:focus {
  border-color: var(--fg);
}

.mobile-chat-input::placeholder {
  color: var(--muted);
  opacity: 0.5;
}

.mobile-chat-send {
  width: 100%;
  box-sizing: border-box;
  font-family: var(--font-main);
  font-weight: bold;
  font-size: 0.8rem;
  padding: 0.6rem 0.8rem;
  cursor: pointer;
  background: var(--bg);
  color: var(--fg);
  border: 2px solid var(--border);
  -webkit-tap-highlight-color: transparent;
  min-height: 40px;
  text-align: center;
}

.mobile-chat-send:active:not(:disabled) {
  background: var(--fg);
  color: var(--bg);
}

.mobile-chat-send:disabled {
  opacity: 0.35;
}

/* ---- indicators ---- */
.mobile-cursor {
  animation: mobile-blink 1s step-end infinite;
}

@keyframes mobile-blink {
  50% { opacity: 0; }
}

.mobile-thinking {
  color: var(--muted);
  font-style: italic;
  animation: mobile-think-pulse 1.5s ease-in-out infinite;
}

@keyframes mobile-think-pulse {
  0%, 100% { opacity: 0.4; }
  50% { opacity: 1; }
}
</style>
