<script setup lang="ts">
import { useSpacesStore } from '~/stores/spaces'
import { useSpaceSlug } from '~/composables/useSpaceSlug'
import { useToast } from '~/composables/useToast'

definePageMeta({ layout: 'app', middleware: 'auth' })

const spacesStore = useSpacesStore()
const { slug: routeSlug, space } = useSpaceSlug()
const { toast, show } = useToast(2500)

const spaceName = ref('')
const dangerInput = ref('')
const saveLoading = ref(false)
const deleteLoading = ref(false)
const error = ref('')

onMounted(() => {
  if (space.value) spaceName.value = space.value.name
})

async function save() {
  error.value = ''
  saveLoading.value = true
  try {
    await spacesStore.renameSpace(routeSlug.value, spaceName.value)
    show('Настройки сохранены')
  } catch (e: any) {
    error.value = e.message
  } finally {
    saveLoading.value = false
  }
}

async function deleteSpace() {
  deleteLoading.value = true
  try {
    await spacesStore.deleteSpace(routeSlug.value, dangerInput.value)
    await navigateTo('/spaces')
  } catch (e: any) {
    error.value = e.message
    deleteLoading.value = false
  }
}

const breadcrumb = computed(() => [space.value?.name || routeSlug.value, 'Настройки'])
</script>

<template>
  <div style="max-width: 700px; margin: 0 auto; padding: 32px 32px 80px">
    <h1 style="font-size: 24px; font-weight: 700; margin-bottom: 28px; letter-spacing: -0.01em">Настройки пространства</h1>

    <!-- Main settings -->
    <div style="background: var(--bg-0); border: 0.5px solid var(--border-default); border-radius: 8px; padding: 24px; margin-bottom: 16px">
      <div style="font-size: 14px; font-weight: 600; margin-bottom: 18px">Основное</div>
      <div style="display: flex; flex-direction: column; gap: 18px">
        <div>
          <label style="font-size: 12px; color: var(--fg-2); font-weight: 500; display: block; margin-bottom: 6px">Название пространства</label>
          <NInput v-model="spaceName" />
        </div>
        <div>
          <label style="font-size: 12px; color: var(--fg-2); font-weight: 500; display: block; margin-bottom: 6px">Slug</label>
          <div style="display: flex; align-items: center; height: 36px; border: 0.5px solid var(--border-strong); border-radius: 6px; overflow: hidden; font-family: var(--font-mono); font-size: 13px">
            <span style="padding: 0 10px; background: var(--bg-2); color: var(--fg-3); height: 100%; display: flex; align-items: center; border-right: 0.5px solid var(--border-default); white-space: nowrap">app.nerion.ru/</span>
            <span style="flex: 1; padding: 0 10px; color: var(--fg-1)">{{ routeSlug }}</span>
          </div>
          <p style="font-size: 11px; color: var(--fg-3); margin-top: 5px">Slug нельзя изменить. Изменение потребует обновления всех интеграций.</p>
        </div>
      </div>
      <p v-if="error" style="color: var(--red-600, #dc2626); font-size: 13px; margin-top: 12px">{{ error }}</p>
      <div style="margin-top: 20px; display: flex; justify-content: flex-end">
        <NButton variant="primary" size="md" :disabled="saveLoading" @click="save">
          {{ saveLoading ? 'Сохранение…' : 'Сохранить' }}
        </NButton>
      </div>
    </div>

    <!-- Danger zone -->
    <div style="border: 0.5px solid #FECACA; background: #FFF5F5; border-radius: 8px; padding: 24px">
      <div style="font-size: 13px; font-weight: 600; color: #B91C1C; margin-bottom: 16px; display: flex; align-items: center; gap: 8px">
        <NIcon name="warn" :size="14" color="#B91C1C" />
        Опасная зона
      </div>
      <div style="display: flex; align-items: center; justify-content: space-between; padding: 14px 0; border-top: 0.5px solid #FECACA">
        <div>
          <div style="font-size: 13px; font-weight: 500; color: var(--fg-1)">Удалить пространство</div>
          <div style="font-size: 12px; color: var(--fg-2); margin-top: 2px">Все таблицы, записи и API-ключи будут удалены без возможности восстановления.</div>
        </div>
      </div>
      <div style="margin-top: 10px">
        <label style="font-size: 12px; color: var(--fg-2); font-weight: 500; display: block; margin-bottom: 6px">
          Введи <code style="font-size: 11px; font-family: var(--font-mono)">{{ routeSlug }}</code> для подтверждения
        </label>
        <div style="display: flex; gap: 8px">
          <NInput v-model="dangerInput" :placeholder="routeSlug" style="flex: 1" />
          <NButton
            variant="danger"
            size="md"
            :disabled="dangerInput !== routeSlug || deleteLoading"
            @click="deleteSpace"
          >
            {{ deleteLoading ? 'Удаление…' : 'Удалить пространство' }}
          </NButton>
        </div>
      </div>
    </div>
  </div>

  <NToast v-if="toast" tone="success" :title="toast" @close="toast = null" />
</template>
