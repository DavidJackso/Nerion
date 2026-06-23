<script setup lang="ts">
// Auth layout — wraps login / register / reset pages in a centered card design
</script>

<template>
  <div
    style="min-height: 100vh; background: var(--bg-1); display: grid; grid-template-rows: auto 1fr auto"
  >
    <header style="padding: 28px 40px; display: flex; align-items: center; gap: 10px">
      <div
        style="
          width: 28px;
          height: 28px;
          border-radius: 7px;
          background: var(--brand-primary);
          color: #fff;
          display: grid;
          place-items: center;
          font-weight: 700;
          font-size: 14px;
        "
      >N</div>
      <span
        style="font-size: 16px; font-weight: 700; letter-spacing: -0.01em; color: var(--fg-1)"
      >Nerion</span>
    </header>

    <main style="display: grid; place-items: center; padding: 0 24px 48px">
      <div style="width: 100%; max-width: 400px; animation: slideUp 220ms both">
        <slot />
      </div>
    </main>

    <footer
      style="
        padding: 20px 40px;
        display: flex;
        justify-content: space-between;
        color: var(--fg-3);
        font-size: 12px;
      "
    >
      <span>© Nerion · 2026</span><span>Помощь</span>
    </footer>
  </div>
</template>
