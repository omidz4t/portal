<script lang="ts">
	import DocLayout from '$lib/DocLayout.svelte';
	import { botUrl, repoUrl } from '$lib/links';
	import type { DocsCopy, Locale } from '$lib/content';
	import { docsPath } from '$lib/content';

	let { docs, locale }: { docs: DocsCopy; locale: Locale } = $props();
</script>

<svelte:head>
	<title>{docs.pairingTitle} — TGPORTAL</title>
</svelte:head>

<div class="mx-auto max-w-3xl px-5 py-16">
	<DocLayout title={docs.pairingTitle} lede={docs.pairingLede} {docs} {locale}>
		<p>
			{docs.pairingP1}
			<a href={docsPath(locale, 'persona')}>{docs.pages[1].title}</a>
		</p>
		<p>
			{docs.pairingP2}
			<a href={docsPath(locale, 'trust')}>{docs.pages[0].title}</a>
			·
			<a href={docsPath(locale, 'self-host')}>{docs.trustLinkSelf}</a>
		</p>
		<h2>{docs.pairingTgFirst}</h2>
		<ol>
			{#each docs.pairingTgItems as item}
				<li>{item}</li>
			{/each}
		</ol>
		<p>
			<a href={botUrl}>@tgdeltabridgebot</a>
		</p>
		<h2>{docs.pairingDcFirst}</h2>
		<ol>
			{#each docs.pairingDcItems as item}
				<li>{item}</li>
			{/each}
		</ol>
		<h2>{docs.pairingCmds}</h2>
		<ul>
			{#each docs.pairingCmdItems as item}
				<li>{item}</li>
			{/each}
		</ul>
		<p>
			{docs.pairingRepo}
			<a href="{repoUrl}/blob/main/docs/pairing.md">docs/pairing.md</a>
		</p>
	</DocLayout>
</div>
