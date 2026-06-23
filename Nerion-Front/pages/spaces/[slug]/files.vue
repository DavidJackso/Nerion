<script setup lang="ts">
import { listLists, createList, updateList } from '~/api/lists'
import { useSchemaStore } from '~/stores/schema'
import { useSpaceSlug } from '~/composables/useSpaceSlug'
import { useToast } from '~/composables/useToast'

definePageMeta({ layout: 'app', middleware: 'auth' })

interface List {
  id: string
  slug: string
  table_slug: string
  published: boolean
  row_limit?: number
  limit?: number
  created_at: string
}

const schemaStore = useSchemaStore()
const { slug, space } = useSpaceSlug()
const { toast, show } = useToast()

const tab = ref<'lists' | 'all'>('lists')
const activeList = ref<List | null>(null)
const lists = ref<List[]>([])
const loading = ref(false)
const showCreate = ref(false)

const newSlugInput = ref('')
const newTableSlug = ref('')
const newPublished = ref(true)
const newLimit = ref<number | undefined>(undefined)
const createError = ref('')
const createLoading = ref(false)

onMounted(async () => {
  loading.value = true
  try {
    await Promise.all([
      schemaStore.fetchTables(slug.value),
      listLists(slug.value).then((r: List[]) => { lists.value = r }),
    ])
  } catch (e: any) {
    show(e?.message ?? 'Ошибка загрузки')
  } finally {
    loading.value = false
  }
})

function onTablePick(v: string) {
  newTableSlug.value = v
  if (!newSlugInput.value) newSlugInput.value = v
}

async function handleCreateList() {
  createError.value = ''
  if (!newTableSlug.value) { createError.value = 'Выбери таблицу'; return }
  if (!newSlugInput.value) { createError.value = 'Введи slug'; return }
  createLoading.value = true
  try {
    const payload: Record<string, unknown> = {
      slug: newSlugInput.value,
      table_slug: newTableSlug.value,
      published: newPublished.value,
    }
    if (newLimit.value) payload.row_limit = newLimit.value
    const l: List = await createList(slug.value, payload)
    lists.value = [...lists.value, l]
    showCreate.value = false
    activeList.value = l
    newSlugInput.value = ''
    newTableSlug.value = ''
    newPublished.value = true
    newLimit.value = undefined
    show(`Список «${l.slug}» создан`)
  } catch (e: any) {
    createError.value = e?.message ?? 'Ошибка создания'
  } finally {
    createLoading.value = false
  }
}

async function togglePublished(l: List) {
  try {
    const updated: List = await updateList(slug.value, l.slug, { published: !l.published })
    lists.value = lists.value.map(x => x.id === updated.id ? updated : x)
    if (activeList.value?.id === updated.id) activeList.value = updated
    show(updated.published ? 'Список опубликован' : 'Список снят с публикации')
  } catch (e: any) {
    show(e?.message ?? 'Ошибка обновления')
  }
}

function fmtDate(dt: string | undefined): string {
  if (!dt) return '—'
  return new Date(dt).toLocaleDateString('ru')
}

function listEndpoint(l: List): string {
  return `/lists/${slug.value}/${l.slug}`
}

function listCodeSnippet(l: List): string {
  return `const res = await fetch('https://app.nerion.ru${listEndpoint(l)}');
const { data } = await res.json();
console.log(data); // массив записей таблицы ${l.table_slug}`
}

function heroCodeSnippet(): string {
  const first = lists.value[0]
  if (!first) {
    return `const res = await fetch('https://app.nerion.ru/lists/${slug.value}/my-list');
const { data } = await res.json();`
  }
  return `const { data } = await (await fetch('https://app.nerion.ru${listEndpoint(first)}')).json();
data.forEach(rec => render(rec)); // ${first.row_limit ?? 50} записей, без хардкода`
}

function copyEndpoint(l: List) {
  navigator.clipboard.writeText(`https://app.nerion.ru${listEndpoint(l)}`).then(() => show('Адрес скопирован'))
}

function copyCode(text: string) {
  navigator.clipboard.writeText(text).then(() => show('Код скопирован'))
}

function tableName(tableSlug: string): string {
  return schemaStore.tables.find((t: { slug: string; name: string }) => t.slug === tableSlug)?.name ?? tableSlug
}

function openCreate() {
  showCreate.value = true
  activeList.value = null
  tab.value = 'lists'
}

function closeCreate() {
  showCreate.value = false
  createError.value = ''
}

const crumb = computed(() =>
  activeList.value
    ? [space.value?.name ?? slug.value, 'Файлы', activeList.value.slug]
    : [space.value?.name ?? slug.value, 'Файлы']
)
</script>

<template>
  <AppShell :breadcrumb="crumb">
    <template #actions>
      <NButton variant="primary" size="sm" @click="openCreate">
        <NIcon name="plus" :size="13" color="#fff" />
        Новый список
      </NButton>
    </template>

    <!-- Tab bar (only shown in overview mode) -->
    <div
      v-if="!activeList"
      style="background:var(--bg-0);border-bottom:0.5px solid var(--border-default);display:flex;align-items:center;padding:0 32px;gap:4px;height:44px"
    >
      <button
        v-for="[k, l] in ([['lists','Списки'],['all','Все файлы']] as [string,string][])"
        :key="k"
        @click="tab = k as 'lists' | 'all'"
        :style="{
          height: '100%', padding: '0 14px', border: 0,
          borderBottom: `2px solid ${tab === k ? 'var(--brand-primary)' : 'transparent'}`,
          background: 'transparent',
          color: tab === k ? 'var(--fg-1)' : 'var(--fg-2)',
          fontSize: '13px', fontWeight: tab === k ? 600 : 500,
          cursor: 'pointer', fontFamily: 'inherit',
        }"
      >{{ l }}</button>
    </div>

    <!-- ─── LIST DETAIL ──────────────────────────────────────────────── -->
    <div v-if="activeList" style="padding:20px 32px 56px;max-width:1080px">
      <button
        @click="activeList = null"
        style="background:0;border:0;color:var(--fg-3);font-size:12px;cursor:pointer;display:flex;align-items:center;gap:6px;margin-bottom:14px;padding:0;font-family:inherit"
      >
        <NIcon name="arrow" :size="12" style="transform:rotate(180deg)" />Все списки
      </button>

      <div style="display:flex;align-items:center;gap:12px;margin-bottom:6px">
        <h1 style="font-size:22px;font-weight:700;letter-spacing:-0.01em">{{ activeList.slug }}</h1>
        <NBadge tone="brand">{{ tableName(activeList.table_slug) }}</NBadge>
        <NBadge :tone="activeList.published ? 'success' : 'neutral'" :dot="true">
          {{ activeList.published ? 'публичный' : 'по ключу' }}
        </NBadge>
      </div>

      <p style="font-size:13px;color:var(--fg-2);margin-bottom:18px">
        Список публикует записи таблицы <strong>{{ tableName(activeList.table_slug) }}</strong>
        по постоянному адресу. Фронт забирает их без знания ID.
      </p>

      <!-- Endpoint bar -->
      <div style="display:flex;align-items:center;gap:10px;background:var(--bg-0);border:0.5px solid var(--border-strong);border-radius:8px;padding:8px 8px 8px 12px;margin-bottom:20px">
        <span style="font-size:9px;font-weight:700;padding:2px 6px;border-radius:3px;background:#DBEAFE;color:#1E40AF;font-family:var(--font-mono)">GET</span>
        <code style="flex:1;font-size:12.5px;font-family:var(--font-mono);color:var(--fg-1);overflow:hidden;text-overflow:ellipsis;white-space:nowrap">{{ listEndpoint(activeList) }}</code>
        <NButton variant="secondary" size="sm" @click="copyEndpoint(activeList)">
          <NIcon name="copy" :size="12" />Копировать
        </NButton>
        <NButton :variant="activeList.published ? 'ghost' : 'primary'" size="sm" @click="togglePublished(activeList)">
          {{ activeList.published ? 'Снять с публикации' : 'Опубликовать' }}
        </NButton>
      </div>

      <!-- Code panel -->
      <div style="margin-bottom:28px">
        <div style="font-size:11px;color:var(--fg-3);text-transform:uppercase;letter-spacing:0.06em;font-weight:600;margin-bottom:10px">Получить на фронте</div>
        <div style="position:relative;background:#111827;border-radius:8px;overflow:hidden">
          <button
            @click="copyCode(listCodeSnippet(activeList))"
            style="position:absolute;top:8px;right:8px;background:rgba(255,255,255,.08);border:0;color:rgba(255,255,255,.7);padding:4px 9px;border-radius:4px;cursor:pointer;font-size:11px;display:flex;align-items:center;gap:4px;font-family:inherit;z-index:1"
          >
            <NIcon name="copy" :size="11" color="rgba(255,255,255,.7)" />Копировать
          </button>
          <pre style="margin:0;padding:16px 18px;color:rgba(255,255,255,.9);font-size:12px;font-family:var(--font-mono);line-height:1.6;overflow:auto;white-space:pre-wrap">{{ listCodeSnippet(activeList) }}</pre>
        </div>
      </div>

      <!-- Info grid -->
      <div style="background:var(--bg-0);border:0.5px solid var(--border-default);border-radius:8px;padding:16px 20px">
        <div style="display:grid;grid-template-columns:140px 1fr;gap:8px 16px;font-size:13px">
          <div style="color:var(--fg-3)">Таблица</div>
          <div style="color:var(--fg-1);font-weight:500">{{ tableName(activeList.table_slug) }}</div>
          <div style="color:var(--fg-3)">Slug</div>
          <div><code style="font-family:var(--font-mono);font-size:12px;color:var(--fg-1)">{{ activeList.slug }}</code></div>
          <div style="color:var(--fg-3)">Лимит записей</div>
          <div style="color:var(--fg-1)">{{ activeList.row_limit ?? 50 }}</div>
          <div style="color:var(--fg-3)">Создан</div>
          <div style="color:var(--fg-2)">{{ fmtDate(activeList.created_at) }}</div>
          <div style="color:var(--fg-3)">Статус</div>
          <div>
            <NBadge :tone="activeList.published ? 'success' : 'neutral'" :dot="true">
              {{ activeList.published ? 'Опубликован' : 'Скрыт' }}
            </NBadge>
          </div>
        </div>
      </div>
    </div>

    <!-- ─── LISTS OVERVIEW ───────────────────────────────────────────── -->
    <div v-else-if="tab === 'lists'" style="padding:20px 32px 56px;max-width:1080px">
      <div style="margin-bottom:18px">
        <h1 style="font-size:20px;font-weight:700;margin-bottom:4px">Списки</h1>
        <div style="font-size:12px;color:var(--fg-3)">Публикуй записи таблиц по постоянному адресу — фронт забирает их без знания ID</div>
      </div>

      <!-- Hero snippet -->
      <div style="background:linear-gradient(90deg,var(--brand-tint),#F8F4FF);border:0.5px solid var(--purple-200);border-radius:12px;padding:18px;margin-bottom:22px">
        <div style="display:flex;align-items:center;gap:14px;margin-bottom:12px">
          <div style="width:38px;height:38px;border-radius:9px;background:var(--brand-primary);color:#fff;display:grid;place-items:center;flex-shrink:0">
            <NIcon name="code" :size="19" color="#fff" />
          </div>
          <div style="flex:1">
            <div style="font-size:14px;font-weight:600;margin-bottom:2px">Один запрос — все записи</div>
            <div style="font-size:12.5px;color:var(--fg-2)">Создай список, опубликуй его, и фронт получит записи без хардкода ID.</div>
          </div>
          <NButton v-if="lists.length" variant="primary" size="md" @click="activeList = lists[0]">
            Открыть список
          </NButton>
        </div>
        <div style="background:#111827;border-radius:8px;padding:12px 16px">
          <pre style="margin:0;color:rgba(255,255,255,.9);font-size:12px;font-family:var(--font-mono);line-height:1.55;white-space:pre-wrap">{{ heroCodeSnippet() }}</pre>
        </div>
      </div>

      <!-- Loading -->
      <div v-if="loading" style="text-align:center;padding:32px;color:var(--fg-3);font-size:13px">Загрузка…</div>

      <template v-else>
        <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:12px">
          <div style="font-size:11px;color:var(--fg-3);text-transform:uppercase;letter-spacing:0.06em;font-weight:600">Списки · {{ lists.length }}</div>
          <NButton variant="secondary" size="sm" @click="showCreate = true">
            <NIcon name="plus" :size="13" />Новый список
          </NButton>
        </div>

        <!-- Empty state -->
        <div
          v-if="!lists.length"
          style="padding:40px;text-align:center;background:var(--bg-0);border:0.5px solid var(--border-default);border-radius:10px"
        >
          <NIcon name="folder" :size="28" color="var(--fg-3)" />
          <div style="font-size:14px;font-weight:600;margin-top:12px;margin-bottom:4px">Нет списков</div>
          <div style="font-size:13px;color:var(--fg-2);margin-bottom:16px">Создай первый список, чтобы опубликовать записи таблицы</div>
          <NButton variant="primary" size="md" @click="showCreate = true">
            <NIcon name="plus" :size="14" color="#fff" />Создать список
          </NButton>
        </div>

        <!-- List cards -->
        <div v-else style="display:grid;grid-template-columns:repeat(auto-fill,minmax(320px,1fr));gap:12px">
          <div
            v-for="l in lists"
            :key="l.id"
            @click="activeList = l"
            style="background:var(--bg-0);border:0.5px solid var(--border-default);border-radius:10px;padding:16px;cursor:pointer;transition:box-shadow 160ms,transform 160ms;display:flex;flex-direction:column;gap:12px"
            @mouseenter="(e) => { (e.currentTarget as HTMLElement).style.boxShadow = 'var(--shadow-2)'; (e.currentTarget as HTMLElement).style.transform = 'translateY(-2px)'; }"
            @mouseleave="(e) => { (e.currentTarget as HTMLElement).style.boxShadow = 'none'; (e.currentTarget as HTMLElement).style.transform = 'none'; }"
          >
            <div style="display:flex;align-items:flex-start;gap:12px">
              <div style="width:38px;height:38px;border-radius:9px;background:var(--brand-tint);color:var(--purple-600);display:grid;place-items:center;flex-shrink:0">
                <NIcon name="table" :size="19" />
              </div>
              <div style="flex:1;min-width:0">
                <div style="font-size:14px;font-weight:600;line-height:1.3">{{ l.slug }}</div>
                <div style="font-size:11.5px;color:var(--fg-3);margin-top:3px;display:flex;align-items:center;gap:5px">
                  <NIcon name="table" :size="10" />
                  <span>{{ tableName(l.table_slug) }}</span>
                </div>
              </div>
              <NBadge :tone="l.published ? 'success' : 'neutral'" :dot="true">
                {{ l.published ? 'публичный' : 'скрыт' }}
              </NBadge>
            </div>
            <div style="display:flex;align-items:center;gap:8px;padding-top:12px;border-top:0.5px solid var(--border-default)">
              <code style="flex:1;font-size:11px;font-family:var(--font-mono);color:var(--fg-2);background:var(--bg-1);padding:4px 8px;border-radius:4px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap">{{ listEndpoint(l) }}</code>
              <span style="font-size:12px;color:var(--fg-2);font-weight:500;white-space:nowrap">{{ fmtDate(l.created_at) }}</span>
              <NIcon name="chev" :size="14" color="var(--fg-3)" />
            </div>
          </div>
        </div>
      </template>
    </div>

    <!-- ─── ALL FILES (empty state) ──────────────────────────────────── -->
    <div v-else style="padding:20px 32px 56px;max-width:1080px">
      <div style="margin-bottom:16px">
        <h1 style="font-size:20px;font-weight:700;margin-bottom:4px">Все файлы</h1>
        <div style="font-size:12px;color:var(--fg-3)">Файлы, загруженные в поля записей</div>
      </div>
      <div style="padding:56px 20px;text-align:center;background:var(--bg-0);border:0.5px dashed var(--border-strong);border-radius:12px">
        <div style="width:48px;height:48px;border-radius:12px;background:var(--bg-2);display:grid;place-items:center;margin:0 auto 14px;color:var(--fg-3)">
          <NIcon name="folder" :size="22" />
        </div>
        <div style="font-size:15px;font-weight:700;margin-bottom:6px">Просмотр файлов</div>
        <div style="font-size:13px;color:var(--fg-2);max-width:360px;margin:0 auto">
          Файлы из полей типа «Файл» отображаются здесь. Загрузка файлов — через поля в таблицах или API генератора PDF.
        </div>
      </div>
    </div>
  </AppShell>

  <!-- ─── Create list modal ─────────────────────────────────────────── -->
  <NModal
    :open="showCreate"
    @close="closeCreate"
    title="Новый список"
    subtitle="Публикует записи таблицы по постоянному адресу."
    :width="480"
  >
    <div style="display:flex;flex-direction:column;gap:16px">
      <div>
        <label style="font-size:12px;color:var(--fg-2);font-weight:500;display:block;margin-bottom:6px">Таблица</label>
        <NSelect
          :model-value="newTableSlug"
          @update:model-value="onTablePick($event as string)"
          placeholder="— выбери таблицу —"
          :options="schemaStore.tables.map((t: { slug: string; name: string }) => ({ value: t.slug, label: t.name }))"
        />
      </div>
      <div>
        <label style="font-size:12px;color:var(--fg-2);font-weight:500;display:block;margin-bottom:6px">Slug</label>
        <NInput v-model="newSlugInput" placeholder="my-list" style="font-family:var(--font-mono)" />
        <div style="font-size:11px;color:var(--fg-3);margin-top:5px">
          Адрес: <code style="font-size:11px;font-family:var(--font-mono)">/lists/{{ slug }}/{{ newSlugInput || '…' }}</code>
        </div>
      </div>
      <div>
        <label style="font-size:12px;color:var(--fg-2);font-weight:500;display:block;margin-bottom:6px">Лимит записей</label>
        <NInput
          :model-value="newLimit !== undefined ? String(newLimit) : ''"
          @update:model-value="(v: string) => { newLimit = v ? Number(v) : undefined }"
          placeholder="50"
          type="number"
        />
      </div>
      <div>
        <label style="font-size:12px;color:var(--fg-2);font-weight:500;display:block;margin-bottom:8px">Доступ</label>
        <div style="display:flex;gap:8px">
          <button
            v-for="[v, t, h] in ([[true,'Публичный','Без ключа'],[false,'По ключу','Bearer-токен']] as [boolean,string,string][])"
            :key="String(v)"
            @click="newPublished = v"
            :style="{
              flex: 1, textAlign: 'left', padding: '10px 12px', borderRadius: '8px', cursor: 'pointer',
              background: newPublished === v ? 'var(--brand-tint)' : 'var(--bg-0)',
              border: `1px solid ${newPublished === v ? 'var(--brand-primary)' : 'var(--border-default)'}`,
              fontFamily: 'inherit',
            }"
          >
            <div :style="{ display: 'flex', alignItems: 'center', gap: '6px', fontSize: '13px', fontWeight: 600, color: newPublished === v ? 'var(--purple-700)' : 'var(--fg-1)' }">
              <NIcon :name="v ? 'extlink' : 'lock'" :size="13" />{{ t }}
            </div>
            <div style="font-size:11px;color:var(--fg-3);margin-top:3px">{{ h }}</div>
          </button>
        </div>
      </div>
      <p v-if="createError" style="color:var(--red-600,#dc2626);font-size:13px;margin:0">{{ createError }}</p>
    </div>
    <template #footer>
      <NButton variant="ghost" size="md" @click="closeCreate">Отмена</NButton>
      <NButton variant="primary" size="md" :disabled="createLoading" @click="handleCreateList">
        <NIcon name="check" :size="13" color="#fff" />{{ createLoading ? 'Создание…' : 'Создать список' }}
      </NButton>
    </template>
  </NModal>

  <NToast v-if="toast" tone="success" :title="toast" @close="toast = null" />
</template>
