<script lang="ts">
	import type { Snippet } from 'svelte';
	import type { DocsCopy, Locale } from '$lib/content';
	import { basePath, docsPath } from '$lib/content';

	let {
		title,
		lede,
		docs,
		locale,
		children
	}: {
		title: string;
		lede?: string;
		docs: DocsCopy;
		locale: Locale;
		children: Snippet;
	} = $props();
</script>

<article class="prose max-w-none prose-neutral prose-headings:scroll-mt-24">
	<nav aria-label="Breadcrumb" class="not-prose text-sm text-mist">
		<ol class="flex flex-wrap items-center gap-2">
			<li>
				<a
					class="inline-flex min-h-11 items-center underline-offset-2 hover:underline"
					href={basePath(locale)}>{docs.crumbHome}</a
				>
			</li>
			<li aria-hidden="true">/</li>
			<li>
				<a
					class="inline-flex min-h-11 items-center underline-offset-2 hover:underline"
					href={docsPath(locale)}>{docs.crumbDocs}</a
				>
			</li>
			<li aria-hidden="true">/</li>
			<li class="text-ink" aria-current="page">{title}</li>
		</ol>
	</nav>
	<p class="text-sm text-mist">Docs</p>
	<h1 class="mt-2 text-3xl font-semibold tracking-tight text-balance text-ink">{title}</h1>
	{#if lede}
		<p class="docs-copy max-w-prose text-lg text-mist">{lede}</p>
	{/if}
	<div
		class="docs-copy max-w-prose text-mist [&_a]:text-sky [&_a]:underline-offset-2 [&_a]:hover:underline [&_code]:text-ink [&_pre]:overflow-x-auto [&_strong]:text-ink"
	>
		{@render children()}
	</div>
</article>
