import { siteOrigin } from '$lib/links';

/** Public pathnames (trailing slash, no origin, no `/portal` prefix — origin already includes it). */
const pages = [
	'/',
	'/fa/',
	'/docs/',
	'/docs/trust/',
	'/docs/persona/',
	'/docs/pairing/',
	'/docs/self-host/',
	'/docs/config/',
	'/fa/docs/',
	'/fa/docs/trust/',
	'/fa/docs/persona/',
	'/fa/docs/pairing/',
	'/fa/docs/self-host/',
	'/fa/docs/config/',
	'/versions/',
	'/fa/versions/',
	'/show/',
	'/fa/show/'
] as const;

function twin(path: string): string {
	if (path === '/') return '/fa/';
	if (path === '/fa/') return '/';
	if (path.startsWith('/fa/')) return path.slice(3);
	return `/fa${path}`;
}

function escapeXml(value: string): string {
	return value
		.replaceAll('&', '&amp;')
		.replaceAll('<', '&lt;')
		.replaceAll('>', '&gt;')
		.replaceAll('"', '&quot;')
		.replaceAll("'", '&apos;');
}

export function getSitemapPathnames(): string[] {
	return [...pages];
}

export function buildSitemapXml(): string {
	const urls = pages
		.map((path) => {
			const loc = `${siteOrigin}${path}`;
			const en = `${siteOrigin}${path.startsWith('/fa') ? twin(path) : path}`;
			const fa = `${siteOrigin}${path.startsWith('/fa') ? path : twin(path)}`;
			return [
				'  <url>',
				`    <loc>${escapeXml(loc)}</loc>`,
				`    <xhtml:link rel="alternate" hreflang="en" href="${escapeXml(en)}" />`,
				`    <xhtml:link rel="alternate" hreflang="fa" href="${escapeXml(fa)}" />`,
				`    <xhtml:link rel="alternate" hreflang="x-default" href="${escapeXml(en)}" />`,
				'  </url>'
			].join('\n');
		})
		.join('\n');

	return `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9"
        xmlns:xhtml="http://www.w3.org/1999/xhtml">
${urls}
</urlset>
`;
}
