import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// https://vite.dev/config/
export default defineConfig({
  plugins: [vue()],
  server: {
    proxy: {
      // override with VITE_API_TARGET when the backend lives elsewhere
      // (e.g. the docker-compose nginx at http://localhost:8089)
      '/api': process.env.VITE_API_TARGET || 'http://localhost:8080',
    },
  },
})
