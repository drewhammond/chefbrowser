import path from 'node:path'
import {defineConfig} from 'vite'

export default defineConfig({
  // root: path.resolve(import.meta.dirname, 'src'),

  // must be set to the UI root
  base: "",
  appType: "custom",
  resolve: {
    alias: {
      '~bootstrap': path.resolve(import.meta.dirname, 'node_modules/bootstrap'),
      '~bootswatch': path.resolve(import.meta.dirname, 'node_modules/bootswatch'),
    }
  },
  build: {
    // generate manifest.json at the root of outDir (vite 5+ defaults to .vite/manifest.json,
    // which the Go embed would skip and the server reads from dist/manifest.json)
    manifest: 'manifest.json',
  },
})
