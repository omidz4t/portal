<script lang="ts">
	import ChatPhone from '$lib/ChatPhone.svelte';
	import type { TutorialCopy, TutorialScene } from '$lib/content';

	let { tutorial }: { tutorial: TutorialCopy } = $props();

	let storyId = $state<string | null>(null);
	let step = $state(0);

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

	const storyGroups = $derived.by(() => {
		const groups: { name: string; stories: typeof tutorial.stories }[] = [];
		for (const s of tutorial.stories) {
			const name = s.group ?? '';
			const last = groups[groups.length - 1];
			if (last && last.name === name) last.stories.push(s);
			else groups.push({ name, stories: [s] });
		}
		return groups;
	});
</script>

<div class="try-walk">
	<h1 class="try-headline whitespace-pre-line">{tutorial.title}</h1>

	<div class="try-story-nav mt-4">
		{#each storyGroups as g}
			<div class="try-story-group" class:has-label={Boolean(g.name)}>
				{#if g.name}
					<p class="try-story-group-label">{g.name}</p>
				{/if}
				<div class="flex flex-wrap gap-2" role="tablist" aria-label={g.name || tutorial.title}>
					{#each g.stories as s}
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
			</div>
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
