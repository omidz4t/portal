<script lang="ts">
	import KoboyoIcon from '$lib/KoboyoIcon.svelte';

	let {
		telegram,
		name,
		delta,
		iconClass = 'h-7 w-7'
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
		gap: 0.4rem;
	}

	.portal-mark-name {
		font-size: clamp(1.05rem, 2.4vw, 1.35rem);
		font-weight: 700;
		line-height: 1;
	}

	.portal-mark-side {
		font-size: clamp(0.75rem, 1.8vw, 0.95rem);
		font-weight: 700;
		line-height: 1;
		letter-spacing: 0.02em;
	}

	.portal-mark :global(.portal-mark-flip) {
		transform: scaleX(-1);
	}
</style>
