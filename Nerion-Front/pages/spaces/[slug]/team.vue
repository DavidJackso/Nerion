<script setup lang="ts">
import { useAuthStore } from '~/stores/auth'
import { useSpaceSlug } from '~/composables/useSpaceSlug'
import { useToast } from '~/composables/useToast'
import { listMembers, inviteMember, removeMember } from '~/api/members'

definePageMeta({ layout: 'app', middleware: 'auth' })

interface SpaceMember {
  id: string
  user_id: string
  name: string
  email: string
  role: 'admin' | 'member'
  joined_at: string
}

const authStore = useAuthStore()
const { slug, space } = useSpaceSlug()
const { toast, show } = useToast(2500)

const members = ref<SpaceMember[]>([])
const showInvite = ref(false)
const inviteEmail = ref('')
const error = ref('')

onMounted(async () => {
  try {
    members.value = await listMembers(slug.value)
  } catch (e: any) {
    show(e.message)
  }
})

async function invite() {
  error.value = ''
  try {
    await inviteMember(slug.value, inviteEmail.value)
    showInvite.value = false
    inviteEmail.value = ''
    show('Приглашение отправлено')
    members.value = await listMembers(slug.value)
  } catch (e: any) {
    error.value = e.message
  }
}

async function remove(userId: string) {
  try {
    await removeMember(slug.value, userId)
    members.value = members.value.filter(m => m.user_id !== userId)
    show('Участник удалён')
  } catch (e: any) {
    show(e.message)
  }
}

function isMe(m: SpaceMember): boolean {
  return m.user_id === authStore.user?.id
}

function getInitials(name: string): string {
  return String(name || '').split(' ').map((p: string) => p[0]).join('').slice(0, 2).toUpperCase()
}

const PALETTE = [
  { bg: 'var(--purple-200)', fg: 'var(--purple-700)' },
  { bg: '#DBEAFE', fg: '#1E40AF' },
  { bg: '#FFE4E6', fg: '#9F1239' },
  { bg: '#D1FAE5', fg: 'var(--green-700)' },
]
function palette(i: number) { return PALETTE[i % PALETTE.length] }

const breadcrumb = computed(() => [space.value?.name || slug.value, 'Настройки', 'Команда'])
</script>

<template>
  <div style="max-width: 860px; margin: 0 auto; padding: 32px 32px 80px">
    <div style="margin-bottom: 24px; display: flex; align-items: flex-start; justify-content: space-between">
      <div>
        <h1 style="font-size: 24px; font-weight: 700; margin-bottom: 6px; letter-spacing: -0.01em">Команда</h1>
        <p style="font-size: 13px; color: var(--fg-2)">{{ members.length }} участников</p>
      </div>
      <NButton variant="primary" size="sm" @click="showInvite = true">
        <NIcon name="plus" :size="14" color="#fff" />
        Добавить участника
      </NButton>
    </div>

    <div style="background: var(--bg-0); border: 0.5px solid var(--border-default); border-radius: 8px; overflow: hidden">
      <div v-if="!members.length" style="padding: 32px; text-align: center; color: var(--fg-3); font-size: 13px">
        Нет участников
      </div>
      <div
        v-for="(m, i) in members"
        :key="m.user_id"
        :style="{
          display: 'flex', alignItems: 'center', padding: '12px 16px', gap: '12px',
          borderTop: i ? '0.5px solid var(--border-default)' : '0',
        }"
      >
        <div :style="{
          width: '32px', height: '32px', borderRadius: '50%',
          background: palette(i).bg, color: palette(i).fg,
          display: 'grid', placeItems: 'center',
          fontWeight: '700', fontSize: '12px', flexShrink: '0',
        }">{{ getInitials(m.name) }}</div>
        <div style="flex: 1; min-width: 0">
          <div style="font-size: 13px; font-weight: 500; display: flex; align-items: center; gap: 6px">
            {{ m.name }}
            <NBadge v-if="isMe(m)" tone="brand">это ты</NBadge>
          </div>
          <div style="font-size: 11px; color: var(--fg-3); font-family: var(--font-mono)">{{ m.email }}</div>
        </div>
        <NBadge :tone="m.role === 'admin' ? 'brand' : 'neutral'">{{ m.role === 'admin' ? 'Admin' : 'Member' }}</NBadge>
        <button
          :disabled="isMe(m)"
          @click="!isMe(m) && remove(m.user_id)"
          :style="{
            background: '0', border: '0', padding: '6px', display: 'flex',
            cursor: isMe(m) ? 'default' : 'pointer',
            color: isMe(m) ? 'var(--neutral-300)' : 'var(--fg-3)',
          }"
        >
          <NIcon name="trash" :size="14" />
        </button>
      </div>
    </div>

    <!-- Role descriptions -->
    <div style="margin-top: 24px; padding: 16px 20px; background: var(--bg-0); border: 0.5px solid var(--border-default); border-radius: 8px">
      <div style="font-size: 13px; font-weight: 600; margin-bottom: 10px">Что умеет каждая роль</div>
      <div style="display: grid; grid-template-columns: 100px 1fr; gap: 8px 16px; font-size: 12px">
        <div style="color: var(--fg-1); font-weight: 500">Admin</div>
        <div style="color: var(--fg-2)">Всё: таблицы, API-ключи, биллинг, состав команды, настройки пространства.</div>
        <div style="color: var(--fg-1); font-weight: 500">Member</div>
        <div style="color: var(--fg-2)">Чтение и редактирование записей, генерация PDF.</div>
      </div>
    </div>
  </div>

  <!-- Invite modal -->
  <NModal :open="showInvite" @close="showInvite = false; error = ''" title="Пригласить участника" :width="420">
    <div style="display: flex; flex-direction: column; gap: 14px">
      <div>
        <label style="font-size: 12px; color: var(--fg-2); font-weight: 500; display: block; margin-bottom: 6px">Email</label>
        <NInput v-model="inviteEmail" type="email" placeholder="colleague@university.ru" />
      </div>
      <p v-if="error" style="color: var(--red-600, #dc2626); font-size: 13px; margin: 0">{{ error }}</p>
    </div>
    <template #footer>
      <NButton variant="ghost" size="md" @click="showInvite = false; error = ''">Отмена</NButton>
      <NButton variant="primary" size="md" @click="invite">Отправить приглашение</NButton>
    </template>
  </NModal>

  <NToast v-if="toast" tone="success" :title="toast" @close="toast = null" />
</template>
