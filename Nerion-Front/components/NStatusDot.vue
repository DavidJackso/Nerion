<script setup lang="ts">
type Status = 'online' | 'offline' | 'setup'

defineProps<{
  status?: Status
}>()

const colors: Record<Status, string> = {
  online: 'var(--green-500)',
  offline: 'var(--red-500)',
  setup: 'var(--amber-500)',
}

const labels: Record<Status, string> = {
  online: 'Работает',
  offline: 'Недоступен',
  setup: 'Настройка',
}
</script>

<template>
  <span style="display: inline-flex; align-items: center; gap: 6px; font-size: 12px; color: var(--fg-2)">
    <span :style="{
      width: '6px', height: '6px', borderRadius: '50%',
      background: colors[status ?? 'online'],
      boxShadow: (status ?? 'online') === 'online' ? `0 0 0 3px ${colors.online}33` : 'none',
    }"/>
    {{ labels[status ?? 'online'] }}
  </span>
</template>
