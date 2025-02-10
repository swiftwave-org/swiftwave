// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  compatibilityDate: "2024-11-01",
  devtools: { enabled: false },
  modules: [
    "@nuxtjs/apollo",
    "@pinia/nuxt",
    "shadcn-nuxt",
    "@nuxtjs/tailwindcss",
    "@nuxtjs/color-mode",
    "nuxt-lucide-icons",
  ],
  colorMode: {
    classSuffix: "",
  },

  ssr: false,
  apollo: {
    clients: {
      default: {
        httpEndpoint:
          process.env.NUXT_PUBLIC_GRAPHQL_HTTP_BASE_URL! + "/graphql",
        wsEndpoint: process.env.NUXT_PUBLIC_GRAPHQL_WS_BASE_URL! + "/graphql",
        authHeader: "Authorization",
        authType: "Bearer",
        tokenStorage: "localStorage",
        tokenName: "token",
      },
    },
  },
  runtimeConfig: {
    public: {
      apiBaseUrl:
        process.env.NUXT_PUBLIC_HTTP_BASE_URL || "http://localhost:3000",
    },
  },
  shadcn: {
    /**
     * Prefix for all the imported component
     */
    prefix: "",
    /**
     * Directory that the component lives in.
     * @default "./components/ui"
     */
    componentDir: "./components/ui",
  },
  lucide: {
    namePrefix: "I",
  },
});
