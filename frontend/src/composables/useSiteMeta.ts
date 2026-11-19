import { onMounted, shallowRef } from 'vue'
import { fetchMeta, type MetaResponse } from '../api/meta'

const meta = shallowRef<MetaResponse | null>(null)
let loading = false

export function useSiteMeta() {
  onMounted(async () => {
    if (meta.value || loading) return
    loading = true
    try {
      meta.value = await fetchMeta()
    } catch {
      // Metadata is optional. The next consumer can retry.
    } finally {
      loading = false
    }
  })
  return meta
}
