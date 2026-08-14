<script lang="ts">
	import './layout.css';

	let { children } = $props();
</script>

<svelte:head>
	<title>TGPORTAL — Telegram ↔ Delta Chat media bridge</title>
	<meta
		name="description"
		content="Pair once, then stickers, GIFs, Lottie, and short video move both ways between Telegram and Delta Chat."
	/>
	<meta property="og:title" content="TGPORTAL" />
	<meta
		property="og:description"
		content="Bidirectional Telegram ↔ Delta Chat bridge for stickers, GIFs, and more."
	/>
	<meta property="og:image" content="/poster.jpg" />
	<link rel="preconnect" href="https://fonts.googleapis.com" />
	<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin="anonymous" />
	<link
		href="https://fonts.googleapis.com/css2?family=IBM+Plex+Mono:wght@400;500&family=IBM+Plex+Sans:wght@400;500;600;700&display=swap"
		rel="stylesheet"
	/>
</svelte:head>

<div class="min-h-screen">
	{@render children()}
</div>
