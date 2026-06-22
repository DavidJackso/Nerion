<script setup lang="ts">
definePageMeta({
  layout: 'auth'
})

const route = useRoute()
</script>

<template>
  <div style="display: flex; flex-direction: column; align-items: center; text-align: center">
    <div style="font-family: var(--font-mono); font-size: 76px; font-weight: 700; line-height: 1; letter-spacing: -0.04em; color: var(--brand-primary); margin-bottom: 20px">404</div>

    <h1 style="font-size: 24px; font-weight: 700; margin-bottom: 10px; letter-spacing: -0.02em">Страница не найдена</h1>
    <p style="color: var(--fg-2); font-size: 14px; line-height: 1.55; margin-bottom: 22px; max-width: 340px">
      Возможно, ссылку удалили, переименовали или ты ошибся в адресе.
    </p>

    <div style="width: 100%; display: flex; align-items: center; gap: 10px; padding: 10px 14px; background: var(--bg-0); border: 0.5px solid var(--border-default); border-radius: 10px; margin-bottom: 24px; text-align: left">
      <NIcon name="link" :size="15" style="flex-shrink: 0; color: var(--fg-3)" />
      <span style="flex: 1; min-width: 0; font-family: var(--font-mono); font-size: 12px; color: var(--fg-2); overflow: hidden; text-overflow: ellipsis; white-space: nowrap">{{ route.fullPath }}</span>
      <NBadge tone="neutral">404</NBadge>
    </div>

    <div style="width: 100%; display: flex; flex-direction: column; gap: 10px">
      <NButton variant="primary" size="md" style="width: 100%" @click="navigateTo('/spaces')">На главную</NButton>
      <NButton variant="ghost" size="md" style="width: 100%" @click="navigateTo('/spaces')">← Вернуться к пространствам</NButton>
    </div>

    <p style="margin-top: 22px; font-size: 12px; color: var(--fg-3); line-height: 1.5">
      Что-то сломалось?
      <span style="color: var(--brand-primary)">Написать в поддержку</span>
    </p>
  </div>
</template>
