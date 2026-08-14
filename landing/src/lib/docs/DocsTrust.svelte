<script lang="ts">
	import DocLayout from '$lib/DocLayout.svelte';
	import { botUrl } from '$lib/links';
	import type { DocsCopy, Locale } from '$lib/content';
	import { copies, docsPath } from '$lib/content';

	let { docs, locale }: { docs: DocsCopy; locale: Locale } = $props();
</script>

<svelte:head>
	<title>{docs.trustTitle} — {copies[locale].brandName}</title>
</svelte:head>

<div class="mx-auto max-w-3xl px-5 py-16">
	<DocLayout title={docs.trustTitle} lede={docs.trustLede} {docs} {locale}>
		<p>{docs.trustP1}</p>
		<p>{docs.trustIncludes}</p>
		<ul>
			{#each docs.trustItems as item}
				<li>{item}</li>
			{/each}
		</ul>
		<p>{docs.trustP2}</p>
		<p>
			{docs.trustP3}
			<a href={botUrl}>@tgdeltabridgebot</a>
		</p>
		<p>
			<strong>{docs.trustP4Lead}</strong>
			{docs.trustP4}
			<a href={docsPath(locale, 'self-host')}>{docs.trustLinkSelf}</a>
		</p>
	</DocLayout>
</div>
