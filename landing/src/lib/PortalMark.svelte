<script lang="ts">
	import KoboyoIcon from '$lib/KoboyoIcon.svelte';

	let {
		telegram,
		name,
		delta,
		iconClass = 'h-12 w-12'
	}: {
		telegram: string;
		name: string;
		delta: string;
		iconClass?: string;
	} = $props();
</script>

<p class="portal-mark">
	<span class="portal-mark-side">{telegram}</span>
	<KoboyoIcon name="mirror-portal" class={iconClass} />
	<span class="portal-mark-name">{name}</span>
	<KoboyoIcon name="mirror-portal" class="portal-mark-flip {iconClass}" />
	<span class="portal-mark-side">{delta}</span>
</p>

<style>
	.portal-mark {
		display: inline-flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 0.85rem;
	}

	.portal-mark-name {
		font-size: clamp(1.6rem, 4vw, 2.4rem);
		font-weight: 700;
		line-height: 1;
	}

	.portal-mark-side {
		font-size: clamp(1rem, 2.4vw, 1.35rem);
		font-weight: 700;
		line-height: 1;
		letter-spacing: 0.02em;
	}

	.portal-mark :global(.portal-mark-flip) {
		transform: scaleX(-1);
	}
</style>
