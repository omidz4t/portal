<script lang="ts">
	import Walkthrough from '$lib/Walkthrough.svelte';
	import type { Locale, TutorialCopy } from '$lib/content';
	import { basePath } from '$lib/content';

	let { tutorial, locale }: { tutorial: TutorialCopy; locale: Locale } = $props();
</script>

<div class="try-page">
	<div class="try-page-inner">
		<Walkthrough {tutorial} homeHref="{basePath(locale)}#try" />
	</div>
</div>
