<script setup lang="ts">
definePageMeta({ layout: 'app', middleware: 'auth' })

import { useSchemaStore } from '~/stores/schema'
import { useSpacesStore } from '~/stores/spaces'
import { useSpaceSlug } from '~/composables/useSpaceSlug'

interface Endpoint {
  method: 'GET' | 'POST' | 'PATCH' | 'DELETE'
  path: string
  description: string
  table: { slug: string; name: string }
  coll: boolean
}

const { slug, space } = useSpaceSlug()
const schemaStore = useSchemaStore()

const picked = ref<Endpoint | null>(null)
const tab = ref<'curl' | 'js' | 'py'>('curl')
const copied = ref(false)

onMounted(async () => {
  await schemaStore.fetchTables(slug.value)
  if (schemaStore.tables.length && !picked.value) {
    picked.value = endpoints.value[0] ?? null
  }
})

watch(() => schemaStore.tables, (tables) => {
  if (tables.length && !picked.value) picked.value = endpoints.value[0] ?? null
})

const METHOD_COLOR: Record<string, [string, string]> = {
  GET:    ['#DBEAFE', '#1E40AF'],
  POST:   ['#D1FAE5', '#065F46'],
  PATCH:  ['#FEF3C7', '#92400E'],
  DELETE: ['#FEE2E2', '#991B1B'],
}

function mc(method: string): string { return (METHOD_COLOR[method] ?? ['#eee', '#333'])[0] }
function fc(method: string): string { return (METHOD_COLOR[method] ?? ['#eee', '#333'])[1] }

const endpoints = computed<Endpoint[]>(() => {
  const result: Endpoint[] = []
  for (const t of schemaStore.tables) {
    result.push({ table: t, method: 'GET',    path: `/api/${slug.value}/${t.slug}`,      description: `Список ${t.name} с фильтрами и пагинацией`, coll: true  })
    result.push({ table: t, method: 'GET',    path: `/api/${slug.value}/${t.slug}/{id}`, description: `Одна запись ${t.name} по ID`,                coll: false })
    result.push({ table: t, method: 'POST',   path: `/api/${slug.value}/${t.slug}`,      description: `Создать запись в ${t.name}`,                  coll: true  })
    result.push({ table: t, method: 'PATCH',  path: `/api/${slug.value}/${t.slug}/{id}`, description: `Обновить поля записи ${t.name}`,              coll: false })
    result.push({ table: t, method: 'DELETE', path: `/api/${slug.value}/${t.slug}/{id}`, description: `Удалить запись из ${t.name}`,                 coll: false })
  }
  return result
})

const groupedEndpoints = computed(() => {
  const groups = new Map<string, { table: { slug: string; name: string }; eps: Endpoint[] }>()
  for (const ep of endpoints.value) {
    const key = ep.table.slug
    if (!groups.has(key)) groups.set(key, { table: ep.table, eps: [] })
    groups.get(key)!.eps.push(ep)
  }
  return [...groups.values()]
})

const ep = computed(() => picked.value)

const BASE_URL = 'https://app.nerion.ru'
const SAMPLE_ID = 'rec_a8f2b4'
const SAMPLE_KEY = 'nrn_live_••••••••3f2a'

function resolvedPath(e: Endpoint): string {
  return e.path.replace('{id}', SAMPLE_ID)
}

const SAMPLES: Record<'curl' | 'js' | 'py', (e: Endpoint) => string> = {
  curl: (e) => `curl -X ${e.method} '${BASE_URL}${resolvedPath(e)}' \\
  -H 'X-Api-Key: ${SAMPLE_KEY}' \\
  -H 'Content-Type: application/json'`,

  js: (e) => `const res = await fetch(
  '${BASE_URL}${resolvedPath(e)}',
  {
    method: '${e.method}',
    headers: { 'X-Api-Key': process.env.NERION_KEY },
  }
);
const data = await res.json();`,

  py: (e) => `import requests, os
r = requests.${e.method.toLowerCase()}(
  '${BASE_URL}${resolvedPath(e)}',
  headers={'X-Api-Key': os.environ['NERION_KEY']},
)
data = r.json()`,
}

const RESPONSE_EXAMPLE = computed<string>(() => {
  if (!ep.value) return '{}'
  if (ep.value.method === 'DELETE') return JSON.stringify({ message: 'Запись удалена' }, null, 2)
  const sample = { id: 1 }
  if (ep.value.coll && ep.value.method === 'GET') {
    return JSON.stringify({ data: [sample], meta: { total: 1, limit: 50, offset: 0 } }, null, 2)
  }
  return JSON.stringify(sample, null, 2)
})

const QUERY_PARAMS: [string, string, string][] = [
  ['page',     'integer', 'Номер страницы, по умолчанию 1.'],
  ['per_page', 'integer', 'Записей на странице, до 200. По умолчанию 50.'],
  ['sort',     'string',  'Slug поля; с минусом — убывание (-created_at).'],
  ['q',        'string',  'Полнотекстовый поиск по текстовым полям.'],
]

function copy() {
  if (!ep.value) return
  navigator.clipboard.writeText(SAMPLES[tab.value](ep.value)).then(() => {
    copied.value = true
    setTimeout(() => { copied.value = false }, 1800)
  })
}

const breadcrumb = computed(() => [space.value?.name || slug.value, 'REST API'])
</script>

<template>
  <div :data-breadcrumb="JSON.stringify(breadcrumb)">
    <div
      v-if="schemaStore.loading"
      style="display: flex; align-items: center; justify-content: center; height: 200px; color: var(--fg-3); font-size: 13px"
    >
      Загрузка…
    </div>

    <div
      v-else-if="!schemaStore.tables.length"
      style="display: flex; flex-direction: column; align-items: center; justify-content: center; height: 240px; gap: 10px"
    >
      <NIcon name="table" :size="32" color="var(--fg-3)" />
      <div style="font-size: 15px; font-weight: 700">Нет таблиц</div>
      <div style="font-size: 13px; color: var(--fg-2)">Создай таблицу — API появится автоматически</div>
    </div>

    <div v-else style="display: grid; grid-template-columns: 260px 1fr; min-height: calc(100vh - 56px)">
      <!-- Sidebar -->
      <aside style="background: var(--bg-0); border-right: 0.5px solid var(--border-default); padding: 20px 0; overflow-y: auto">
        <div style="display: flex; align-items: center; gap: 10px; padding: 0 16px 12px">
          <NStatusDot status="online" />
          <span style="font-size: 11px; color: var(--fg-3); margin-left: auto">API активен</span>
        </div>

        <!-- Link to keys page -->
        <div style="padding: 0 16px 12px">
          <NButton
            variant="secondary"
            size="sm"
            style="width: 100%"
            @click="navigateTo(`/spaces/${slug}/api/keys`)"
          >
            <NIcon name="key" :size="13" />
            Ключи API
          </NButton>
        </div>

        <template v-for="group in groupedEndpoints" :key="group.table.slug">
          <div
            style="padding: 8px 16px 6px; font-size: 10px; color: var(--fg-3); text-transform: uppercase; letter-spacing: 0.06em; font-weight: 600; margin-top: 4px"
          >{{ group.table.name }}</div>
          <button
            v-for="(e, i) in group.eps"
            :key="i"
            @click="picked = e"
            :style="{
              width: '100%', padding: '8px 16px', display: 'flex', alignItems: 'center', gap: '10px',
              background: picked === e ? 'var(--brand-tint)' : 'transparent',
              border: 0,
              borderLeft: `3px solid ${picked === e ? 'var(--brand-primary)' : 'transparent'}`,
              cursor: 'pointer', textAlign: 'left',
              boxSizing: 'border-box',
            }"
          >
            <span
              :style="{
                fontSize: '9px', fontWeight: 700, padding: '2px 6px', borderRadius: '3px',
                background: mc(e.method), color: fc(e.method),
                fontFamily: 'var(--font-mono)', minWidth: '38px', textAlign: 'center',
                flexShrink: 0,
              }"
            >{{ e.method === 'DELETE' ? 'DEL' : e.method }}</span>
            <span
              style="font-size: 11px; font-family: var(--font-mono); color: var(--fg-2); overflow: hidden; text-overflow: ellipsis; white-space: nowrap"
            >{{ e.path.replace(`/api/${slug}/`, '/') }}</span>
          </button>
        </template>
      </aside>

      <!-- Main detail panel -->
      <main v-if="ep" style="padding: 32px 40px 80px; max-width: 840px">
        <!-- Method + path header -->
        <div style="display: flex; align-items: center; gap: 10px; margin-bottom: 10px">
          <span
            :style="{
              fontSize: '11px', fontWeight: 700, padding: '4px 10px', borderRadius: '4px',
              background: mc(ep.method), color: fc(ep.method),
              fontFamily: 'var(--font-mono)',
            }"
          >{{ ep.method }}</span>
          <code style="font-size: 14px; font-family: var(--font-mono); color: var(--fg-1)">{{ ep.path }}</code>
        </div>
        <p style="font-size: 14px; color: var(--fg-2); margin-bottom: 28px">{{ ep.description }}.</p>

        <!-- Query params — only for GET collection -->
        <section v-if="ep.method === 'GET' && ep.coll" style="margin-bottom: 28px">
          <div
            style="font-size: 12px; font-weight: 600; color: var(--fg-3); text-transform: uppercase; letter-spacing: 0.06em; margin-bottom: 10px"
          >Параметры query</div>
          <div style="background: var(--bg-0); border: 0.5px solid var(--border-default); border-radius: 8px; overflow: hidden">
            <div
              v-for="([k, t, d], i) in QUERY_PARAMS"
              :key="k"
              :style="{
                display: 'grid', gridTemplateColumns: '160px 90px 1fr',
                padding: '10px 16px', gap: '12px',
                borderTop: i ? '0.5px solid var(--border-default)' : '0',
                alignItems: 'start',
              }"
            >
              <code style="font-size: 12px; font-family: var(--font-mono); color: var(--fg-1)">{{ k }}</code>
              <span style="font-size: 12px; color: var(--fg-3)">{{ t }}</span>
              <span style="font-size: 13px; color: var(--fg-2)">{{ d }}</span>
            </div>
          </div>
        </section>

        <!-- Code samples -->
        <section style="margin-bottom: 28px">
          <div style="display: flex; align-items: center; justify-content: space-between; margin-bottom: 10px">
            <div
              style="font-size: 12px; font-weight: 600; color: var(--fg-3); text-transform: uppercase; letter-spacing: 0.06em"
            >Запрос</div>
            <div style="display: flex; gap: 4px">
              <button
                v-for="[k, lbl] in ([['curl', 'cURL'], ['js', 'JavaScript'], ['py', 'Python']] as [string, string][])"
                :key="k"
                @click="tab = k as 'curl' | 'js' | 'py'"
                :style="{
                  padding: '5px 10px', fontSize: '12px',
                  background: tab === k ? 'var(--bg-0)' : 'transparent',
                  border: tab === k ? '0.5px solid var(--border-strong)' : '0.5px solid transparent',
                  borderRadius: '4px', cursor: 'pointer',
                  color: tab === k ? 'var(--fg-1)' : 'var(--fg-2)',
                  fontFamily: 'inherit',
                }"
              >{{ lbl }}</button>
            </div>
          </div>
          <div style="position: relative; background: #111827; border-radius: 8px; overflow: hidden">
            <button
              @click="copy"
              style="position: absolute; top: 12px; right: 12px; background: rgba(255,255,255,.08); border: 0; color: rgba(255,255,255,.7); padding: 5px 10px; border-radius: 4px; cursor: pointer; font-size: 11px; display: flex; align-items: center; gap: 4px; font-family: inherit"
            >
              <NIcon :name="copied ? 'check' : 'copy'" :size="11" color="rgba(255,255,255,.7)" />
              {{ copied ? 'Скопировано' : 'Копировать' }}
            </button>
            <pre
              style="margin: 0; padding: 20px 24px; color: rgba(255,255,255,.9); font-size: 13px; font-family: var(--font-mono); line-height: 1.6; overflow: auto; white-space: pre-wrap"
            >{{ SAMPLES[tab](ep) }}</pre>
          </div>
        </section>

        <!-- Response example -->
        <section>
          <div
            style="font-size: 12px; font-weight: 600; color: var(--fg-3); text-transform: uppercase; letter-spacing: 0.06em; margin-bottom: 10px"
          >
            Ответ
            <span style="color: var(--green-600); font-family: var(--font-mono); margin-left: 8px; text-transform: none; font-weight: 500">200 OK</span>
          </div>
          <div style="background: #111827; border-radius: 8px; padding: 20px 24px">
            <pre
              style="margin: 0; color: rgba(255,255,255,.9); font-size: 13px; font-family: var(--font-mono); line-height: 1.6; white-space: pre-wrap"
            >{{ RESPONSE_EXAMPLE }}</pre>
          </div>
        </section>
      </main>

      <!-- Placeholder when nothing picked yet -->
      <main
        v-else
        style="display: flex; align-items: center; justify-content: center; height: 200px; color: var(--fg-3); font-size: 13px"
      >
        Выбери эндпоинт в списке слева
      </main>
    </div>
  </div>
</template>
