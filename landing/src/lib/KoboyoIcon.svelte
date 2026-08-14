<script lang="ts">
	import { asset } from '$lib/links';

	let {
		name,
		label = '',
		class: className = 'h-12 w-12'
	}: {
		name: string;
		label?: string;
		class?: string;
	} = $props();

	const icon = $derived(asset(`/icons/${name}.svg`));
</script>

<span
	class="inline-block shrink-0 bg-current {className}"
	style="mask: url({icon}) center / contain no-repeat; -webkit-mask: url({icon}) center / contain no-repeat;"
	role={label ? 'img' : undefined}
	aria-label={label || undefined}
	aria-hidden={label ? undefined : true}
></span>
