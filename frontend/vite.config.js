import { defineConfig, loadEnv } from 'vite';
import solidPlugin from 'vite-plugin-solid';

const DEFAULT_API_PROXY_TARGET = 'http://127.0.0.1:8080';

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), 'VITE_');
  const apiProxyTarget = env.VITE_API_PROXY_TARGET?.trim() || DEFAULT_API_PROXY_TARGET;

  return {
    plugins: [solidPlugin()],
    server: {
      proxy: {
        '/api': {
          target: apiProxyTarget,
          changeOrigin: true,
        },
      },
    },
    build: {
      // Output directly to the Go backend's asset folder
      outDir: '../internal/web/assets',
      emptyOutDir: true, // Clears the folder before building
      rollupOptions: {
        output: {
          entryFileNames: 'app.js',
          assetFileNames: (assetInfo) => {
            if (assetInfo.name === 'style.css') return 'styles.css';
            return '[name][extname]';
          },
        },
      },
    },
  };
});
