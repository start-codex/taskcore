import adapter from '@sveltejs/adapter-static';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	kit: {
		adapter: adapter({
			pages: '../ui/dist',
			assets: '../ui/dist',
			fallback: 'index.html'  // SPA mode — todas las rutas caen en index.html
		})
	}
};

export default config;
