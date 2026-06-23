export default defineNuxtConfig({
  ssr: false,
  modules: ['@pinia/nuxt'],
  css: ['~/assets/style.css'],
  runtimeConfig: {
    public: {
      apiUrl: ''
    }
  },
  imports: {
    dirs: ['stores', 'api']
  }
})
