<script lang="ts">
	import DocLayout from '$lib/DocLayout.svelte';
	import type { DocsCopy, Locale } from '$lib/content';
	import { renderMarkdown, splitMarkdown } from '$lib/md';
	import enMd from './persona.en.md?raw';
	import faMd from './persona.fa.md?raw';

	let { docs, locale }: { docs: DocsCopy; locale: Locale } = $props();

	const parsed = $derived(splitMarkdown(locale === 'fa' ? faMd : enMd));
	const html = $derived(renderMarkdown(parsed.body));
</script>

<div class="mx-auto max-w-3xl px-5 py-16">
	<DocLayout title={parsed.title || docs.personaTitle} lede={parsed.lede} {docs} {locale}>
		{@html html}
	</DocLayout>
</div>
