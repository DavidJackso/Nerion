<script setup lang="ts">
import { ref } from 'vue'

defineProps<{
  modelValue?: string
  placeholder?: string
  type?: string
  error?: string
  mono?: boolean
  disabled?: boolean
}>()

defineEmits<{
  'update:modelValue': [value: string]
}>()

const focus = ref(false)
</script>

<template>
  <div style="display: flex; flex-direction: column; gap: 4px; width: 100%">
    <input
      :type="type ?? 'text'"
      :value="modelValue ?? ''"
      :placeholder="placeholder ?? ''"
      :disabled="disabled"
      :style="{
        height: '36px',
        padding: '0 12px',
        borderRadius: '6px',
        border: `0.5px solid ${error ? 'var(--danger-solid)' : focus ? 'var(--border-focus)' : 'var(--border-strong)'}`,
        boxShadow: focus ? 'var(--ring-focus)' : 'none',
        background: 'var(--bg-0)',
        color: 'var(--fg-1)',
        fontFamily: mono ? 'var(--font-mono)' : 'var(--font-sans)',
        fontSize: '14px',
        outline: 0,
        width: '100%',
        boxSizing: 'border-box',
        transition: 'border-color 120ms, box-shadow 120ms',
      }"
      @input="$emit('update:modelValue', ($event.target as HTMLInputElement).value)"
      @focus="focus = true"
      @blur="focus = false"
    />
    <div v-if="error" style="font-size: 11px; color: var(--danger-fg)">{{ error }}</div>
  </div>
</template>
