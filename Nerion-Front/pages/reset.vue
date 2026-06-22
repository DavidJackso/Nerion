<script setup lang="ts">
definePageMeta({
  layout: 'auth',
  middleware: []
})

import { resetRequest } from '~/api/auth'

const email = ref('')
const sent = ref(false)
const loading = ref(false)
const error = ref('')

async function submit() {
  error.value = ''
  loading.value = true
  try {
    await resetRequest(email.value)
    sent.value = true
  } catch (e: any) {
    error.value = e.message || 'Ошибка'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div>
    <template v-if="sent">
      <div style="text-align: center; padding: 32px 0">
        <div style="width: 56px; height: 56px; border-radius: 50%; background: var(--green-100); color: var(--green-700); display: grid; place-items: center; margin: 0 auto 20px">
          <NIcon name="check" :size="24" />
        </div>
        <h1 style="font-size: 22px; font-weight: 700; margin-bottom: 10px; letter-spacing: -0.01em">Письмо отправлено</h1>
        <p style="color: var(--fg-2); font-size: 14px; margin-bottom: 28px">
          Проверь <strong>{{ email }}</strong> — там ссылка для сброса пароля.
        </p>
        <NButton variant="secondary" size="md" @click="navigateTo('/login')" style="width: 100%">← Назад к входу</NButton>
      </div>
    </template>

    <template v-else>
      <h1 style="font-size: 26px; font-weight: 700; margin-bottom: 8px; letter-spacing: -0.02em">Сброс пароля</h1>
      <p style="color: var(--fg-2); font-size: 14px; margin-bottom: 28px">Пришлём ссылку для сброса на email.</p>
      <form @submit.prevent="submit" style="display: flex; flex-direction: column; gap: 14px">
        <div>
          <label style="font-size: 12px; color: var(--fg-2); font-weight: 500; display: block; margin-bottom: 6px">Email</label>
          <NInput v-model="email" placeholder="you@university.ru" type="email" />
        </div>
        <p v-if="error" style="color: var(--red-600, #dc2626); font-size: 13px; margin: 0">{{ error }}</p>
        <NButton type="submit" variant="primary" size="md" :disabled="loading" style="width: 100%; margin-top: 4px">
          {{ loading ? 'Отправка…' : 'Отправить ссылку' }}
        </NButton>
      </form>
      <p style="margin-top: 24px; font-size: 13px; color: var(--fg-2); text-align: center">
        <span @click="navigateTo('/login')" style="color: var(--brand-primary); font-weight: 500; cursor: pointer">← Назад к входу</span>
      </p>
    </template>
  </div>
</template>
