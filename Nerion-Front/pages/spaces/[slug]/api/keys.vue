<script setup lang="ts">
definePageMeta({ layout: 'app', middleware: 'auth' })

import { listKeys, createKey, revokeKey } from '~/api/apikeys'
import { useSpacesStore } from '~/stores/spaces'
import { useSpaceSlug } from '~/composables/useSpaceSlug'
import { useToast } from '~/composables/useToast'

interface ApiKey {
  id: string
  name: string
  scope: 'read' | 'write'
  prefix?: string
  key?: string
  created_at: string
  last_used_at?: string
}

const { slug, space } = useSpaceSlug()
const { toast, show } = useToast(2500)

const keys = ref<ApiKey[]>([])
const freshKey = ref<string | null>(null)
const showCreate = ref(false)
const newKeyName = ref('')
const newKeyScope = ref<'read' | 'write'>('read')
const error = ref('')
const revokeConfirmId = ref<string | null>(null)

onMounted(async () => {
  try {
    keys.value = await listKeys(slug.value)
  } catch (e: any) {
    show(e.message)
  }
})

async function handleCreateKey() {
  error.value = ''
  try {
    const res: ApiKey = await createKey(slug.value, newKeyName.value, newKeyScope.value)
    freshKey.value = res.key ?? null
    keys.value.push(res)
    showCreate.value = false
    newKeyName.value = ''
    newKeyScope.value = 'read'
    show('Ключ создан — скопируй сейчас')
  } catch (e: any) {
    error.value = e.message
  }
}

async function handleRevokeKey(id: string) {
  revokeConfirmId.value = null
  try {
    await revokeKey(slug.value, id)
    keys.value = keys.value.filter(k => k.id !== id)
    show('Ключ удалён')
  } catch (e: any) {
    show(e.message)
  }
}

function copyKey() {
  if (!freshKey.value) return
  navigator.clipboard.writeText(freshKey.value).then(() => show('Ключ скопирован'))
}

function fmtKeyDate(dt?: string): string {
  if (!dt) return 'никогда'
  return new Intl.DateTimeFormat('ru-RU', { day: '2-digit', month: 'short', year: 'numeric' }).format(new Date(dt))
}

const breadcrumb = computed(() => [space.value?.name || slug.value, 'REST API', 'Ключи'])
</script>

<template>
  <div :data-breadcrumb="JSON.stringify(breadcrumb)" style="max-width: 880px; margin: 0 auto; padding: 32px 32px 80px">
    <div style="display: flex; align-items: flex-end; justify-content: space-between; margin-bottom: 24px">
      <div>
        <h1 style="font-size: 24px; font-weight: 700; margin-bottom: 6px; letter-spacing: -0.01em">Ключи API</h1>
        <p style="font-size: 13px; color: var(--fg-2)">
          Используй в заголовке
          <code style="font-size: 12px; font-family: var(--font-mono)">X-Api-Key: …</code>
        </p>
      </div>
      <NButton variant="primary" size="md" @click="showCreate = true">
        <NIcon name="plus" :size="14" color="#fff" />
        Создать ключ
      </NButton>
    </div>

    <!-- Fresh key banner -->
    <div
      v-if="freshKey"
      style="background: var(--bg-0); border: 1.5px solid var(--brand-primary); border-radius: 8px; padding: 16px 20px; margin-bottom: 20px"
    >
      <div style="display: flex; align-items: center; gap: 10px; margin-bottom: 10px">
        <NIcon name="key" :size="14" color="var(--purple-600)" />
        <span style="font-size: 13px; font-weight: 600">Ключ создан — скопируй сейчас</span>
        <NBadge tone="warning" :dot="true">видно один раз</NBadge>
      </div>
      <div style="display: flex; gap: 8px; align-items: center">
        <code
          style="flex: 1; padding: 10px 14px; background: var(--bg-2); border-radius: 6px; font-size: 12px; color: var(--fg-1); font-family: var(--font-mono); border: 0.5px solid var(--border-default); overflow: hidden; text-overflow: ellipsis; white-space: nowrap"
        >{{ freshKey }}</code>
        <NButton variant="primary" size="md" @click="copyKey">
          <NIcon name="copy" :size="13" color="#fff" />
          Копировать
        </NButton>
      </div>
    </div>

    <!-- Keys table -->
    <div style="background: var(--bg-0); border: 0.5px solid var(--border-default); border-radius: 8px; overflow: hidden">
      <div
        style="display: grid; grid-template-columns: 1fr 1fr 120px 180px 80px; padding: 0 16px; height: 38px; align-items: center; background: var(--bg-1); border-bottom: 0.5px solid var(--border-default); font-size: 10px; color: var(--fg-3); text-transform: uppercase; letter-spacing: 0.06em; font-weight: 600"
      >
        <div>Имя</div>
        <div>Ключ</div>
        <div>Scope</div>
        <div>Последнее использование</div>
        <div></div>
      </div>

      <div
        v-if="!keys.length"
        style="padding: 32px; text-align: center; color: var(--fg-3); font-size: 13px"
      >
        Нет ключей. Создай первый.
      </div>

      <div
        v-for="(k, i) in keys"
        :key="k.id"
        :style="{
          display: 'grid',
          gridTemplateColumns: '1fr 1fr 120px 180px 80px',
          padding: '12px 16px',
          alignItems: 'center',
          borderTop: i ? '0.5px solid var(--border-default)' : '0',
        }"
      >
        <div style="font-size: 13px; font-weight: 500; color: var(--fg-1)">{{ k.name }}</div>
        <div>
          <code style="font-size: 11px; font-family: var(--font-mono); color: var(--fg-2)">{{ k.prefix }}…</code>
        </div>
        <div>
          <NBadge :tone="k.scope === 'write' ? 'warning' : 'neutral'">{{ k.scope }}</NBadge>
        </div>
        <div style="font-size: 13px; color: var(--fg-2)">{{ fmtKeyDate(k.last_used_at) }}</div>
        <div style="display: flex; gap: 4px; justify-content: flex-end">
          <button
            v-if="revokeConfirmId !== k.id"
            @click="revokeConfirmId = k.id"
            style="background: 0; border: 0; padding: 5px; cursor: pointer; color: var(--red-500); display: flex"
            title="Отозвать ключ"
          >
            <NIcon name="trash" :size="13" />
          </button>
          <template v-else>
            <button
              @click="handleRevokeKey(k.id)"
              style="background: var(--red-500); border: 0; padding: 4px 8px; cursor: pointer; color: #fff; border-radius: 4px; font-size: 11px; font-family: inherit"
            >Удалить</button>
            <button
              @click="revokeConfirmId = null"
              style="background: 0; border: 0; padding: 4px 8px; cursor: pointer; color: var(--fg-2); font-size: 11px; font-family: inherit"
            >Отмена</button>
          </template>
        </div>
      </div>
    </div>

    <!-- Security note -->
    <div
      style="margin-top: 20px; padding: 14px 16px; background: #FFFBEB; border: 0.5px solid #FDE68A; border-radius: 8px; display: flex; gap: 10px; align-items: flex-start"
    >
      <NIcon name="lock" :size="15" color="#92400E" style="flex-shrink: 0; margin-top: 1px" />
      <div style="font-size: 12px; color: #78350F; line-height: 1.55">
        <strong>Безопасность.</strong> Ключи держи в переменных окружения, никогда — в коде.
      </div>
    </div>
  </div>

  <!-- Create key modal -->
  <NModal :open="showCreate" @close="showCreate = false; error = ''" title="Новый ключ API" :width="440">
    <div style="display: flex; flex-direction: column; gap: 16px">
      <div>
        <label style="font-size: 12px; color: var(--fg-2); font-weight: 500; display: block; margin-bottom: 6px">Название</label>
        <NInput v-model="newKeyName" placeholder="Production · LMS" />
      </div>
      <div>
        <label style="font-size: 12px; color: var(--fg-2); font-weight: 500; display: block; margin-bottom: 8px">Права доступа</label>
        <div style="display: flex; gap: 8px">
          <label
            v-for="[val, label] in ([['read', 'Только чтение'], ['write', 'Чтение и запись']] as [string, string][])"
            :key="val"
            :style="{
              display: 'flex', alignItems: 'center', gap: '6px', cursor: 'pointer',
              fontSize: '13px', padding: '8px 12px', borderRadius: '6px', flex: '1',
              border: '0.5px solid',
              background: newKeyScope === val ? 'var(--brand-tint)' : 'var(--bg-0)',
              borderColor: newKeyScope === val ? 'var(--purple-300)' : 'var(--border-default)',
            }"
          >
            <input type="radio" :value="val" v-model="newKeyScope" style="accent-color: var(--brand-primary)" />
            {{ label }}
          </label>
        </div>
      </div>
      <p v-if="error" style="color: var(--red-600, #dc2626); font-size: 13px; margin: 0">{{ error }}</p>
    </div>
    <template #footer>
      <NButton variant="ghost" size="md" @click="showCreate = false; error = ''">Отмена</NButton>
      <NButton variant="primary" size="md" @click="handleCreateKey">Создать ключ</NButton>
    </template>
  </NModal>

  <NToast v-if="toast" tone="success" :title="toast" @close="toast = null" />
</template>
