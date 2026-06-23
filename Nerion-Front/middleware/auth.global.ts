export default defineNuxtRouteMiddleware((to) => {
  // Only run on the client side (localStorage is not available server-side)
  if (!import.meta.client) return

  // Routes that don't require authentication
  const publicPaths = ['/', '/login', '/register', '/reset', '/unauthorized']
  const isPublic =
    publicPaths.includes(to.path) ||
    to.matched.length === 0 // 404 — treat as public

  const token = localStorage.getItem('nerion_access_token')

  if (isPublic) {
    // Redirect already-authenticated users away from login/register to /spaces
    if (token && (to.path === '/login' || to.path === '/register')) {
      return navigateTo('/spaces')
    }
    // On landing page, authenticated users go to /spaces
    if (token && to.path === '/') {
      return navigateTo('/spaces')
    }
    return
  }

  // Protected route — must have a token
  if (!token) {
    return navigateTo({ path: '/unauthorized', query: { from: to.fullPath } })
  }
})
