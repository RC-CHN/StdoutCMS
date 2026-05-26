<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue'
import { useRoute } from 'vue-router'
import { useChat } from '../composables/useChat'
import { parseMarkdown } from '../utils/md'

const route = useRoute()
const { messages, input, loading, error, canSend, send, clear } = useChat()

const isOpen = ref(false)
const scrollRef = ref<HTMLDivElement>()
const inputRef = ref<HTMLTextAreaElement>()

// 文章页自动获取 ctx
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

// 自动滚动到底部
watch([messages, loading], () => {
  nextTick(() => {
    scrollRef.value?.scrollTo({ top: scrollRef.value.scrollHeight, behavior: 'smooth' })
  })
}, { deep: true })
</script>

<template>
  <!-- 收起状态按钮 -->
  <button
    v-if="!isOpen"
    class="chat-trigger"
    @click="toggle"
    title="AI Assistant"
  >
    [&gt;_]
  </button>

  <!-- 聊天窗口 -->
  <div v-else class="chat-window">
    <!-- 标题栏 -->
    <div class="chat-header">
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
        <div class="chat-msg-body" v-html="parseMarkdown(msg.content)" />
      </div>

      <!-- 加载中 -->
      <div v-if="loading" class="chat-msg assistant">
        <div class="chat-msg-prefix">[STDOUT_CMS_ELF]</div>
        <div class="chat-msg-body">
          <span class="typing-cursor">█</span>
        </div>
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
        class="chat-send"
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
  bottom: 24px;
  right: 24px;
  width: 48px;
  height: 48px;
  border: 2px solid var(--border);
  background: var(--bg);
  box-shadow: var(--shadow);
  font-family: var(--font-main);
  font-size: 0.9rem;
  font-weight: bold;
  color: var(--fg);
  cursor: pointer;
  z-index: 100;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.15s;
}
.chat-trigger:hover {
  background: var(--fg);
  color: var(--bg);
  transform: translate(-2px, -2px);
}

/* ---- 聊天窗口 ---- */
.chat-window {
  position: fixed;
  bottom: 24px;
  right: 24px;
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
  cursor: pointer;
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
  align-items: flex-end;
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
  padding-bottom: 4px;
  user-select: none;
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
}
.chat-input::placeholder {
  color: var(--muted);
  opacity: 0.5;
}
.chat-send {
  background: none;
  border: none;
  color: var(--fg);
  font-family: var(--font-main);
  font-size: 0.75rem;
  cursor: pointer;
  padding: 4px 0;
  white-space: nowrap;
  opacity: 1;
}
.chat-send:disabled {
  opacity: 0.3;
  cursor: default;
}

/* ---- 打字机光标 ---- */
.typing-cursor {
  animation: blink 1s step-end infinite;
}
@keyframes blink {
  50% { opacity: 0; }
}

/* ---- 响应式 ---- */
@media (max-width: 600px) {
  .chat-window {
    width: calc(100vw - 32px);
    height: 65vh;
    right: 16px;
    bottom: 16px;
  }
  .chat-trigger {
    bottom: 16px;
    right: 16px;
  }
}
</style>
