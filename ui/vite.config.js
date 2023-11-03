const path = require('path')
import { defineConfig } from 'vite'
export default defineConfig({
	// root: path.resolve(__dirname, 'src'),

	// must be set to the UI root
	base: "",
	appType: "custom",
	resolve: {
		alias: {
			'~bootstrap': path.resolve(__dirname, 'node_modules/bootstrap'),
			'~bootswatch': path.resolve(__dirname, 'node_modules/bootswatch'),
		}
	},
	build: {
		// generate manifest.json in outDir
		manifest: true,
	},
})
