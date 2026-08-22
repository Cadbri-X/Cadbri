import { defineConfig } from 'vite';
import { svelte } from '@sveltejs/vite-plugin-svelte';
import tailwindcss from '@tailwindcss/vite';

export default defineConfig({
  plugins: [
    tailwindcss(),
    svelte()
  ],
  server: {
    port: 1111,
    host: '0.0.0.0',
    proxy: {
      '/api': {
        target: 'http://localhost:2222',
        changeOrigin: true,
        rewrite: (path: string) => path.replace(/^\/api/, '')
      },
      '/autocompleter': {
        target: 'http://localhost:2222',
        changeOrigin: true
      }
    }
  }
});
