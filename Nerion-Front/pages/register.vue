<script setup lang="ts">
definePageMeta({
  layout: 'auth',
  middleware: []
})

import { register } from '~/api/auth'

const name = ref('')
const email = ref('')
const pwd = ref('')
const error = ref('')
const loading = ref(false)

async function submit() {
  error.value = ''
  loading.value = true
  try {
    await register(name.value, email.value, pwd.value)
    await navigateTo('/login?registered=true')
  } catch (e: any) {
    error.value = e.message || 'Ошибка регистрации'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div>
    <h1 style="font-size: 26px; font-weight: 700; margin-bottom: 8px; letter-spacing: -0.02em">Создать аккаунт</h1>
    <p style="color: var(--fg-2); font-size: 14px; margin-bottom: 28px">14 дней Pro бесплатно, карта не нужна.</p>

    <form @submit.prevent="submit" style="display: flex; flex-direction: column; gap: 14px">
      <div>
        <label style="font-size: 12px; color: var(--fg-2); font-weight: 500; display: block; margin-bottom: 6px">Имя</label>
        <NInput v-model="name" placeholder="Как тебя зовут" />
      </div>
      <div>
        <label style="font-size: 12px; color: var(--fg-2); font-weight: 500; display: block; margin-bottom: 6px">Email</label>
        <NInput v-model="email" placeholder="you@university.ru" type="email" />
      </div>
      <div>
        <label style="font-size: 12px; color: var(--fg-2); font-weight: 500; display: block; margin-bottom: 6px">Пароль</label>
        <NInput v-model="pwd" type="password" placeholder="Минимум 8 символов" />
      </div>
      <p v-if="error" style="color: var(--red-600, #dc2626); font-size: 13px; margin: 0">{{ error }}</p>
      <NButton type="submit" variant="primary" size="md" :disabled="loading" style="width: 100%; margin-top: 4px">
        {{ loading ? 'Регистрация…' : 'Зарегистрироваться' }}
      </NButton>
    </form>

    <p style="margin-top: 24px; font-size: 13px; color: var(--fg-2); text-align: center">
      Уже есть аккаунт?
      <span @click="navigateTo('/login')" style="color: var(--brand-primary); font-weight: 500; cursor: pointer">Войти</span>
    </p>
    <p style="margin-top: 16px; font-size: 11px; color: var(--fg-3); text-align: center; line-height: 1.5">
      Регистрируясь, ты соглашаешься с условиями использования и политикой конфиденциальности.
    </p>
  </div>
</template>
