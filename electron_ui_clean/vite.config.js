import { defineConfig } from 'vite';
import { svelte } from '@sveltejs/vite-plugin-svelte';

export default defineConfig({
  plugins: [
    svelte({
      compilerOptions: {
        // Svelte 5 で従来の `new App(...)` API を許可
        compatibility: {
          componentApi: 4,
        },
      },
    }),
  ],

  // Electronモジュールを外部化
  build: {
    rollupOptions: {
      external: ['electron'],
    },
  },

  // Electronを最適化から除外（Rolldown移行対応）
  optimizeDeps: {
    exclude: ['electron'],
    rolldownOptions: {
      external: ['electron'],
    },
  },
});