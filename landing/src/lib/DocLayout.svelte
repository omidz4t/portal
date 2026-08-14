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

<article class="prose max-w-none prose-invert prose-headings:scroll-mt-24">
	<nav aria-label="Breadcrumb" class="not-prose text-sm text-mist">
		<ol class="flex flex-wrap items-center gap-2">
			<li>
				<a class="inline-flex min-h-11 items-center underline-offset-2 hover:underline" href="/"
					>Home</a
				>
			</li>
			<li aria-hidden="true">/</li>
			<li>
				<a
					class="inline-flex min-h-11 items-center underline-offset-2 hover:underline"
					href="/docs/">Docs</a
				>
			</li>
			<li aria-hidden="true">/</li>
			<li class="text-paper" aria-current="page">{title}</li>
		</ol>
	</nav>
	<p class="font-mono text-xs tracking-[0.2em] text-teal uppercase">Docs</p>
	<h1 class="mt-2 text-3xl font-semibold tracking-tight text-balance text-paper">{title}</h1>
	{#if lede}
		<p class="max-w-prose text-lg text-mist">{lede}</p>
	{/if}
	<div
		class="max-w-prose text-mist [&_a]:text-teal [&_a]:underline-offset-2 [&_a]:hover:underline [&_code]:text-paper [&_pre]:overflow-x-auto [&_strong]:text-paper"
	>
		{@render children()}
	</div>
</article>
