<script lang="ts">
	import ChatPhone from '$lib/ChatPhone.svelte';
	import KoboyoIcon from '$lib/KoboyoIcon.svelte';
	import type { TutorialCopy, TutorialScene } from '$lib/content';

	let { tutorial }: { tutorial: TutorialCopy } = $props();

	let storyId = $state<string | null>(null);
	let step = $state(0);
	let dialog = $state<HTMLDialogElement | null>(null);

	const story = $derived(
		tutorial.stories.find((s) => s.id === (storyId ?? tutorial.stories[0].id)) ??
			tutorial.stories[0]
	);
	const scene = $derived(
		story.scenes[Math.min(step, story.scenes.length - 1)] as TutorialScene
	);
	const total = $derived(story.scenes.length);
	const stepLabel = $derived(
		tutorial.stepOf.replace('{n}', String(step + 1)).replace('{total}', String(total))
	);

	function pick(id: string) {
		storyId = id;
		step = 0;
	}

	function next() {
		if (step < story.scenes.length - 1) step += 1;
	}

	function back() {
		if (step > 0) step -= 1;
	}

	function restart() {
		step = 0;
	}

	function openShow() {
		dialog?.showModal();
	}

	function closeShow() {
		dialog?.close();
	}

	function onDialogClick(e: MouseEvent) {
		if (e.target === dialog) closeShow();
	}

	function bubbleKey(b: TutorialScene['bubbles'][number]) {
		return `${b.side}|${b.who}|${b.kind ?? ''}|${b.text}|${b.link ?? ''}`;
	}

	function bubbles(side: 'tg' | 'dc') {
		const seen = new Set<string>();
		const out: TutorialScene['bubbles'] = [];
		for (let i = 0; i <= step && i < story.scenes.length; i++) {
			for (const b of story.scenes[i].bubbles) {
				if (b.side !== side) continue;
				const key = bubbleKey(b);
				if (seen.has(key)) continue;
				seen.add(key);
				out.push(b);
			}
		}
		return out;
	}

	function listTitle(title: string) {
		return title === 'Chats' || title === 'چت‌ها' || title === '—';
	}

	function lastInbox(side: 'tg' | 'dc') {
		for (let i = step; i >= 0; i--) {
			const s = story.scenes[i] as TutorialScene;
			const rows = side === 'tg' ? s.inboxTg : s.inboxDc;
			if (Array.isArray(rows)) return rows;
		}
		return [];
	}

	function isInbox(side: 'tg' | 'dc') {
		if (bubbles(side).length > 0) return false;
		const title = side === 'tg' ? scene.tgTitle : scene.dcTitle;
		if (!listTitle(title) && scene.view !== 'inbox') return false;
		return lastInbox(side).length > 0;
	}
</script>

<section id="try" class="try-band">
	<div class="try-cta mx-auto flex w-full max-w-6xl flex-col items-start justify-center px-5">
		<div class="try-cta-row">
			<h2 class="try-headline">
				{tutorial.title}
			</h2>
			<KoboyoIcon name="very-happy-beaming" class="try-beam" />
		</div>
		<button type="button" class="try-open doodle-btn doodle-btn-ink" onclick={openShow}>
			{tutorial.open}
		</button>
	</div>
</section>

<dialog class="try-modal" bind:this={dialog} onclick={onDialogClick} aria-labelledby="try-modal-title">
	<div class="try-modal-inner">
		<div class="try-modal-bar">
			<h3 id="try-modal-title" class="text-lg font-semibold sm:text-xl">{tutorial.title}</h3>
			<button type="button" class="doodle-btn" onclick={closeShow}>{tutorial.close}</button>
		</div>

		<div class="mt-4 flex flex-wrap gap-2" role="tablist" aria-label={tutorial.title}>
			{#each tutorial.stories as s}
				<button
					type="button"
					class="doodle-btn"
					class:doodle-btn-ink={s.id === story.id}
					role="tab"
					aria-selected={s.id === story.id}
					onclick={() => pick(s.id)}
				>
					{s.label}
				</button>
			{/each}
		</div>

		{#if story.intro}
			<p class="mt-4 max-w-3xl text-mist">{story.intro}</p>
		{/if}

		<p class="mt-4 max-w-3xl text-lg leading-relaxed font-semibold" aria-live="polite">
			{scene.caption}
		</p>
		<p class="mt-1 font-mono text-xs text-mist">{stepLabel}</p>

		<div class="phones mt-5">
			<ChatPhone
				variant="tg"
				app={tutorial.telegram}
				peer={scene.tgTitle}
				logo="/logos/telegram.svg"
				inbox={isInbox('tg')}
				rows={lastInbox('tg')}
				bubbles={bubbles('tg')}
				empty={tutorial.emptyTg}
				you={tutorial.you}
			/>
			<ChatPhone
				variant="dc"
				app={tutorial.delta}
				peer={scene.dcTitle}
				logo="/logos/deltachat.svg"
				inbox={isInbox('dc')}
				rows={lastInbox('dc')}
				bubbles={bubbles('dc')}
				empty={tutorial.emptyDc}
				you={tutorial.you}
			/>
		</div>

		{#if scene.why}
			<aside class="doodle-card mt-5 max-w-3xl p-5" aria-label={tutorial.whyLabel}>
				<p class="font-mono text-xs uppercase">{tutorial.whyLabel}</p>
				<p class="mt-2 leading-relaxed">{scene.why}</p>
			</aside>
		{/if}

		<div class="mt-5 flex flex-wrap gap-2">
			<button type="button" class="doodle-btn" onclick={back} disabled={step === 0}
				>{tutorial.back}</button
			>
			<button
				type="button"
				class="doodle-btn doodle-btn-ink"
				onclick={next}
				disabled={step >= total - 1}>{tutorial.next}</button
			>
			<button type="button" class="doodle-btn doodle-btn-ghost" onclick={restart}
				>{tutorial.restart}</button
			>
		</div>
	</div>
</dialog>
