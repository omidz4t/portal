<script lang="ts">
	import type { Copy, Locale } from '$lib/content';
	import { basePath } from '$lib/content';
	import versions from '$lib/versions-data';
	import { repoUrl } from '$lib/links';

	let { copy, locale }: { copy: Copy; locale: Locale } = $props();
	const home = $derived(basePath(locale));
</script>

<div class="mx-auto max-w-xl px-5 py-16">
	<p class="text-sm text-mist">
		<a class="underline-offset-2 hover:underline" href={home}>{copy.brandName}</a>
	</p>
	<h1 class="mt-3 text-3xl font-semibold tracking-tight">{copy.versionsTitle}</h1>
	<p class="mt-2 text-mist">{copy.versionsLead}</p>
	<ol class="version-list mt-8">
		{#each versions.releases as rel}
			<li>
				<a href="{repoUrl}/releases/tag/{rel.tag}" rel="noreferrer">{rel.tag}</a>
				{#if rel.date}
					<span>{rel.date}</span>
				{/if}
			</li>
		{/each}
	</ol>
</div>
