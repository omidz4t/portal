import tailwindcss from '@tailwindcss/vite';
import adapter from '@sveltejs/adapter-static';
import { sveltekit } from '@sveltejs/kit/vite';
import { mdsvex } from 'mdsvex';
import { defineConfig } from 'vite';

// GitHub Pages project site is https://omidz4t.github.io/portal — set in CI.
const env = (globalThis as { process?: { env?: Record<string, string | undefined> } }).process?.env;
const base = (env?.BASE_PATH ?? '') as '' | `/${string}`;

export default defineConfig({
	plugins: [
		tailwindcss(),
		sveltekit({
			compilerOptions: {
				// Force runes mode for the project, except for libraries. Can be removed in svelte 6.
				runes: ({ filename }) =>
					filename.split(/[/\\]/).includes('node_modules') ? undefined : true
			},
			adapter: adapter({
				fallback: '404.html',
				strict: true
			}),
			paths: { base, relative: false },
			extensions: ['.svelte', '.svx', '.md'],
			preprocess: [
				mdsvex({
					extensions: ['.svx', '.md']
				})
			]
		})
	]
});
