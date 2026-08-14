<script lang="ts">
	import DocsConfig from '$lib/docs/DocsConfig.svelte';
	import { docsCopies } from '$lib/content';
</script>

<DocsConfig docs={docsCopies.en} locale="en" />
