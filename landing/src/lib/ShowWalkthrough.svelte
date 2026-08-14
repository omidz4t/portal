<script lang="ts">
	import Walkthrough from '$lib/Walkthrough.svelte';
	import type { Locale, TutorialCopy } from '$lib/content';
	import { basePath } from '$lib/content';

	let { tutorial, locale }: { tutorial: TutorialCopy; locale: Locale } = $props();
</script>

<svelte:head>
	<title>{tutorial.title.replace('\n', ' ')}</title>
</svelte:head>

<div class="try-page">
	<div class="try-page-inner">
		<a href="{basePath(locale)}#try" class="doodle-btn">{tutorial.homeBack}</a>
		<Walkthrough {tutorial} />
	</div>
</div>
