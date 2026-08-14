import { base } from '$app/paths';

export const botUrl = 'https://t.me/tgdeltabridgebot';
export const repoUrl = 'https://github.com/omidz4t/portal';
export const repoDocs = `${repoUrl}/tree/main/docs`;

/** Public origin for canonical / Open Graph URLs. */
export const siteOrigin = 'https://omidz4t.github.io/portal';

/** Static file under `paths.base` (GitHub Pages `/portal`). */
export function asset(path: string): string {
	const p = path.startsWith('/') ? path : `/${path}`;
	return `${base}${p}`;
}
