<script lang="ts">
	import './layout.css';
	import SiteHeader from '$lib/SiteHeader.svelte';
	import SiteFooter from '$lib/SiteFooter.svelte';
	import { page } from '$app/state';
	import { copies, localeFromPath } from '$lib/content';
	import { seoFor } from '$lib/seo';

	let { children } = $props();

	const locale = $derived(localeFromPath(page.url.pathname));
	const copy = $derived(copies[locale]);
	const onDocs = $derived(page.url.pathname.includes('/docs'));
	const onShow = $derived(page.url.pathname.includes('/show'));
	const isHome = $derived(
		page.url.pathname === '/' || page.url.pathname === '/fa' || page.url.pathname === '/fa/'
	);
	const headerCurrent = $derived(onDocs ? 'docs' : onShow ? 'try' : '');
	const seo = $derived(seoFor(page.url.pathname));

	$effect(() => {
		document.documentElement.lang = locale === 'fa' ? 'fa' : 'en';
		document.documentElement.dir = locale === 'fa' ? 'rtl' : 'ltr';
	});
</script>

<svelte:head>
	<title>{seo.title}</title>
	<meta name="description" content={seo.description} />
	<link rel="canonical" href={seo.canonical} />
	<link rel="alternate" hreflang="en" href={seo.alternateEn} />
	<link rel="alternate" hreflang="fa" href={seo.alternateFa} />
	<link rel="alternate" hreflang="x-default" href={seo.alternateEn} />
	<meta property="og:type" content="website" />
	<meta property="og:site_name" content={seo.siteName} />
	<meta property="og:locale" content={seo.ogLocale} />
	<meta property="og:locale:alternate" content={seo.ogLocaleAlt} />
	<meta property="og:url" content={seo.canonical} />
	<meta property="og:title" content={seo.title} />
	<meta property="og:description" content={seo.description} />
	<meta property="og:image" content={seo.image} />
	<meta property="og:image:alt" content={seo.imageAlt} />
	<meta name="twitter:card" content="summary_large_image" />
	<meta name="twitter:title" content={seo.title} />
	<meta name="twitter:description" content={seo.description} />
	<meta name="twitter:image" content={seo.image} />
	<meta name="twitter:image:alt" content={seo.imageAlt} />
	<link rel="preconnect" href="https://fonts.googleapis.com" />
	<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin="anonymous" />
	<link
		href="https://fonts.googleapis.com/css2?family=IBM+Plex+Mono:wght@400;500&family=Patrick+Hand&display=swap"
		rel="stylesheet"
	/>
</svelte:head>

<div
	class="flex min-h-dvh flex-col"
	dir={locale === 'fa' ? 'rtl' : 'ltr'}
	lang={locale === 'fa' ? 'fa' : 'en'}
>
	<a class="skip-link" href="#main-content">{copy.skip}</a>
	<SiteHeader {copy} {locale} pathname={page.url.pathname} current={headerCurrent} />
	<main id="main-content" class="flex-1" tabindex="-1">{@render children()}</main>
	{#if !isHome}
		<SiteFooter {copy} {locale} />
	{/if}
</div>
