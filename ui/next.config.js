const {
	PHASE_DEVELOPMENT_SERVER,
	PHASE_PRODUCTION_BUILD,
} = require('next/constants')

module.exports = (phase) => {
	// when started in development mode `next dev` or `npm run dev` regardless of the value of STAGING environmental variable
	const isDev = phase === PHASE_DEVELOPMENT_SERVER
	// when `next build` or `npm run build` is used
	const isProd = phase === PHASE_PRODUCTION_BUILD

	console.log(`isDev:${isDev}  isProd:${isProd}`)

	const env = {
		BASE_URL: (() => {
			if (isDev) return 'http://localhost:8090'
			if (isProd) return 'http://localhost:8090'
		})()
	}

	// next.config.js object
	return {
		basePath: '/ui',
		reactStrictMode: true,
		env,
	}
}
