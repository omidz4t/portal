<script lang="ts">
	import './layout.css';
	import SiteHeader from '$lib/SiteHeader.svelte';
	import SiteFooter from '$lib/SiteFooter.svelte';
	import { page } from '$app/state';

	let { children } = $props();

	const onDocs = $derived(page.url.pathname.startsWith('/docs'));
</script>

<svelte:head>
	<title>TGPORTAL — Telegram ↔ Delta Chat media bridge</title>
	<meta
		name="description"
		content="Pair once, then stickers, GIFs, Lottie, and short video move both ways between Telegram and Delta Chat. Self-host if you can: you trust whoever runs the bot with bridged data."
	/>
	<meta property="og:title" content="TGPORTAL" />
	<meta
		property="og:description"
		content="Bidirectional Telegram ↔ Delta Chat bridge. Prefer self-hosting — the operator sees bridged traffic."
	/>
	<meta property="og:image" content="/poster.jpg" />
	<link rel="preconnect" href="https://fonts.googleapis.com" />
	<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin="anonymous" />
	<link
		href="https://fonts.googleapis.com/css2?family=IBM+Plex+Mono:wght@400;500&family=IBM+Plex+Sans:wght@400;500;600;700&display=swap"
		rel="stylesheet"
	/>
</svelte:head>

<div class="flex min-h-dvh flex-col">
	<a class="skip-link" href="#main-content">Skip to main content</a>
	<SiteHeader current={onDocs ? 'docs' : ''} />
	<main id="main-content" class="flex-1" tabindex="-1">{@render children()}</main>
	<SiteFooter />
</div>
