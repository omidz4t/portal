<script lang="ts">
	import KoboyoIcon from '$lib/KoboyoIcon.svelte';
	import SiteFooter from '$lib/SiteFooter.svelte';
	import type { Copy, Locale, TutorialCopy } from '$lib/content';
	import { showPath } from '$lib/content';

	let {
		tutorial,
		locale = 'en',
		copy
	}: { tutorial: TutorialCopy; locale?: Locale; copy: Copy } = $props();

	type Sparkle = { id: number; x: number; y: number; size: number; rot: number; color: string };

	let band = $state<HTMLElement | undefined>();
	let sparkles = $state<Sparkle[]>([]);
	let lastX = 0;
	let lastY = 0;
	let nextId = 0;
	const gap = 20;
	const life = 680;
	const colors = ['#7c3aed', '#c026d3', '#db2777', '#2563eb', '#06b6d4', '#8b5cf6', '#e11d48'];

	function onPointerMove(e: PointerEvent) {
		if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) return;
		if (!band) return;
		if ((e.target as Element | null)?.closest('footer')) return;
		const rect = band.getBoundingClientRect();
		const x = e.clientX - rect.left;
		const y = e.clientY - rect.top;
		const dx = x - lastX;
		const dy = y - lastY;
		if (sparkles.length && dx * dx + dy * dy < gap * gap) return;
		lastX = x;
		lastY = y;
		const sparkle: Sparkle = {
			id: ++nextId,
			x,
			y,
			size: 12 + Math.random() * 18,
			rot: Math.random() * 360,
			color: colors[nextId % colors.length]
		};
		sparkles = [...sparkles.slice(-36), sparkle];
		window.setTimeout(() => {
			sparkles = sparkles.filter((item) => item.id !== sparkle.id);
		}, life);
	}
</script>

<section id="try" class="try-band">
	<div class="try-sparkle-stage" bind:this={band} onpointermove={onPointerMove}>
		<div class="try-sparkles" aria-hidden="true">
			{#each sparkles as sparkle (sparkle.id)}
				<span
					class="try-sparkle"
					style="left: {sparkle.x}px; top: {sparkle.y}px; width: {sparkle.size}px; height: {sparkle.size}px; --rot: {sparkle.rot}deg; color: {sparkle.color}"
				></span>
			{/each}
		</div>
		<div class="try-cta mx-auto flex w-full max-w-6xl flex-col items-start justify-center px-5">
			<div class="try-cta-row">
				<span class="try-beam-wrap">
					<KoboyoIcon name="very-happy-beaming" class="try-beam" />
				</span>
				<h2 class="try-headline whitespace-pre-line">
					{tutorial.homeTitle}
				</h2>
			</div>
			<a href={showPath(locale)} class="try-open doodle-btn doodle-btn-ink">
				{tutorial.open}
			</a>
		</div>
	</div>
	<SiteFooter {copy} {locale} />
</section>
