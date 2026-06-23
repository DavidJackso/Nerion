<script setup lang="ts">
defineProps<{
  method: string
}>()

const colors: Record<string, string> = {
  GET: '#2563EB',
  POST: '#10A37F',
  PATCH: '#D97706',
  PUT: '#D97706',
  DELETE: '#DC2626',
}
</script>

<template>
  <span :style="{
    display: 'inline-block',
    minWidth: '52px',
    textAlign: 'center',
    padding: '3px 8px',
    borderRadius: '4px',
    background: colors[method] || '#888',
    color: '#fff',
    fontFamily: 'var(--font-mono)',
    fontSize: '11px',
    fontWeight: 700,
  }">{{ method === 'DELETE' ? 'DEL' : method }}</span>
</template>
