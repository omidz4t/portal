<script lang="ts">
	import './layout.css';
	import SiteHeader from '$lib/SiteHeader.svelte';
	import SiteFooter from '$lib/SiteFooter.svelte';
	import { page } from '$app/state';
	import { copies, localeFromPath } from '$lib/content';

	let { children } = $props();

	const locale = $derived(localeFromPath(page.url.pathname));
	const copy = $derived(copies[locale]);
	const onDocs = $derived(page.url.pathname.includes('/docs'));
	const onShow = $derived(page.url.pathname.includes('/show'));
	const headerCurrent = $derived(onDocs ? 'docs' : onShow ? 'try' : '');

	$effect(() => {
		document.documentElement.lang = locale === 'fa' ? 'fa' : 'en';
		document.documentElement.dir = locale === 'fa' ? 'rtl' : 'ltr';
	});
</script>

<svelte:head>
	<title>{copy.metaTitle}</title>
	<meta name="description" content={copy.metaDescription} />
	<meta property="og:title" content={copy.brandName} />
	<meta property="og:description" content={copy.metaDescription} />
	<meta property="og:image" content="/poster.jpg" />
	<link rel="alternate" hreflang="en" href="/" />
	<link rel="alternate" hreflang="fa" href="/fa/" />
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
	<SiteFooter {copy} {locale} />
</div>
