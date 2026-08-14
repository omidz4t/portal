<script lang="ts">
	import Walkthrough from '$lib/Walkthrough.svelte';
	import type { Locale, TutorialCopy } from '$lib/content';
	import { basePath } from '$lib/content';

	let { tutorial, locale }: { tutorial: TutorialCopy; locale: Locale } = $props();
</script>

<div class="try-page">
	<div class="try-page-inner">
		<div class="try-page-bar">
			<a href="{basePath(locale)}#try" class="doodle-btn try-page-back">{tutorial.homeBack}</a>
			<h1 class="try-headline try-page-title">{tutorial.title}</h1>
		</div>
		<Walkthrough {tutorial} hideTitle />
	</div>
</div>
