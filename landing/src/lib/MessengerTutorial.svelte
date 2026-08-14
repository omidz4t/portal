<script lang="ts">
	import KoboyoIcon from '$lib/KoboyoIcon.svelte';
	import SiteFooter from '$lib/SiteFooter.svelte';
	import type { Copy, Locale, TutorialCopy } from '$lib/content';
	import { showPath } from '$lib/content';

	let {
		tutorial,
		locale = 'en',
		copy
	}: { tutorial: TutorialCopy; locale?: Locale; copy: Copy } = $props();
</script>

<section id="try" class="try-band">
	<div class="try-cta mx-auto flex w-full max-w-6xl flex-col items-start justify-center px-5">
		<div class="try-cta-row">
			<span class="try-beam-wrap">
				<KoboyoIcon name="very-happy-beaming" class="try-beam" />
			</span>
			<h2 class="try-headline whitespace-pre-line">
				{tutorial.title}
			</h2>
		</div>
		<a href={showPath(locale)} class="try-open doodle-btn doodle-btn-ink">
			{tutorial.open}
		</a>
	</div>
	<SiteFooter {copy} {locale} />
</section>
