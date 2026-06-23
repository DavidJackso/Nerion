<script setup lang="ts">
defineProps<{
  modelValue?: boolean
}>()

defineEmits<{
  'update:modelValue': [value: boolean]
}>()
</script>

<template>
  <button
    @click="$emit('update:modelValue', !(modelValue ?? false))"
    :style="{
      width: '32px', height: '18px',
      borderRadius: '999px',
      background: modelValue ? 'var(--brand-primary)' : 'var(--neutral-300)',
      border: 0, padding: 0, cursor: 'pointer',
      position: 'relative',
      transition: 'background 150ms',
      flexShrink: 0,
    }"
  >
    <span :style="{
      position: 'absolute', top: '2px',
      left: modelValue ? '16px' : '2px',
      width: '14px', height: '14px',
      borderRadius: '50%',
      background: '#fff',
      transition: 'left 150ms',
      boxShadow: '0 1px 2px rgba(0,0,0,.2)',
    }"/>
  </button>
</template>
