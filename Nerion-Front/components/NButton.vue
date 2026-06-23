<script setup lang="ts">
import { ref, computed } from 'vue'

type Variant = 'primary' | 'secondary' | 'ghost' | 'danger' | 'danger-ghost'
type Size = 'sm' | 'md' | 'lg'

const props = defineProps<{
  variant?: Variant
  size?: Size
  disabled?: boolean
  type?: string
}>()

defineEmits<{
  click: [event: MouseEvent]
}>()

const hov = ref(false)

const sizes: Record<Size, { height: string; padding: string; fontSize: string }> = {
  sm: { height: '28px', padding: '0 10px', fontSize: '13px' },
  md: { height: '36px', padding: '0 14px', fontSize: '14px' },
  lg: { height: '44px', padding: '0 20px', fontSize: '15px' },
}

const vars: Record<Variant, { background: string; color: string; border: string }> = {
  primary:      { background: 'var(--brand-primary)', color: '#fff', border: 'none' },
  secondary:    { background: 'var(--bg-0)', color: 'var(--fg-1)', border: '0.5px solid var(--border-strong)' },
  ghost:        { background: 'transparent', color: 'var(--fg-1)', border: 'none' },
  danger:       { background: 'var(--danger-solid)', color: '#fff', border: 'none' },
  'danger-ghost': { background: 'transparent', color: 'var(--danger-fg)', border: '0.5px solid var(--danger-border)' },
}

const hovStyles: Record<Variant, { background: string }> = {
  primary:      { background: 'var(--brand-primary-hover)' },
  secondary:    { background: 'var(--bg-2)' },
  ghost:        { background: 'var(--bg-2)' },
  danger:       { background: 'var(--red-600)' },
  'danger-ghost': { background: 'var(--danger-bg)' },
}

const style = computed(() => {
  const v = props.variant ?? 'primary'
  const s = props.size ?? 'md'
  return {
    display: 'inline-flex',
    alignItems: 'center',
    justifyContent: 'center',
    gap: '6px',
    border: vars[v]?.border || 'none',
    borderRadius: '6px',
    cursor: props.disabled ? 'not-allowed' : 'pointer',
    fontFamily: 'var(--font-sans)',
    fontWeight: 500,
    lineHeight: 1,
    transition: 'background 120ms, transform 80ms',
    opacity: props.disabled ? 0.5 : 1,
    whiteSpace: 'nowrap',
    boxSizing: 'border-box',
    ...sizes[s],
    ...vars[v],
    ...(hov.value && !props.disabled ? hovStyles[v] : {}),
  }
})
</script>

<template>
  <button
    :type="(type as any) ?? 'button'"
    :disabled="disabled"
    :style="style"
    @mouseenter="hov = true"
    @mouseleave="hov = false"
    @click="$emit('click', $event)"
  >
    <slot />
  </button>
</template>
