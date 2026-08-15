<script lang="ts">
	import DocLayout from '$lib/DocLayout.svelte';
	import type { DocsCopy, Locale } from '$lib/content';
	import { renderMarkdown, splitMarkdown } from '$lib/md';
	import fullConfig from './full-config.yml?raw';
	import enMd from './config.en.md?raw';
	import faMd from './config.fa.md?raw';

	let { docs, locale }: { docs: DocsCopy; locale: Locale } = $props();

	const parsed = $derived(splitMarkdown(locale === 'fa' ? faMd : enMd));
	const parts = $derived(parsed.body.split('<!--full-config-->'));
	const before = $derived(renderMarkdown(parts[0] ?? ''));
	const after = $derived(renderMarkdown(parts.slice(1).join('<!--full-config-->')));
</script>

<div class="mx-auto max-w-5xl px-5 py-16">
	<DocLayout title={parsed.title || docs.cfgTitle} lede={parsed.lede} {docs} {locale} wide>
		{@html before}
		<pre class="config-yml" dir="ltr"><code>{fullConfig.trimEnd()}</code></pre>
		{@html after}
	</DocLayout>
</div>
