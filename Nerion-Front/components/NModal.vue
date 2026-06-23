<script setup lang="ts">
defineProps<{
  open?: boolean
  title?: string
  subtitle?: string
  width?: number
}>()

const emit = defineEmits<{
  close: []
}>()
</script>

<template>
  <Teleport to="body">
    <div v-if="open" @click="emit('close')" :style="{
      position: 'fixed', inset: 0,
      background: 'var(--overlay)',
      backdropFilter: 'blur(2px)',
      display: 'grid', placeItems: 'center',
      zIndex: 100,
      animation: 'modalFade 180ms',
    }">
      <div @click.stop :style="{
        background: 'var(--bg-0)',
        borderRadius: '12px',
        boxShadow: 'var(--shadow-3)',
        width: `${width ?? 480}px`,
        maxWidth: 'calc(100vw - 32px)',
        maxHeight: 'calc(100vh - 64px)',
        display: 'flex',
        flexDirection: 'column',
        overflow: 'hidden',
        animation: 'modalIn 250ms',
      }">
        <div :style="{
          padding: '18px 24px',
          borderBottom: '0.5px solid var(--border-default)',
          display: 'flex',
          alignItems: 'flex-start',
          justifyContent: 'space-between',
          gap: '12px',
        }">
          <div>
            <div style="font-size: 16px; font-weight: 600; color: var(--fg-1)">{{ title }}</div>
            <div v-if="subtitle" style="font-size: 12px; color: var(--fg-2); margin-top: 3px">{{ subtitle }}</div>
          </div>
          <button @click="emit('close')" :style="{
            background: 0, border: 0, cursor: 'pointer',
            color: 'var(--fg-3)', padding: '4px', lineHeight: 1, flexShrink: 0,
          }">
            <NIcon name="x" :size="16" />
          </button>
        </div>
        <div style="padding: 24px; overflow: auto; flex: 1">
          <slot />
        </div>
        <div v-if="$slots.footer" :style="{
          padding: '14px 24px',
          borderTop: '0.5px solid var(--border-default)',
          display: 'flex',
          justifyContent: 'flex-end',
          gap: '8px',
        }">
          <slot name="footer" />
        </div>
      </div>
    </div>
  </Teleport>
</template>
