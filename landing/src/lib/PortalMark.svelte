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
	let leftGate: HTMLElement | undefined = $state();
	let rightGate: HTMLElement | undefined = $state();

	let tgPath = $state({ enter: '0px', exitStart: '0px', exitEnd: '0px' });
	let dcPath = $state({ enter: '0px', exitStart: '0px', exitEnd: '0px' });

	function deltaX(from: HTMLElement, to: HTMLElement) {
		return `${to.offsetLeft + to.offsetWidth / 2 - (from.offsetLeft + from.offsetWidth / 2)}px`;
	}

	function measure() {
		if (!leftSlot || !rightSlot || !leftGate || !rightGate) return;
		tgPath = {
			enter: deltaX(leftSlot, leftGate),
			exitStart: deltaX(leftSlot, rightGate),
			exitEnd: deltaX(leftSlot, rightSlot)
		};
		dcPath = {
			enter: deltaX(rightSlot, rightGate),
			exitStart: deltaX(rightSlot, leftGate),
			exitEnd: deltaX(rightSlot, leftSlot)
		};
	}

	$effect(() => {
		if (!root) return;
		measure();
		const observer = new ResizeObserver(measure);
		observer.observe(root);
		return () => observer.disconnect();
	});
</script>

<p class="portal-mark" bind:this={root} dir="ltr">
	<span
		class="portal-mark-side portal-mark-tg"
		bind:this={leftSlot}
		style="--enter: {tgPath.enter}; --exit-start: {tgPath.exitStart}; --exit-end: {tgPath.exitEnd}"
		>{telegram}</span
	>
	<span class="portal-gate" bind:this={leftGate}>
		<KoboyoIcon name="mirror-portal" class={iconClass} />
	</span>
	<span class="portal-mark-name">{name}</span>
	<span class="portal-gate portal-mark-flip" bind:this={rightGate}>
		<KoboyoIcon name="mirror-portal" class={iconClass} />
	</span>
	<span
		class="portal-mark-side portal-mark-dc"
		bind:this={rightSlot}
		style="--enter: {dcPath.enter}; --exit-start: {dcPath.exitStart}; --exit-end: {dcPath.exitEnd}"
		>{delta}</span
	>
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
		animation: portal-swap 6.8s linear infinite;
	}

	.portal-gate {
		position: relative;
		z-index: 2;
		display: inline-flex;
		flex-shrink: 0;
		animation: gate-pulse 6.8s ease-in-out infinite;
	}

	.portal-mark-flip {
		transform: scaleX(-1);
		animation-name: gate-pulse-flip;
	}

	@keyframes portal-swap {
		0%,
		12% {
			transform: translateX(0);
			opacity: 1;
		}
		28% {
			transform: translateX(var(--enter));
			opacity: 0;
		}
		28.01% {
			transform: translateX(var(--exit-start));
			opacity: 0;
		}
		44%,
		56% {
			transform: translateX(var(--exit-end));
			opacity: 1;
		}
		72% {
			transform: translateX(var(--exit-start));
			opacity: 0;
		}
		72.01% {
			transform: translateX(var(--enter));
			opacity: 0;
		}
		88%,
		100% {
			transform: translateX(0);
			opacity: 1;
		}
	}

	@keyframes gate-pulse {
		0%,
		22%,
		34%,
		66%,
		78%,
		100% {
			transform: scale(1);
		}
		28%,
		72% {
			transform: scale(1.12);
		}
	}

	@keyframes gate-pulse-flip {
		0%,
		22%,
		34%,
		66%,
		78%,
		100% {
			transform: scaleX(-1) scale(1);
		}
		28%,
		72% {
			transform: scaleX(-1) scale(1.12);
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.portal-mark-side,
		.portal-gate {
			animation: none;
		}
	}
</style>
