import { defineConfig } from 'vitest/config';
import solidPlugin from 'vite-plugin-solid';

export default defineConfig({
  plugins: [solidPlugin()],
  test: {
    environment: 'jsdom',
    globals: true,
    server: {
      deps: {
        inline: [/solid-js/, /lucide-solid/],
      },
    },
  },
  resolve: {
    conditions: ['development', 'browser'],
  },
});
