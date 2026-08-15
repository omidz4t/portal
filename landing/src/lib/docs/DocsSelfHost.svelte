<script lang="ts">
	import DocLayout from '$lib/DocLayout.svelte';
	import type { DocsCopy, Locale } from '$lib/content';
	import { renderMarkdown, splitMarkdown } from '$lib/md';
	import OsPicker from './OsPicker.svelte';
	import enMd from './self-host.en.md?raw';
	import faMd from './self-host.fa.md?raw';

	let { docs, locale }: { docs: DocsCopy; locale: Locale } = $props();

	const parsed = $derived(splitMarkdown(locale === 'fa' ? faMd : enMd));
	const parts = $derived(parsed.body.split('<!--os-picker-->'));
	const before = $derived(renderMarkdown(parts[0] ?? ''));
	const after = $derived(renderMarkdown(parts.slice(1).join('<!--os-picker-->')));
</script>

<div class="mx-auto max-w-3xl px-5 py-16">
	<DocLayout title={parsed.title || docs.selfTitle} lede={parsed.lede} {docs} {locale}>
		{@html before}
		<OsPicker {locale} />
		{@html after}
	</DocLayout>
</div>
