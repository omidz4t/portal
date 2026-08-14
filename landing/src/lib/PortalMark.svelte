<script lang="ts">
	import KoboyoIcon from '$lib/KoboyoIcon.svelte';

	let {
		telegram,
		name,
		delta,
		iconClass = 'h-10 w-10'
	}: {
		telegram: string;
		name: string;
		delta: string;
		iconClass?: string;
	} = $props();

	let root: HTMLElement | undefined = $state();
	let leftSlot: HTMLElement | undefined = $state();
	let rightSlot: HTMLElement | undefined = $state();
	let span = $state(0);

	function measure() {
		if (!leftSlot || !rightSlot) return;
		span = Math.max(0, rightSlot.offsetLeft - leftSlot.offsetLeft);
	}

	$effect(() => {
		if (!root) return;
		measure();
		const observer = new ResizeObserver(measure);
		observer.observe(root);
		return () => observer.disconnect();
	});
</script>

<p
	class="portal-mark"
	bind:this={root}
	style="--portal-span: {span}px"
	dir="ltr"
>
	<span class="portal-mark-side portal-mark-tg" bind:this={leftSlot}>{telegram}</span>
	<span class="portal-gate">
		<KoboyoIcon name="mirror-portal" class={iconClass} />
	</span>
	<span class="portal-mark-name">{name}</span>
	<span class="portal-gate portal-mark-flip">
		<KoboyoIcon name="mirror-portal" class={iconClass} />
	</span>
	<span class="portal-mark-side portal-mark-dc" bind:this={rightSlot}>{delta}</span>
</p>

<style>
	.portal-mark {
		display: inline-flex;
		flex-wrap: nowrap;
		align-items: center;
		gap: 0.45rem;
	}

	.portal-mark-name {
		position: relative;
		z-index: 2;
		font-size: clamp(0.85rem, 1.8vw, 1.05rem);
		font-weight: 700;
		line-height: 1;
	}

	.portal-mark-side {
		position: relative;
		z-index: 1;
		font-size: clamp(0.7rem, 1.5vw, 0.85rem);
		font-weight: 700;
		line-height: 1;
		letter-spacing: 0.02em;
		white-space: nowrap;
	}

	.portal-gate {
		position: relative;
		z-index: 2;
		display: inline-flex;
		flex-shrink: 0;
		animation: gate-pulse 6.4s ease-in-out infinite;
	}

	.portal-mark-flip {
		transform: scaleX(-1);
		animation-name: gate-pulse-flip;
	}

	.portal-mark-tg {
		animation: telegram-teleport 6.4s ease-in-out infinite;
	}

	.portal-mark-dc {
		animation: delta-walk-left 6.4s ease-in-out infinite;
	}

	@keyframes telegram-teleport {
		0%,
		16% {
			transform: translateX(0) scale(1);
			opacity: 1;
			filter: blur(0);
		}
		24% {
			transform: translateX(0.55rem) scale(0.25);
			opacity: 0;
			filter: blur(3px);
		}
		25%,
		32% {
			transform: translateX(var(--portal-span)) scale(0.15);
			opacity: 0;
			filter: blur(4px);
		}
		40%,
		66% {
			transform: translateX(var(--portal-span)) scale(1);
			opacity: 1;
			filter: blur(0);
		}
		74% {
			transform: translateX(calc(var(--portal-span) - 0.55rem)) scale(0.25);
			opacity: 0;
			filter: blur(3px);
		}
		75%,
		82% {
			transform: translateX(0) scale(0.15);
			opacity: 0;
			filter: blur(4px);
		}
		90%,
		100% {
			transform: translateX(0) scale(1);
			opacity: 1;
			filter: blur(0);
		}
	}

	@keyframes delta-walk-left {
		0%,
		14% {
			transform: translateX(0);
			opacity: 1;
		}
		40%,
		66% {
			transform: translateX(calc(-1 * var(--portal-span)));
			opacity: 1;
		}
		92%,
		100% {
			transform: translateX(0);
			opacity: 1;
		}
	}

	@keyframes gate-pulse {
		0%,
		20%,
		46%,
		70%,
		100% {
			transform: scale(1);
		}
		24%,
		74% {
			transform: scale(1.14);
		}
	}

	@keyframes gate-pulse-flip {
		0%,
		20%,
		46%,
		70%,
		100% {
			transform: scaleX(-1) scale(1);
		}
		24%,
		74% {
			transform: scaleX(-1) scale(1.14);
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.portal-mark-tg,
		.portal-mark-dc,
		.portal-gate {
			animation: none;
		}
	}
</style>
