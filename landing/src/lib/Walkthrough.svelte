<script lang="ts">
	import ChatPhone from '$lib/ChatPhone.svelte';
	import type { TutorialCopy, TutorialScene } from '$lib/content';
	import { asset } from '$lib/links';

	let {
		tutorial,
		hideTitle = false,
		homeHref = ''
	}: { tutorial: TutorialCopy; hideTitle?: boolean; homeHref?: string } = $props();

	let storyId = $state<string | null>(null);
	let step = $state(0);

	const story = $derived(
		tutorial.stories.find((s) => s.id === (storyId ?? tutorial.stories[0].id)) ??
			tutorial.stories[0]
	);
	const tgApp = $derived(
		'telegram' in story && story.telegram ? story.telegram : tutorial.telegram
	);
	const scene = $derived(
		story.scenes[Math.min(step, story.scenes.length - 1)] as TutorialScene
	);
	const phoneSide = $derived.by(() => {
		const text = `${scene.caption} ${scene.why ?? ''}`;
		const bTg = scene.bubbles.some((b) => b.side === 'tg');
		const bDc = scene.bubbles.some((b) => b.side === 'dc');
		if (bTg !== bDc) return bTg ? 'tg' : 'dc';
		let last: 'tg' | 'dc' | null = null;
		const re = /delta chat|دلتا چت|arcanechat|telegram|تلگرام|right phone|گوشی راست/gi;
		for (const m of text.matchAll(re)) {
			const w = m[0].toLowerCase();
			last =
				w.includes('delta') ||
				w.includes('دلتا') ||
				w.includes('arcane') ||
				w.includes('right') ||
				w.includes('راست')
					? 'dc'
					: 'tg';
		}
		if (last) return last;
		if (isAlice(scene.dcTitle) || scene.dcTitle.includes('ساخت')) return 'dc';
		return 'tg';
	});

	function isAlice(name: string) {
		return name === 'Alice' || name === 'آلیس';
	}
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

	const hiRe =
		/PERSONA_ACCOUNT_QR|\/unpair-bot|\/pair-bot|\/start|\/pair|\/bots|@YourBot|@BotFather|BotFather|ArcaneChat|Delta Chat|Telegram|Portal|K7H2MQNP|tgportal|دلتا چت|تلگرام|پورتال|ربات پورتال/gi;

	function hiParts(text: string) {
		const parts: { t: string; mark: boolean }[] = [];
		let last = 0;
		for (const m of text.matchAll(hiRe)) {
			const i = m.index ?? 0;
			if (i > last) parts.push({ t: text.slice(last, i), mark: false });
			parts.push({ t: m[0], mark: true });
			last = i + m[0].length;
		}
		if (last < text.length) parts.push({ t: text.slice(last), mark: false });
		return parts;
	}

	function isInbox(side: 'tg' | 'dc') {
		const title = side === 'tg' ? scene.tgTitle : scene.dcTitle;
		if (!listTitle(title) && scene.view !== 'inbox') return false;
		if (!listTitle(title)) return false;
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

	const showWhyLabel = $derived.by(() => {
		if (step !== 0) return false;
		const group = storyGroups.find((g) => g.stories.some((s) => s.id === story.id));
		return group?.stories[0]?.id === story.id;
	});

	const whyPacks = $derived.by(() => {
		const firstIds = new Set(storyGroups.map((g) => g.stories[0]?.id));
		return tutorial.stories.flatMap((s) =>
			s.scenes.map((sc, i) => ({
				label: i === 0 && firstIds.has(s.id),
				intro: i === 0 && s.intro ? s.intro : '',
				caption: sc.caption,
				why: sc.why ?? ''
			}))
		);
	});

	let whyMin = $state(0);
	let whyMeasure = $state<HTMLElement | undefined>();

	$effect(() => {
		void whyPacks;
		const root = whyMeasure;
		if (!root) return;
		function measure() {
			let max = 0;
			for (const child of root.children) {
				max = Math.max(max, (child as HTMLElement).offsetHeight);
			}
			if (max > 0) whyMin = max;
		}
		const id = requestAnimationFrame(measure);
		const ro = new ResizeObserver(measure);
		ro.observe(root);
		window.addEventListener('resize', measure);
		return () => {
			cancelAnimationFrame(id);
			ro.disconnect();
			window.removeEventListener('resize', measure);
		};
	});
</script>

<div class="try-walk">
	<div class="try-walk-head">
		{#if homeHref}
			<div class="try-page-bar">
				<a href={homeHref} class="doodle-btn try-page-back">{tutorial.homeBack}</a>
				<h1 class="try-headline try-page-title">{tutorial.title}</h1>
			</div>
		{:else if !hideTitle}
			<h1 class="try-headline whitespace-pre-line">{tutorial.title}</h1>
		{/if}
		<div class="try-story-nav">
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
	</div>

	<div class="phones mt-5" class:phones-show-tg={phoneSide === 'tg'} class:phones-show-dc={phoneSide === 'dc'}>
		<ChatPhone
			variant="tg"
			app={tgApp}
			peer={scene.tgTitle}
			logo={asset('/logos/telegram.svg')}
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
			logo={asset('/logos/deltachat.svg')}
			inbox={isInbox('dc')}
			rows={lastInbox('dc')}
			bubbles={bubbles('dc')}
			empty={tutorial.emptyDc}
			you={tutorial.you}
		/>
	</div>

	<aside
		class="doodle-card try-walk-why"
		aria-label={tutorial.whyLabel}
		aria-live="polite"
		style:min-height={whyMin ? `${whyMin}px` : undefined}
	>
		{#if showWhyLabel}
			<p class="try-walk-why-label">{tutorial.whyLabel}</p>
		{/if}
		{#if story.intro && step === 0}
			<p class="try-walk-why-intro">
				{#each hiParts(story.intro) as p}{#if p.mark}<mark>{p.t}</mark>{:else}{p.t}{/if}{/each}
			</p>
		{/if}
		<p class="try-walk-why-caption">
			{#each hiParts(scene.caption) as p}{#if p.mark}<mark>{p.t}</mark>{:else}{p.t}{/if}{/each}
		</p>
		{#if scene.why}
			<p class="try-walk-why-body">
				{#each hiParts(scene.why) as p}{#if p.mark}<mark>{p.t}</mark>{:else}{p.t}{/if}{/each}
			</p>
		{/if}
	</aside>
	<div class="try-walk-why-measure" bind:this={whyMeasure} aria-hidden="true">
		{#each whyPacks as pack}
			<div class="doodle-card try-walk-why">
				{#if pack.label}
					<p class="try-walk-why-label">{tutorial.whyLabel}</p>
				{/if}
				{#if pack.intro}
					<p class="try-walk-why-intro">
						{#each hiParts(pack.intro) as p}{#if p.mark}<mark>{p.t}</mark>{:else}{p.t}{/if}{/each}
					</p>
				{/if}
				<p class="try-walk-why-caption">
					{#each hiParts(pack.caption) as p}{#if p.mark}<mark>{p.t}</mark>{:else}{p.t}{/if}{/each}
				</p>
				{#if pack.why}
					<p class="try-walk-why-body">
						{#each hiParts(pack.why) as p}{#if p.mark}<mark>{p.t}</mark>{:else}{p.t}{/if}{/each}
					</p>
				{/if}
			</div>
		{/each}
	</div>

	<div class="try-walk-nav">
		<p class="try-walk-step">{stepLabel}</p>
		<div class="try-walk-nav-actions">
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
</div>
