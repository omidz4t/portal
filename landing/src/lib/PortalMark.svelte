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
</script>

<p class="portal-mark" dir="ltr">
	<span class="sr-only">{telegram} {name} {delta}</span>
	<span class="portal-window portal-window-left" aria-hidden="true">
		<span class="portal-sizers">
			<span class="portal-sizer">{telegram}</span>
			<span class="portal-sizer">{delta}</span>
		</span>
		<span class="portal-rider portal-out-east">{telegram}</span>
		<span class="portal-rider portal-in-east">{delta}</span>
	</span>
	<span class="portal-gate">
		<KoboyoIcon name="mirror-portal" class={iconClass} />
	</span>
	<span class="portal-mark-name">{name}</span>
	<span class="portal-gate portal-mark-flip">
		<KoboyoIcon name="mirror-portal" class={iconClass} />
	</span>
	<span class="portal-window portal-window-right" aria-hidden="true">
		<span class="portal-sizers">
			<span class="portal-sizer">{telegram}</span>
			<span class="portal-sizer">{delta}</span>
		</span>
		<span class="portal-rider portal-out-west">{delta}</span>
		<span class="portal-rider portal-in-west">{telegram}</span>
	</span>
</p>

<style>
	.portal-mark {
		display: inline-flex;
		flex-wrap: nowrap;
		align-items: center;
		gap: 0.2rem;
	}

	.portal-mark-name {
		position: relative;
		z-index: 2;
		padding-inline: 0.25rem;
		font-size: clamp(0.85rem, 1.8vw, 1.05rem);
		font-weight: 700;
		line-height: 1;
	}

	.portal-window {
		position: relative;
		overflow: hidden;
		min-width: 7.25em;
		min-height: 1.15em;
	}

	.portal-window-left {
		margin-inline-end: -0.2rem;
	}

	.portal-window-right {
		margin-inline-start: -0.2rem;
	}

	.portal-sizers {
		display: grid;
	}

	.portal-sizer {
		grid-area: 1 / 1;
		visibility: hidden;
		padding-inline: 0.55rem;
		font-size: clamp(0.7rem, 1.5vw, 0.85rem);
		font-weight: 700;
		letter-spacing: 0.02em;
		line-height: 1.15;
		white-space: nowrap;
	}

	.portal-rider {
		position: absolute;
		inset: 0;
		padding-inline: 0.55rem;
		font-size: clamp(0.7rem, 1.5vw, 0.85rem);
		font-weight: 700;
		letter-spacing: 0.02em;
		line-height: 1.15;
		white-space: nowrap;
		will-change: transform;
	}

	.portal-window-left .portal-rider {
		text-align: end;
	}

	.portal-window-right .portal-rider {
		text-align: start;
	}

	.portal-gate {
		position: relative;
		z-index: 2;
		display: inline-flex;
		flex-shrink: 0;
	}

	.portal-mark-flip {
		transform: scaleX(-1);
	}

	.portal-out-east {
		animation: out-east 5.6s ease-in-out infinite;
	}

	.portal-in-east {
		animation: in-east 5.6s ease-in-out infinite;
	}

	.portal-out-west {
		animation: out-west 5.6s ease-in-out infinite;
	}

	.portal-in-west {
		animation: in-west 5.6s ease-in-out infinite;
	}

	@keyframes out-east {
		0%,
		16% {
			transform: translateX(0);
		}
		42%,
		58% {
			transform: translateX(115%);
		}
		84%,
		100% {
			transform: translateX(0);
		}
	}

	@keyframes in-east {
		0%,
		16% {
			transform: translateX(115%);
		}
		42%,
		58% {
			transform: translateX(0);
		}
		84%,
		100% {
			transform: translateX(115%);
		}
	}

	@keyframes out-west {
		0%,
		16% {
			transform: translateX(0);
		}
		42%,
		58% {
			transform: translateX(-115%);
		}
		84%,
		100% {
			transform: translateX(0);
		}
	}

	@keyframes in-west {
		0%,
		16% {
			transform: translateX(-115%);
		}
		42%,
		58% {
			transform: translateX(0);
		}
		84%,
		100% {
			transform: translateX(-115%);
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.portal-rider {
			animation: none;
		}

		.portal-in-east,
		.portal-in-west {
			display: none;
		}
	}
</style>
