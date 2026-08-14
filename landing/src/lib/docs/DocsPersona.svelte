<script lang="ts">
	import DocLayout from '$lib/DocLayout.svelte';
	import { repoUrl } from '$lib/links';
	import type { DocsCopy, Locale } from '$lib/content';
	import { docsPath } from '$lib/content';

	let { docs, locale }: { docs: DocsCopy; locale: Locale } = $props();
</script>

<div class="mx-auto max-w-3xl px-5 py-16">
	<DocLayout title={docs.personaTitle} lede={docs.personaLede} {docs} {locale}>
		<h2>{docs.personaWhy}</h2>
		<p>{docs.personaWhyP1}</p>
		<p>{docs.personaWhyP2}</p>
		<h2>{docs.personaGhosts}</h2>
		<p>{docs.personaGhostsP1}</p>
		<p>{docs.personaGhostsP2}</p>
		<p>{docs.personaGhostsP3}</p>
		<h2>{docs.personaSetup}</h2>
		<ol>
			{#each docs.personaSetupItems as item}
				<li>{item}</li>
			{/each}
		</ol>
		<p>{docs.personaLimits}</p>
		<h2>{docs.personaGroups}</h2>
		<p>{docs.personaGroupsLead}</p>
		<ol>
			{#each docs.personaGroupsItems as item}
				<li>{item}</li>
			{/each}
		</ol>
		<h2>{docs.personaTrust}</h2>
		<p>
			{docs.personaTrustP1}
			<a href={docsPath(locale, 'self-host')}>{docs.trustLinkSelf}</a>
			·
			<a href={docsPath(locale, 'trust')}>{docs.pages[0].title}</a>
		</p>
		<p>
			{docs.personaRepo}
			<a href="{repoUrl}/blob/main/docs/persona.md">docs/persona.md</a>
			·
			<a href="{repoUrl}/blob/main/docs/persona-design.md">persona-design.md</a>
		</p>
	</DocLayout>
</div>
