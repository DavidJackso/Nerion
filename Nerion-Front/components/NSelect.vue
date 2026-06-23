<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, nextTick } from 'vue'

interface SelectOption {
  value?: unknown
  label?: string
  color?: string
  dim?: boolean
  group?: string
  items?: SelectOption[]
}

defineOptions({ inheritAttrs: false })

const props = defineProps<{
  modelValue?: unknown
  options?: SelectOption[]
  placeholder?: string
  disabled?: boolean
  size?: 'sm' | 'md'
}>()

const emit = defineEmits<{
  'update:modelValue': [value: unknown]
}>()

const open = ref(false)
const triggerRef = ref<HTMLElement | null>(null)
const dropRef = ref<HTMLElement | null>(null)
const dropPos = ref<{ top?: number; left?: number; width?: number }>({})

const h = computed(() => props.size === 'sm' ? '30px' : '36px')
const fs = computed(() => props.size === 'sm' ? '12px' : '14px')

const selected = computed(() => {
  const v = props.modelValue
  if (v === null || v === undefined) return null
  for (const opt of props.options ?? []) {
    if (opt.group !== undefined) {
      const found = opt.items?.find(i => i.value === v)
      if (found) return found
    } else if (opt.value === v) {
      return opt
    }
  }
  return null
})

const labelColor = computed(() => {
  if (!selected.value) return 'var(--fg-3)'
  if (selected.value.dim) return 'var(--fg-3)'
  return selected.value.color || 'var(--fg-1)'
})

function toggle() {
  if (props.disabled) return
  open.value = !open.value
  if (open.value) nextTick(position)
}

function position() {
  const r = triggerRef.value?.getBoundingClientRect()
  if (!r) return
  dropPos.value = { top: r.bottom + 4, left: r.left, width: r.width }
}

function pick(val: unknown) {
  emit('update:modelValue', val)
  open.value = false
}

function outside(e: MouseEvent) {
  if (!open.value) return
  if (triggerRef.value?.contains(e.target as Node) || dropRef.value?.contains(e.target as Node)) return
  open.value = false
}

onMounted(() => document.addEventListener('mousedown', outside))
onUnmounted(() => document.removeEventListener('mousedown', outside))
</script>

<template>
  <div
    ref="triggerRef"
    v-bind="$attrs"
    class="nsel-trigger"
    :class="{ 'nsel-open': open }"
    :style="{ height: h, opacity: disabled ? 0.5 : 1, cursor: disabled ? 'not-allowed' : 'pointer' }"
    @click="toggle"
  >
    <span class="nsel-label" :style="{ fontSize: fs, color: labelColor, fontWeight: selected?.color && !selected?.dim ? 500 : 400 }">
      {{ selected?.label || placeholder || '—' }}
    </span>
    <NIcon name="chevd" :size="10" color="var(--fg-3)" :style="{ transform: open ? 'rotate(180deg)' : 'none', transition: 'transform 140ms', flexShrink: 0 }" />
  </div>

  <Teleport to="body">
    <div
      v-if="open"
      ref="dropRef"
      class="nsel-drop"
      :style="{ top: dropPos.top + 'px', left: dropPos.left + 'px', width: dropPos.width + 'px' }"
    >
      <template v-for="opt in (options ?? [])" :key="opt.value ?? opt.group">
        <template v-if="opt.group !== undefined">
          <div class="nsel-group">{{ opt.group }}</div>
          <div
            v-for="item in opt.items"
            :key="String(item.value)"
            class="nsel-opt"
            :class="{ 'nsel-active': modelValue === item.value }"
            :style="{ fontSize: fs }"
            @click="pick(item.value)"
          >{{ item.label }}</div>
        </template>
        <div
          v-else
          class="nsel-opt"
          :class="{ 'nsel-active': modelValue === opt.value }"
          :style="{ fontSize: fs, color: opt.color || undefined }"
          @click="pick(opt.value)"
        >{{ opt.label }}</div>
      </template>
    </div>
  </Teleport>
</template>

<style scoped>
.nsel-trigger {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 6px;
  padding: 0 8px 0 10px;
  border-radius: 6px;
  border: 0.5px solid var(--border-strong);
  background: var(--bg-0);
  user-select: none;
  box-sizing: border-box;
  width: 100%;
  transition: border-color 120ms, box-shadow 120ms;
}
.nsel-trigger:hover {
  border-color: var(--border-focus);
}
.nsel-open {
  border-color: var(--border-focus);
  box-shadow: var(--ring-focus);
}
.nsel-label {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.nsel-drop {
  position: fixed;
  background: var(--bg-0);
  border: 0.5px solid var(--border-strong);
  border-radius: 8px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.12);
  overflow: auto;
  max-height: 260px;
  padding: 4px;
  z-index: 9999;
}
.nsel-group {
  font-size: 10px;
  color: var(--fg-3);
  padding: 6px 8px 2px;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  font-weight: 600;
}
.nsel-opt {
  padding: 6px 10px;
  border-radius: 5px;
  color: var(--fg-1);
  cursor: pointer;
  transition: background 80ms;
}
.nsel-opt:hover {
  background: var(--bg-2);
}
.nsel-active {
  background: var(--brand-tint);
  color: var(--brand-primary) !important;
  font-weight: 500;
}
</style>
