import { ref } from 'vue'

export function useToast(ttl = 2600) {
  const toast = ref<string | null>(null)

  function show(msg: string): void {
    toast.value = msg
    setTimeout(() => {
      toast.value = null
    }, ttl)
  }

  return { toast, show }
}
