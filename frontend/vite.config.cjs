import { defineConfig, loadEnv } from 'vite';
import solidPlugin from 'vite-plugin-solid';

const DEFAULT_API_PROXY_TARGET = 'http://127.0.0.1:8080';
const API_PROXY_PROBE_TIMEOUT_MS = 250;
const API_PROXY_MAX_FALLBACKS = 20;
const MAX_TCP_PORT = 65535;
const API_PROXY_REFRESH_INTERVAL_MS = 3000;

const probeApiStatus = async (target, { logFailures = false } = {}) => {
  if (typeof fetch !== 'function') {
    return false;
  }

  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), API_PROXY_PROBE_TIMEOUT_MS);
  try {
    const statusURL = new URL('/api/status', target);
    const response = await fetch(statusURL, {
      method: 'GET',
      signal: controller.signal,
    });
    if (!response.ok) {
      if (logFailures) {
        // eslint-disable-next-line no-console
        console.debug(`[vite] Backend probe returned HTTP ${response.status}.`);
      }
      return false;
    }
    return true;
  } catch (error) {
    if (logFailures) {
      const errorType = error instanceof Error ? error.name : typeof error;
      // eslint-disable-next-line no-console
      console.debug(`[vite] Backend probe request failed (${errorType}).`);
    }
    return false;
  } finally {
    clearTimeout(timeout);
  }
};

const resolveApiProxyTarget = async (explicitTarget, {
  logFailures = false,
  preferredTarget = '',
} = {}) => {
  const normalizedExplicit = explicitTarget?.trim();
  if (normalizedExplicit) {
    return normalizedExplicit;
  }

  const normalizedPreferred = preferredTarget?.trim();
  if (normalizedPreferred && await probeApiStatus(normalizedPreferred, { logFailures: false })) {
    return normalizedPreferred;
  }

  let defaultURL;
  try {
    defaultURL = new URL(DEFAULT_API_PROXY_TARGET);
  } catch (_) {
    return DEFAULT_API_PROXY_TARGET;
  }

  const defaultPort = Number(defaultURL.port || (defaultURL.protocol === 'https:' ? 443 : 80));
  if (!Number.isInteger(defaultPort) || defaultPort <= 0 || defaultPort > MAX_TCP_PORT) {
    return DEFAULT_API_PROXY_TARGET;
  }

  for (let offset = 0; offset <= API_PROXY_MAX_FALLBACKS; offset += 1) {
    const port = defaultPort + offset;
    if (port > MAX_TCP_PORT) {
      break;
    }
    const candidateURL = new URL(defaultURL.toString());
    candidateURL.port = String(port);
    const candidateTarget = candidateURL.origin;
    if (await probeApiStatus(candidateTarget, { logFailures })) {
      return candidateTarget;
    }
  }

  return DEFAULT_API_PROXY_TARGET;
};

export default defineConfig(async ({ mode }) => {
  const env = loadEnv(mode, process.cwd(), 'VITE_');
  const hasExplicitProxyTarget = Boolean(env.VITE_API_PROXY_TARGET?.trim());
  const shouldLogProbeFailures = mode !== 'production';
  const apiProxyTarget = await resolveApiProxyTarget(env.VITE_API_PROXY_TARGET, {
    logFailures: shouldLogProbeFailures,
  });
  let dynamicApiProxyTarget = apiProxyTarget;
  let refreshInFlight = null;
  let refreshTimer = null;

  const clearRefreshTimer = () => {
    if (refreshTimer) {
      clearInterval(refreshTimer);
      refreshTimer = null;
    }
  };

  const refreshDynamicApiProxyTarget = async (logFailures = false) => {
    if (hasExplicitProxyTarget) {
      return dynamicApiProxyTarget;
    }
    if (refreshInFlight) {
      return refreshInFlight;
    }

    refreshInFlight = resolveApiProxyTarget('', {
      logFailures,
      preferredTarget: dynamicApiProxyTarget,
    })
      .then((nextTarget) => {
        dynamicApiProxyTarget = nextTarget;
        return nextTarget;
      })
      .finally(() => {
        refreshInFlight = null;
      });

    return refreshInFlight;
  };

  const logRefreshFailure = (error) => {
    if (!shouldLogProbeFailures) {
      return;
    }
    const errorType = error instanceof Error ? error.message : typeof error;
    // eslint-disable-next-line no-console
    console.debug(`[vite] Failed to refresh API proxy target (${errorType}).`);
  };

  const refreshAndApplyTarget = (applyCurrentTarget, logFailures = false) => {
    void refreshDynamicApiProxyTarget(logFailures)
      .then(() => {
        applyCurrentTarget();
      })
      .catch((error) => {
        logRefreshFailure(error);
      });
  };

  if (!hasExplicitProxyTarget && apiProxyTarget !== DEFAULT_API_PROXY_TARGET) {
    // eslint-disable-next-line no-console
    console.info(`[vite] Auto-detected backend at ${apiProxyTarget}. Set VITE_API_PROXY_TARGET to override.`);
  }

  return {
    plugins: [
      solidPlugin(),
      {
        name: 'dynamic-api-proxy-refresh-cleanup',
        configureServer(server) {
          const stopRefresh = () => {
            clearRefreshTimer();
          };
          server.httpServer?.once('close', stopRefresh);
          server.watcher.once('close', stopRefresh);
        },
        closeBundle() {
          clearRefreshTimer();
        },
      },
    ],
    server: {
      proxy: {
        '/api': {
          target: dynamicApiProxyTarget,
          changeOrigin: true,
          configure(proxy, options) {
            if (hasExplicitProxyTarget) {
              return;
            }

            const applyCurrentTarget = () => {
              if (options.target !== dynamicApiProxyTarget) {
                options.target = dynamicApiProxyTarget;
              }
            };

            applyCurrentTarget();

            proxy.on('proxyReq', () => {
              applyCurrentTarget();
            });

            proxy.on('error', (error) => {
              const code = error && typeof error === 'object' ? error.code : '';
              if (typeof code === 'string' && ['ECONNREFUSED', 'ECONNRESET', 'EHOSTUNREACH', 'ETIMEDOUT'].includes(code)) {
                refreshAndApplyTarget(applyCurrentTarget, shouldLogProbeFailures);
              }
            });

            if (!refreshTimer) {
              refreshTimer = setInterval(() => {
                refreshAndApplyTarget(applyCurrentTarget, false);
              }, API_PROXY_REFRESH_INTERVAL_MS);
              if (typeof refreshTimer.unref === 'function') {
                refreshTimer.unref();
              }
            }
          },
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
