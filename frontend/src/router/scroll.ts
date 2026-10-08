import { nextTick } from 'vue'

// Wait for the route's data before restoring a position on a formerly long page.
export async function waitForPageContent() {
  await nextTick()
  await new Promise<void>(resolve => requestAnimationFrame(() => resolve()))
  const root = document.getElementById('main-content')
  if (!root?.querySelector('[aria-busy="true"]')) return
  await new Promise<void>(resolve => {
    const finish = () => { observer.disconnect(); clearTimeout(timeout); resolve() }
    const observer = new MutationObserver(() => {
      if (!root.querySelector('[aria-busy="true"]')) finish()
    })
    const timeout = setTimeout(finish, 10000)
    observer.observe(root, { childList: true, subtree: true, attributes: true, attributeFilter: ['aria-busy'] })
  })
}
