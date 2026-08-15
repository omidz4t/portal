import { buildSitemapXml } from '$lib/sitemap';

export const prerender = true;

export function GET() {
	return new Response(buildSitemapXml(), {
		headers: {
			'Content-Type': 'application/xml; charset=utf-8',
			'Cache-Control': 'public, max-age=3600'
		}
	});
}
