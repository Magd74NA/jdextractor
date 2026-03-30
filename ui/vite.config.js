import { defineConfig } from 'vite'
import { svelte } from '@sveltejs/vite-plugin-svelte'
import { compression } from 'vite-plugin-compression2'
import { resolve } from 'path'

export default defineConfig(({ mode }) => ({
  plugins: [
    svelte(),
    ...(mode === 'production'
      ? [compression({ algorithm: 'gzip' })]
      : []),
  ],
  build: {
    outDir: resolve(__dirname, '../jdextract/web/dist'),
    emptyOutDir: true,
    minify: mode === 'production',
  },
  server: {
    proxy: {
      '/api': 'http://localhost:8080',
    },
  },
}))
