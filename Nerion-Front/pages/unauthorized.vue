<script setup lang="ts">
definePageMeta({
  layout: 'auth',
  middleware: []
})

const route = useRoute()
const from = computed(() => (route.query.from as string) || '/')
const requestedPath = computed(() => (route.query.from as string) || route.fullPath)
</script>

<template>
  <div style="display: flex; flex-direction: column; align-items: center; text-align: center">
    <div style="position: relative; width: 64px; height: 64px; margin-bottom: 24px">
      <div style="position: absolute; inset: 0; border-radius: 50%; background: var(--brand-tint)" />
      <div style="position: absolute; inset: 0; display: grid; place-items: center; color: var(--brand-primary)">
        <NIcon name="lock" :size="26" />
      </div>
    </div>

    <h1 style="font-size: 26px; font-weight: 700; margin-bottom: 10px; letter-spacing: -0.02em">Нужно войти</h1>
    <p style="color: var(--fg-2); font-size: 14px; line-height: 1.55; margin-bottom: 24px; max-width: 340px">
      Эта страница доступна только участникам пространства. Войди в аккаунт, чтобы продолжить.
    </p>

    <div style="width: 100%; display: flex; align-items: center; gap: 12px; padding: 12px 14px; background: var(--bg-0); border: 0.5px solid var(--border-default); border-radius: 10px; margin-bottom: 24px; text-align: left">
      <div style="width: 34px; height: 34px; border-radius: 8px; background: var(--purple-100, #ede9fe); color: var(--purple-700, #6d28d9); display: grid; place-items: center; flex-shrink: 0">
        <NIcon name="box" :size="16" />
      </div>
      <div style="flex: 1; min-width: 0">
        <div style="font-size: 11px; color: var(--fg-3); margin-bottom: 3px">Запрошенный ресурс</div>
        <code style="font-size: 12px; font-family: var(--font-mono); color: var(--fg-1); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; display: block">{{ requestedPath }}</code>
      </div>
      <NBadge tone="warning" :dot="true">Закрыто</NBadge>
    </div>

    <div style="width: 100%; display: flex; flex-direction: column; gap: 10px">
      <NButton variant="primary" size="md" style="width: 100%" @click="navigateTo(`/login?from=${encodeURIComponent(from)}`)">Войти</NButton>
      <NButton variant="secondary" size="md" style="width: 100%" @click="navigateTo('/register')">Создать аккаунт</NButton>
    </div>

    <p style="margin-top: 22px; font-size: 12px; color: var(--fg-3); line-height: 1.5">
      Думаешь, это ошибка? Попроси владельца пространства
      <span style="color: var(--brand-primary)">выслать приглашение</span>.
    </p>
  </div>
</template>
