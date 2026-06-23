<script setup lang="ts">
import { ref, computed, watch, nextTick, onMounted } from 'vue'

const emit = defineEmits<{ close: [] }>()

const q = ref('')
const activeIndex = ref(0)

interface PaletteItem {
  id: string
  icon: string
  label: string
  hint: string
  group: string
  to: string
}

const ALL: PaletteItem[] = [
  { id: 'data-prep',    icon: 'table',    label: 'Преподаватели',          hint: '14 записей',               group: 'Таблицы',    to: '/spaces' },
  { id: 'data-courses', icon: 'table',    label: 'Курсы',                  hint: '42 записи',                group: 'Таблицы',    to: '/spaces' },
  { id: 'data-plans',   icon: 'table',    label: 'Учебные планы',          hint: '8 записей',                group: 'Таблицы',    to: '/spaces' },
  { id: 'sch02',        icon: 'plus',     label: 'Создать таблицу',        hint: 'Из шаблона или с нуля',    group: 'Действия',   to: '/spaces' },
  { id: 'api01',        icon: 'code',     label: 'REST API',               hint: 'Документация эндпоинтов',  group: 'Выходы',     to: '/spaces' },
  { id: 'api02',        icon: 'key',      label: 'Ключи API',              hint: 'Создать или отозвать',     group: 'Выходы',     to: '/spaces' },
  { id: 'files',        icon: 'folder',   label: 'Файлы и списки',         hint: 'Списки для фронта · API',  group: 'Выходы',     to: '/spaces' },
  { id: 'pdf01',        icon: 'file',     label: 'PDF',                    hint: 'Шаблоны и генерация',      group: 'Выходы',     to: '/spaces' },
  { id: 'set01',        icon: 'users',    label: 'Команда',                hint: 'Участники и роли',         group: 'Настройки',  to: '/spaces' },
  { id: 'set02',        icon: 'settings', label: 'Настройки пространства', hint: 'Имя, slug, удаление',      group: 'Настройки',  to: '/spaces' },
  { id: 'spaces',       icon: 'box',      label: 'Все пространства',       hint: 'Сменить пространство',     group: 'Навигация',  to: '/spaces' },
]

const filtered = computed<PaletteItem[]>(() => {
  const qv = q.value.trim().toLowerCase()
  return qv
    ? ALL.filter(i => i.label.toLowerCase().includes(qv) || i.hint.toLowerCase().includes(qv))
    : ALL
})

const groups = computed<Record<string, PaletteItem[]>>(() => {
  const acc: Record<string, PaletteItem[]> = {}
  for (const item of filtered.value) {
    ;(acc[item.group] = acc[item.group] || []).push(item)
  }
  return acc
})

// Flat list for keyboard navigation
const flatItems = computed<PaletteItem[]>(() => filtered.value)

function clampIndex(n: number) {
  const max = flatItems.value.length - 1
  if (max < 0) return 0
  return Math.max(0, Math.min(n, max))
}

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'ArrowDown') {
    e.preventDefault()
    activeIndex.value = clampIndex(activeIndex.value + 1)
  } else if (e.key === 'ArrowUp') {
    e.preventDefault()
    activeIndex.value = clampIndex(activeIndex.value - 1)
  } else if (e.key === 'Enter') {
    const item = flatItems.value[activeIndex.value]
    if (item) go(item)
  } else if (e.key === 'Escape') {
    emit('close')
  }
}

// Reset active item when filter changes
watch(filtered, () => {
  activeIndex.value = 0
})

function go(item: PaletteItem) {
  navigateTo(item.to)
  emit('close')
}

// Autofocus input on mount
const inputRef = ref<HTMLInputElement | null>(null)
onMounted(() => {
  nextTick(() => inputRef.value?.focus())
})
</script>

<template>
  <Teleport to="body">
    <div
      @click="emit('close')"
      :style="{
        position: 'fixed',
        inset: 0,
        background: 'var(--overlay)',
        backdropFilter: 'blur(2px)',
        display: 'grid',
        placeItems: 'start center',
        zIndex: 200,
        paddingTop: '100px',
      }"
    >
      <div
        @click.stop
        @keydown="onKeydown"
        :style="{
          background: 'var(--bg-0)',
          borderRadius: '12px',
          boxShadow: 'var(--shadow-3)',
          width: '560px',
          maxWidth: 'calc(100vw - 32px)',
          maxHeight: '60vh',
          display: 'flex',
          flexDirection: 'column',
          overflow: 'hidden',
          animation: 'modalIn 200ms',
        }"
      >
        <!-- Search input -->
        <div
          style="
            display: flex;
            align-items: center;
            gap: 12px;
            padding: 14px 18px;
            border-bottom: 0.5px solid var(--border-default);
          "
        >
          <NIcon name="search" :size="16" color="var(--fg-3)" />
          <input
            ref="inputRef"
            v-model="q"
            autofocus
            placeholder="Поиск или команда…"
            style="
              flex: 1;
              border: 0;
              outline: 0;
              font-size: 15px;
              background: transparent;
              color: var(--fg-1);
              font-family: inherit;
            "
          />
          <span style="font-size: 11px; color: var(--fg-3); font-family: var(--font-mono)">esc</span>
        </div>

        <!-- Results -->
        <div style="overflow: auto; padding: 8px">
          <div
            v-if="!Object.keys(groups).length"
            style="padding: 32px; text-align: center; color: var(--fg-3); font-size: 13px"
          >
            Ничего не найдено
          </div>

          <template v-for="(items, group) in groups" :key="group">
            <div style="margin-bottom: 4px">
              <div
                style="
                  font-size: 10px;
                  font-weight: 600;
                  color: var(--fg-3);
                  text-transform: uppercase;
                  letter-spacing: 0.07em;
                  padding: 8px 12px 4px;
                "
              >{{ group }}</div>

              <div
                v-for="item in items"
                :key="item.id"
                @click="go(item)"
                @mouseenter="activeIndex = flatItems.indexOf(item)"
                :style="{
                  display: 'flex',
                  alignItems: 'center',
                  gap: '10px',
                  padding: '7px 12px',
                  borderRadius: '6px',
                  cursor: 'pointer',
                  background: flatItems.indexOf(item) === activeIndex ? 'var(--bg-1)' : 'transparent',
                  transition: 'background 80ms',
                }"
              >
                <NIcon :name="item.icon" :size="14" color="var(--fg-3)" />
                <span style="font-size: 13px; font-weight: 500; color: var(--fg-1)">{{ item.label }}</span>
                <span style="font-size: 12px; color: var(--fg-3); margin-left: 4px">{{ item.hint }}</span>
              </div>
            </div>
          </template>
        </div>

        <!-- Footer hints -->
        <div
          style="
            border-top: 0.5px solid var(--border-default);
            padding: 8px 16px;
            font-size: 11px;
            color: var(--fg-3);
            display: flex;
            gap: 16px;
          "
        >
          <span>↵ открыть</span><span>↑↓ навигация</span><span>esc закрыть</span>
        </div>
      </div>
    </div>
  </Teleport>
</template>
