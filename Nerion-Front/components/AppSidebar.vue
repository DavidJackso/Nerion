<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useAuthStore } from '~/stores/auth'
import { useSpacesStore } from '~/stores/spaces'
import { useSchemaStore } from '~/stores/schema'

const route = useRoute()
const router = useRouter()

const auth = useAuthStore()
const spacesStore = useSpacesStore()
const schemaStore = useSchemaStore()

const slug = computed(() => route.params.slug as string)
const space = computed(() => spacesStore.spaces.find((s: { slug: string }) => s.slug === slug.value))

onMounted(async () => {
  if (!spacesStore.spaces.length) await spacesStore.fetchSpaces()
  if (slug.value) await schemaStore.fetchTables(slug.value)
})

watch(slug, async (val) => {
  if (val) await schemaStore.fetchTables(val)
})

function go(path: string) {
  router.push(path)
}

function isActive(name: string): boolean {
  return route.name === name
}

function tableActive(t: { slug: string }): boolean {
  return route.name === 'spaces-slug-tables-table' && route.params.table === t.slug
}

async function logout() {
  await auth.logout()
  navigateTo('/login')
}

const hovered = ref<string | null>(null)

interface NavItem {
  icon: string
  label: string
  name: string
  path: string
}

const outputItems: NavItem[] = [
  { icon: 'code',   label: 'REST API', name: 'api-docs', path: `/spaces/${slug.value}/api/docs` },
  { icon: 'folder', label: 'Файлы',    name: 'files',    path: `/spaces/${slug.value}/files`    },
  { icon: 'file',   label: 'PDF',      name: 'pdf',      path: `/spaces/${slug.value}/pdf`      },
]

const spaceItems: NavItem[] = [
  { icon: 'users',    label: 'Команда',    name: 'team',     path: `/spaces/${slug.value}/team`     },
  { icon: 'settings', label: 'Настройки', name: 'settings', path: `/spaces/${slug.value}/settings` },
]

// Recompute path-based nav items when slug changes
const outputNav = computed<NavItem[]>(() => [
  { icon: 'code',   label: 'REST API', name: 'api-docs', path: `/spaces/${slug.value}/api/docs` },
  { icon: 'folder', label: 'Файлы',    name: 'files',    path: `/spaces/${slug.value}/files`    },
  { icon: 'file',   label: 'PDF',      name: 'pdf',      path: `/spaces/${slug.value}/pdf`      },
])

const spaceNav = computed<NavItem[]>(() => [
  { icon: 'users',    label: 'Команда',    name: 'team',     path: `/spaces/${slug.value}/team`     },
  { icon: 'settings', label: 'Настройки', name: 'settings', path: `/spaces/${slug.value}/settings` },
])
</script>

<template>
  <aside
    :style="{
      background: 'var(--bg-0)',
      borderRight: '0.5px solid var(--border-default)',
      padding: '12px 8px',
      display: 'flex',
      flexDirection: 'column',
      position: 'sticky',
      top: 0,
      height: '100vh',
      overflow: 'auto',
      boxSizing: 'border-box',
    }"
  >
    <!-- Space switcher -->
    <NuxtLink
      to="/spaces"
      :style="{
        display: 'flex',
        alignItems: 'center',
        gap: '8px',
        padding: '6px 10px 12px',
        cursor: 'pointer',
        borderBottom: '0.5px solid var(--border-default)',
        marginBottom: '4px',
        textDecoration: 'none',
      }"
    >
      <div
        style="
          width: 26px;
          height: 26px;
          border-radius: 6px;
          background: var(--brand-primary);
          color: #fff;
          display: grid;
          place-items: center;
          font-weight: 700;
          font-size: 13px;
          flex-shrink: 0;
        "
      >N</div>
      <div
        style="
          display: flex;
          flex-direction: column;
          line-height: 1.2;
          overflow: hidden;
          flex: 1;
        "
      >
        <span
          style="
            font-size: 13px;
            font-weight: 600;
            color: var(--fg-1);
            white-space: nowrap;
            overflow: hidden;
            text-overflow: ellipsis;
          "
        >{{ space?.name || slug }}</span>
        <span style="font-size: 10px; color: var(--fg-3); font-family: var(--font-mono)">{{ slug }}</span>
      </div>
      <NIcon name="chevd" :size="12" color="var(--fg-3)" />
    </NuxtLink>

    <!-- Tables section -->
    <div
      :style="{
        fontSize: '10px',
        fontWeight: 600,
        color: 'var(--fg-3)',
        textTransform: 'uppercase',
        letterSpacing: '0.07em',
        padding: '14px 10px 5px',
      }"
    >Таблицы</div>

    <NuxtLink
      v-for="t in schemaStore.tables"
      :key="t.slug"
      :to="`/spaces/${slug}/tables/${t.slug}`"
      @mouseenter="hovered = t.slug"
      @mouseleave="hovered = null"
      :style="{
        display: 'flex',
        alignItems: 'center',
        gap: '8px',
        height: '30px',
        padding: '0 10px',
        borderRadius: '6px',
        cursor: 'pointer',
        marginBottom: '1px',
        background: tableActive(t) ? 'var(--brand-tint)' : hovered === t.slug ? 'var(--bg-1)' : 'transparent',
        color: tableActive(t) ? 'var(--purple-700)' : 'var(--fg-1)',
        fontSize: '13px',
        fontWeight: tableActive(t) ? 500 : 400,
        transition: 'background 100ms',
        textDecoration: 'none',
      }"
    >
      <NIcon name="table" :size="14" :style="{ opacity: tableActive(t) ? 1 : 0.65 }" />
      <span style="flex: 1; overflow: hidden; text-overflow: ellipsis; white-space: nowrap">{{ t.name }}</span>
    </NuxtLink>

    <div
      v-if="schemaStore.loading && !schemaStore.tables.length"
      style="padding: 6px 10px; font-size: 12px; color: var(--fg-3)"
    >Загрузка…</div>

    <NuxtLink
      :to="`/spaces/${slug}/schema/new`"
      @mouseenter="hovered = '__new'"
      @mouseleave="hovered = null"
      :style="{
        display: 'flex',
        alignItems: 'center',
        gap: '8px',
        height: '30px',
        padding: '0 10px',
        borderRadius: '6px',
        cursor: 'pointer',
        marginBottom: '1px',
        background: isActive('spaces-slug-schema-new') ? 'var(--brand-tint)' : hovered === '__new' ? 'var(--bg-1)' : 'transparent',
        color: isActive('spaces-slug-schema-new') ? 'var(--purple-700)' : 'var(--fg-2)',
        fontSize: '13px',
        transition: 'background 100ms',
        textDecoration: 'none',
      }"
    >
      <NIcon name="plus" :size="14" :style="{ opacity: 0.65 }" />
      <span>Новая таблица</span>
    </NuxtLink>

    <!-- Outputs section -->
    <div
      :style="{
        fontSize: '10px',
        fontWeight: 600,
        color: 'var(--fg-3)',
        textTransform: 'uppercase',
        letterSpacing: '0.07em',
        padding: '14px 10px 5px',
      }"
    >Выходы</div>

    <NuxtLink
      v-for="item in outputNav"
      :key="item.name"
      :to="item.path"
      @mouseenter="hovered = item.name"
      @mouseleave="hovered = null"
      :style="{
        display: 'flex',
        alignItems: 'center',
        gap: '8px',
        height: '30px',
        padding: '0 10px',
        borderRadius: '6px',
        cursor: 'pointer',
        marginBottom: '1px',
        background: isActive(item.name) ? 'var(--brand-tint)' : hovered === item.name ? 'var(--bg-1)' : 'transparent',
        color: isActive(item.name) ? 'var(--purple-700)' : 'var(--fg-1)',
        fontSize: '13px',
        fontWeight: isActive(item.name) ? 500 : 400,
        transition: 'background 100ms',
        textDecoration: 'none',
      }"
    >
      <NIcon :name="item.icon" :size="14" :style="{ opacity: isActive(item.name) ? 1 : 0.65 }" />
      <span>{{ item.label }}</span>
    </NuxtLink>

    <!-- Space section -->
    <div
      :style="{
        fontSize: '10px',
        fontWeight: 600,
        color: 'var(--fg-3)',
        textTransform: 'uppercase',
        letterSpacing: '0.07em',
        padding: '14px 10px 5px',
      }"
    >Пространство</div>

    <NuxtLink
      v-for="item in spaceNav"
      :key="item.name"
      :to="item.path"
      @mouseenter="hovered = item.name"
      @mouseleave="hovered = null"
      :style="{
        display: 'flex',
        alignItems: 'center',
        gap: '8px',
        height: '30px',
        padding: '0 10px',
        borderRadius: '6px',
        cursor: 'pointer',
        marginBottom: '1px',
        background: isActive(item.name) ? 'var(--brand-tint)' : hovered === item.name ? 'var(--bg-1)' : 'transparent',
        color: isActive(item.name) ? 'var(--purple-700)' : 'var(--fg-1)',
        fontSize: '13px',
        fontWeight: isActive(item.name) ? 500 : 400,
        transition: 'background 100ms',
        textDecoration: 'none',
      }"
    >
      <NIcon :name="item.icon" :size="14" :style="{ opacity: isActive(item.name) ? 1 : 0.65 }" />
      <span>{{ item.label }}</span>
    </NuxtLink>

    <div style="flex: 1" />

    <!-- User footer -->
    <div
      style="
        border-top: 0.5px solid var(--border-default);
        display: flex;
        align-items: center;
        gap: 8px;
        padding: 10px 10px 4px;
      "
    >
      <div
        style="
          width: 24px;
          height: 24px;
          border-radius: 50%;
          background: var(--purple-200);
          color: var(--purple-700);
          display: grid;
          place-items: center;
          font-weight: 700;
          font-size: 11px;
          flex-shrink: 0;
        "
      >{{ auth.initials() }}</div>
      <div style="flex: 1; overflow: hidden">
        <div
          style="
            font-size: 12px;
            font-weight: 500;
            white-space: nowrap;
            overflow: hidden;
            text-overflow: ellipsis;
          "
        >{{ auth.user?.name }}</div>
        <div style="font-size: 10px; color: var(--fg-3)">
          {{ auth.user?.role === 'admin' ? 'Администратор' : 'Участник' }}
        </div>
      </div>
      <button
        @click="logout"
        style="
          cursor: pointer;
          color: var(--fg-3);
          padding: 4px;
          background: transparent;
          border: none;
        "
      >
        <NIcon name="logout" :size="14" />
      </button>
    </div>
  </aside>
</template>
