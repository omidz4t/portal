<script lang="ts">
	import { botUrl, repoUrl } from '$lib/links';
	import type { Copy, Locale } from '$lib/content';
	import { basePath, docsPath } from '$lib/content';

	let { copy, locale }: { copy: Copy; locale: Locale } = $props();
	const base = $derived(basePath(locale));
</script>

<footer class="doodle-bar">
	<div
		class="mx-auto flex max-w-6xl flex-wrap items-center justify-between gap-4 px-5 py-8 text-sm text-mist"
	>
		<p>
			{copy.footerCredit}
			<a class="underline-offset-2 hover:underline" href="https://koboyo.com/icons" rel="noreferrer"
				>{copy.footerIcons}</a
			>
		</p>
		<nav class="flex flex-wrap gap-1" aria-label="Footer">
			<a class="inline-flex min-h-11 items-center px-2 hover:text-ink" href={docsPath(locale)}
				>{copy.navDocs}</a
			>
			<a
				class="inline-flex min-h-11 items-center px-2 hover:text-ink"
				href={docsPath(locale, 'trust')}>{copy.navTrust}</a
			>
			<a class="inline-flex min-h-11 items-center px-2 hover:text-ink" href={`${base}#persona`}
				>{copy.navPersonaShort}</a
			>
			<a
				class="inline-flex min-h-11 items-center px-2 hover:text-ink"
				href={repoUrl}
				rel="noreferrer">{copy.navSource}</a
			>
			<a
				class="inline-flex min-h-11 items-center px-2 hover:text-ink"
				href={botUrl}
				rel="noreferrer">{copy.navTelegram}</a
			>
		</nav>
	</div>
</footer>
