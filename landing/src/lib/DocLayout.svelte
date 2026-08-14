<script lang="ts">
	import type { Snippet } from 'svelte';

	let {
		title,
		lede,
		children
	}: {
		title: string;
		lede?: string;
		children: Snippet;
	} = $props();
</script>

<article class="prose max-w-none prose-invert prose-headings:scroll-mt-20">
	<p class="font-mono text-xs tracking-[0.2em] text-teal uppercase">Docs</p>
	<h1 class="mt-2 text-3xl font-semibold tracking-tight text-paper">{title}</h1>
	{#if lede}
		<p class="text-lg text-mist">{lede}</p>
	{/if}
	<div class="text-mist [&_a]:text-teal [&_code]:text-paper [&_strong]:text-paper">
		{@render children()}
	</div>
</article>
