<script lang="ts">
	import DocLayout from '$lib/DocLayout.svelte';
	import { repoDocs } from '$lib/links';
	import type { DocsCopy, Locale } from '$lib/content';
	import { docsPath } from '$lib/content';

	let { docs, locale }: { docs: DocsCopy; locale: Locale } = $props();
</script>

<svelte:head>
	<title>{docs.indexTitle} — TGPORTAL</title>
</svelte:head>

<div class="mx-auto max-w-3xl px-5 py-16">
	<DocLayout title={docs.indexTitle} lede={docs.indexLede} {docs} {locale}>
		<ul class="not-prose mt-8 grid gap-4">
			{#each docs.pages as p}
				<li>
					<a href={docsPath(locale, p.slug)} class="doodle-card block p-5">
						<h2 class="text-lg font-semibold text-ink">{p.title}</h2>
						<p class="mt-1 text-sm text-mist">{p.body}</p>
					</a>
				</li>
			{/each}
		</ul>
		<p class="mt-8">
			{docs.indexMore}
			<a href={repoDocs}>docs/</a>
		</p>
	</DocLayout>
</div>
