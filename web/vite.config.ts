import { createRequire } from 'node:module'
import { fileURLToPath, URL } from 'node:url'
import { defineConfig, type PluginOption } from 'vite'

const require = createRequire(import.meta.url)

function hasOptionalPackage(packageName: string) {
  try {
    require.resolve(packageName)
    return true
  } catch {
    return false
  }
}

function loadReactPlugin(): PluginOption {
  const pluginCandidates = ['@vitejs/plugin-react', '@vitejs/plugin-react-oxc']

  for (const pluginName of pluginCandidates) {
    try {
      const pluginModule = require(pluginName)
      const pluginFactory = pluginModule.default ?? pluginModule
      return pluginFactory()
    } catch {
      continue
    }
  }

  throw new Error('React plugin is missing. Install @vitejs/plugin-react for the current workspace.')
}

const hasLightningCss = hasOptionalPackage('lightningcss')
const reactPlugin = loadReactPlugin()

export default defineConfig({
  plugins: [reactPlugin],
  css: hasLightningCss
    ? {
        transformer: 'lightningcss',
      }
    : undefined,
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://localhost:18080',
        changeOrigin: true,
      },
      '/zashboard-ui': {
        target: 'http://localhost:18080',
        changeOrigin: true,
      },
    },
  },
  build: {
    cssMinify: hasLightningCss ? 'lightningcss' : true,
    target: 'es2022',
  },
})
