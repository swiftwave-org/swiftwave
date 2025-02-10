import { useAuthStore } from "~/stores/auth";

export default defineNuxtRouteMiddleware((to) => {
  const authStore = useAuthStore();
  console.log("Running auth middleware");

  // Initialize auth on first load
  authStore.initializeAuth();

  // Public routes that don't require authentication
  const publicRoutes = ["/login"];

  if (!authStore.isLoggedIn && !publicRoutes.includes(to.path)) {
    return navigateTo("/login", {
      replace: true,
      redirectCode: 401,
      external: false,
    });
  }

  if (authStore.isLoggedIn && to.path === "/login") {
    return navigateTo("/");
  }
});
