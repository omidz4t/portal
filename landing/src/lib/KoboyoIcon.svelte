<script lang="ts">
	let {
		name,
		label = '',
		class: className = 'h-12 w-12'
	}: {
		name: string;
		label?: string;
		class?: string;
	} = $props();
</script>

<span
	class="inline-block shrink-0 bg-current {className}"
	style="mask: url(/icons/{name}.svg) center / contain no-repeat; -webkit-mask: url(/icons/{name}.svg) center / contain no-repeat;"
	role={label ? 'img' : undefined}
	aria-label={label || undefined}
	aria-hidden={label ? undefined : true}
></span>
