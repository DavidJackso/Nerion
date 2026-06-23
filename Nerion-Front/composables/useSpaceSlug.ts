import { computed } from 'vue'
import type { Space } from '~/api/spaces'

export function useSpaceSlug() {
  const route = useRoute()
  const spacesStore = useSpacesStore()

  const slug = computed(() => route.params.slug as string)
  const space = computed<Space | undefined>(() =>
    spacesStore.spaces.find((s) => s.slug === slug.value)
  )

  return { slug, space }
}
