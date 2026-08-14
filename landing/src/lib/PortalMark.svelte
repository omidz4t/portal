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
	let travel = $state(0);
	let ready = $state(false);

	function measure() {
		if (!leftSlot || !rightSlot) return;
		const next = Math.round(rightSlot.offsetLeft - leftSlot.offsetLeft);
		if (next <= 0) return;
		travel = next;
		ready = true;
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
	class:is-ready={ready}
	bind:this={root}
	style="--travel: {travel}px"
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

	.portal-mark-name,
	.portal-gate {
		position: relative;
		z-index: 2;
	}

	.portal-mark-name {
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
		will-change: transform;
	}

	.portal-gate {
		display: inline-flex;
		flex-shrink: 0;
	}

	.portal-mark-flip {
		transform: scaleX(-1);
	}

	.portal-mark.is-ready .portal-mark-tg {
		animation: slide-right 5.6s ease-in-out infinite;
	}

	.portal-mark.is-ready .portal-mark-dc {
		animation: slide-left 5.6s ease-in-out infinite;
	}

	@keyframes slide-right {
		0%,
		16% {
			transform: translateX(0);
		}
		42%,
		58% {
			transform: translateX(var(--travel));
		}
		84%,
		100% {
			transform: translateX(0);
		}
	}

	@keyframes slide-left {
		0%,
		16% {
			transform: translateX(0);
		}
		42%,
		58% {
			transform: translateX(calc(-1 * var(--travel)));
		}
		84%,
		100% {
			transform: translateX(0);
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.portal-mark.is-ready .portal-mark-tg,
		.portal-mark.is-ready .portal-mark-dc {
			animation: none;
		}
	}
</style>
