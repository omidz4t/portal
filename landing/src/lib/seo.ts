import {
	copies,
	docsCopies,
	tutorialCopies,
	localeFromPath,
	twinPath,
	type Locale
} from '$lib/content';
import { siteOrigin } from '$lib/links';

function withSlash(pathname: string): string {
	if (pathname === '') return '/';
	return pathname.endsWith('/') ? pathname : `${pathname}/`;
}

export function seoFor(pathname: string) {
	const path = withSlash(pathname);
	const locale: Locale = localeFromPath(path);
	const copy = copies[locale];
	const docs = docsCopies[locale];
	const tutorial = tutorialCopies[locale];
	const brand = copy.brandName;
	const docsRoot = locale === 'fa' ? '/fa/docs/' : '/docs/';
	const showPath = locale === 'fa' ? '/fa/show/' : '/show/';

	let title = copy.metaTitle;
	let description = copy.metaDescription;

	if (path === docsRoot) {
		title = `${docs.indexTitle} — ${brand}`;
		description = docs.indexLede;
	} else if (path === `${docsRoot}trust/`) {
		title = `${docs.trustTitle} — ${brand}`;
		description = docs.trustLede;
	} else if (path === `${docsRoot}persona/`) {
		title = `${docs.personaTitle} — ${brand}`;
		description = docs.personaLede;
	} else if (path === `${docsRoot}pairing/`) {
		title = `${docs.pairingTitle} — ${brand}`;
		description = docs.pairingLede;
	} else if (path === `${docsRoot}self-host/`) {
		title = `${docs.selfTitle} — ${brand}`;
		description = docs.selfLede;
	} else if (path === `${docsRoot}config/`) {
		title = `${docs.cfgTitle} — ${brand}`;
		description = docs.cfgLede;
	} else if (path === showPath) {
		title = `${tutorial.title.replace(/\n/g, ' ')} — ${brand}`;
		description = tutorial.lead.trim() || copy.heroLead;
	}

	return {
		title,
		description,
		locale,
		canonical: `${siteOrigin}${path}`,
		alternateEn: `${siteOrigin}${twinPath(path, 'en')}`,
		alternateFa: `${siteOrigin}${twinPath(path, 'fa')}`,
		ogLocale: locale === 'fa' ? 'fa_IR' : 'en_US',
		ogLocaleAlt: locale === 'fa' ? 'en_US' : 'fa_IR',
		image: `${siteOrigin}/poster.jpg`,
		imageAlt: copy.posterAlt,
		siteName: brand
	};
}
