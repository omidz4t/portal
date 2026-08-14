<script lang="ts">
	import ShowWalkthrough from '$lib/ShowWalkthrough.svelte';
	import { tutorialCopies } from '$lib/content';
</script>

<ShowWalkthrough tutorial={tutorialCopies.en} locale="en" />
