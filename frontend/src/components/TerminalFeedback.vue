<script setup lang="ts">
import { watch, ref, onUnmounted } from 'vue'

export type FeedbackType = 'ok' | 'err' | 'info'

const props = withDefaults(defineProps<{
  message: string | null
  type?: FeedbackType
  duration?: number // ms, 0 = stay forever
  trigger?: number  // increment to force re-show same message
}>(), {
  type: 'info',
  duration: 4000,
  trigger: 0,
})

const visible = ref(false)
let timer: ReturnType<typeof setTimeout> | null = null

function clearTimer() {
  if (timer) { clearTimeout(timer); timer = null }
}

watch([() => props.message, () => props.trigger], ([msg]) => {
  clearTimer()
  if (msg) {
    visible.value = true
    if (props.duration > 0) {
      timer = setTimeout(() => { visible.value = false }, props.duration)
    }
  } else {
    visible.value = false
  }
})

onUnmounted(clearTimer)

const prefix = () => {
  switch (props.type) {
    case 'ok':   return '[ OK ]'
    case 'err':  return '[FAIL]'
    default:     return '[...]'
  }
}
</script>

<template>
  <Transition name="term-fade">
    <div v-if="visible && message" class="term-feedback" :class="'term-' + type">
      <span class="term-prefix">{{ prefix() }}</span>
      <span class="term-msg">{{ message }}</span>
      <span class="term-cursor">_</span>
    </div>
  </Transition>
</template>

<style scoped>
.term-feedback {
  font-family: var(--font-main);
  font-size: 0.82rem;
  padding: 4px 0;
  display: flex;
  align-items: baseline;
  gap: 6px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.term-prefix {
  font-weight: bold;
  flex-shrink: 0;
}

.term-ok .term-prefix { color: #00cc66; }
.term-err .term-prefix { color: #ff4444; }
.term-info .term-prefix { color: var(--muted); }

.term-msg {
  color: var(--fg);
  overflow: hidden;
  text-overflow: ellipsis;
}

.term-err .term-msg { color: #ff6666; }

.term-cursor {
  color: var(--fg);
  animation: cursor-blink 1s step-end infinite;
}

@keyframes cursor-blink {
  0%, 100% { opacity: 1; }
  50% { opacity: 0; }
}

/* Transition */
.term-fade-enter-active { animation: term-slide-in 0.25s ease-out; }
.term-fade-leave-active { animation: term-slide-in 0.2s ease-in reverse; }

@keyframes term-slide-in {
  from { opacity: 0; transform: translateY(-4px); }
  to   { opacity: 1; transform: translateY(0); }
}
</style>
