import { ref, onMounted, onUnmounted } from 'vue'

const MOBILE_BREAKPOINT = 768

const isMobile = ref(false)

function update() {
  isMobile.value = window.innerWidth <= MOBILE_BREAKPOINT
}

let listeners = 0
let resizeHandler: (() => void) | null = null

export function useBreakpoint() {
  onMounted(() => {
    if (listeners === 0) {
      update()
      resizeHandler = update
      window.addEventListener('resize', resizeHandler)
    }
    listeners++
  })

  onUnmounted(() => {
    listeners--
    if (listeners === 0 && resizeHandler) {
      window.removeEventListener('resize', resizeHandler)
      resizeHandler = null
    }
  })

  return { isMobile }
}
