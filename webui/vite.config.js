import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  define: {
    // Do not modify this constant, it is used in the evaluation.
    // To test from another device, temporarily replace localhost with your LAN IP.
    __API_URL__: JSON.stringify('http://localhost:3000')
  },
  server: {
    port: 5173
  },
  build: {
    outDir: 'dist'
  }
})
