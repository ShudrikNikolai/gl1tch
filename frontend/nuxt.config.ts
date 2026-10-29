// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
    ssr: false, // xterm.js работает только в браузере
    compatibilityDate: '2025-07-15',
    devtools: { enabled: false },
    modules: ['@pinia/nuxt'],
    css: ['~/assets/css/main.css'],

    app: {
        head: {
            title: 'GL1TCH',
            meta: [{ name: 'color-scheme', content: 'dark' }],
        },
    },

    // dev: /api/* уходит на gateway без префикса /api, как это делает reverse proxy
    routeRules: {
        '/api/**': { proxy: 'http://localhost:9091/**' }, //localhost
    },
})
