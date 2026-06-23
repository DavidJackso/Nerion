<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'

const props = defineProps<{
  breadcrumb?: string[]
}>()

const showPalette = ref(false)

function onKey(e: KeyboardEvent) {
  if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === 'k') {
    e.preventDefault()
    showPalette.value = true
  }
  if (e.key === 'Escape') showPalette.value = false
}

onMounted(() => window.addEventListener('keydown', onKey))
onUnmounted(() => window.removeEventListener('keydown', onKey))
</script>

<template>
  <div
    :style="{
      display: 'grid',
      gridTemplateColumns: '220px 1fr',
      minHeight: '100vh',
      background: 'var(--bg-1)',
      fontFamily: 'var(--font-sans)',
    }"
  >
    <AppSidebar />

    <div style="display: flex; flex-direction: column; min-width: 0">
      <!-- Top bar -->
      <header
        :style="{
          height: '56px',
          background: 'var(--bg-0)',
          borderBottom: '0.5px solid var(--border-default)',
          display: 'flex',
          alignItems: 'center',
          padding: '0 24px',
          gap: '16px',
          position: 'sticky',
          top: 0,
          zIndex: 10,
          boxSizing: 'border-box',
        }"
      >
        <div
          style="
            display: flex;
            align-items: center;
            gap: 6px;
            font-size: 13px;
            flex: 1;
            min-width: 0;
            overflow: hidden;
          "
        >
          <template v-for="(b, i) in (breadcrumb ?? [])" :key="i">
            <span v-if="i > 0" style="color: var(--fg-3); flex-shrink: 0">/</span>
            <span
              :style="{
                color: i === (breadcrumb ?? []).length - 1 ? 'var(--fg-1)' : 'var(--fg-2)',
                fontWeight: i === (breadcrumb ?? []).length - 1 ? 500 : 400,
                whiteSpace: 'nowrap',
                overflow: 'hidden',
                textOverflow: 'ellipsis',
              }"
            >{{ b }}</span>
          </template>
        </div>

        <button
          @click="showPalette = true"
          :style="{
            display: 'flex',
            alignItems: 'center',
            gap: '8px',
            height: '32px',
            padding: '0 12px',
            background: 'var(--bg-1)',
            border: '0.5px solid var(--border-default)',
            borderRadius: '6px',
            color: 'var(--fg-3)',
            fontSize: '13px',
            cursor: 'pointer',
            minWidth: '220px',
            fontFamily: 'inherit',
          }"
        >
          <NIcon name="search" :size="14" />
          <span style="flex: 1; text-align: left">Поиск или команда…</span>
          <span
            style="
              font-family: var(--font-mono);
              font-size: 11px;
              padding: 1px 5px;
              background: var(--bg-2);
              border-radius: 3px;
              color: var(--fg-2);
              flex-shrink: 0;
            "
          >⌘K</span>
        </button>

        <div v-if="$slots.actions" style="display: flex; gap: 8px; align-items: center">
          <slot name="actions" />
        </div>
      </header>

      <div style="flex: 1; overflow: auto">
        <slot />
      </div>
    </div>

    <CommandPalette v-if="showPalette" @close="showPalette = false" />
  </div>
</template>
