<script setup lang="ts">
definePageMeta({
  layout: 'auth',
  middleware: []
})

import { useAuthStore } from '~/stores/auth'

const auth = useAuthStore()
const route = useRoute()

const email = ref('')
const pwd = ref('')
const showPwd = ref(false)
const error = ref('')
const loading = ref(false)

async function submit() {
  error.value = ''
  loading.value = true
  try {
    await auth.login(email.value, pwd.value)
    const from = route.query.from as string
    // Guard against open redirect: only follow same-origin relative paths
    const safeTo = from && from.startsWith('/') && !from.startsWith('//') ? from : '/spaces'
    await navigateTo(safeTo)
  } catch (e: any) {
    error.value = e.message || 'Ошибка входа'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div>
    <h1 style="font-size: 26px; font-weight: 700; margin-bottom: 8px; letter-spacing: -0.02em">С возвращением</h1>
    <p style="color: var(--fg-2); font-size: 14px; margin-bottom: 28px; line-height: 1.5">Войди, чтобы продолжить работу с пространствами.</p>

    <form @submit.prevent="submit" style="display: flex; flex-direction: column; gap: 14px">
      <div>
        <label style="font-size: 12px; color: var(--fg-2); font-weight: 500; display: block; margin-bottom: 6px">Email</label>
        <NInput v-model="email" placeholder="you@university.ru" type="email" />
      </div>
      <div>
        <div style="display: flex; justify-content: space-between; margin-bottom: 6px">
          <label style="font-size: 12px; color: var(--fg-2); font-weight: 500">Пароль</label>
          <span @click="navigateTo('/reset')" style="font-size: 12px; color: var(--brand-primary); cursor: pointer">Забыл пароль?</span>
        </div>
        <div style="position: relative">
          <NInput v-model="pwd" :type="showPwd ? 'text' : 'password'" placeholder="••••••••" />
          <button
            type="button"
            @click="showPwd = !showPwd"
            :style="{
              position: 'absolute', right: '10px', top: '9px',
              background: 'none', border: '0', cursor: 'pointer', color: 'var(--fg-3)', padding: '2px',
            }"
          >
            <NIcon name="eye" :size="14" />
          </button>
        </div>
      </div>
      <p v-if="error" style="color: var(--red-600, #dc2626); font-size: 13px; margin: 0">{{ error }}</p>
      <NButton type="submit" variant="primary" size="md" :disabled="loading" style="width: 100%; margin-top: 4px">
        {{ loading ? 'Вход…' : 'Войти' }}
      </NButton>
      <div style="display: flex; align-items: center; gap: 10px; color: var(--fg-3); font-size: 12px">
        <div style="flex: 1; height: 0.5px; background: var(--border-default)" /> или
        <div style="flex: 1; height: 0.5px; background: var(--border-default)" />
      </div>
      <NButton variant="secondary" size="md" style="width: 100%">Войти через Яндекс ID</NButton>
    </form>

    <p style="margin-top: 24px; font-size: 13px; color: var(--fg-2); text-align: center">
      Нет аккаунта?
      <span @click="navigateTo('/register')" style="color: var(--brand-primary); font-weight: 500; cursor: pointer">Зарегистрироваться</span>
    </p>
  </div>
</template>
