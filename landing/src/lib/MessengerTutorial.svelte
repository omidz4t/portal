<script lang="ts">
	import KoboyoIcon from '$lib/KoboyoIcon.svelte';
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

	function whoLabel(who: string) {
		if (who === 'you') return tutorial.you;
		if (who === 'alice') return 'Alice';
		return '';
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

	function inboxRows(side: 'tg' | 'dc') {
		return lastInbox(side);
	}

	let tgThread = $state<HTMLElement | null>(null);
	let dcThread = $state<HTMLElement | null>(null);

	$effect(() => {
		void step;
		void story.id;
		queueMicrotask(() => {
			for (const el of [tgThread, dcThread]) {
				if (el) el.scrollTop = el.scrollHeight;
			}
		});
	});
</script>

<section id="try" class="try-band">
	<div class="try-inner mx-auto w-full max-w-6xl px-5">
		<h2 class="text-2xl font-semibold sm:text-3xl">{tutorial.title}</h2>

		<div class="mt-6 flex flex-wrap gap-2" role="tablist" aria-label={tutorial.title}>
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
			<p class="mt-5 max-w-3xl text-mist">{story.intro}</p>
		{/if}

		<p class="mt-6 max-w-3xl text-lg leading-relaxed font-semibold" aria-live="polite">
			{scene.caption}
		</p>
		<p class="mt-1 font-mono text-xs text-mist">{stepLabel}</p>

		<div class="phones mt-6">
			<article class="phone phone-tg" aria-label={tutorial.telegram}>
				<header class="phone-bar">
					<img
						class="app-logo"
						src="/logos/telegram.svg"
						alt=""
						width="28"
						height="28"
					/>
					<div>
						<p class="phone-app">{tutorial.telegram}</p>
						<p class="phone-peer">{scene.tgTitle}</p>
					</div>
				</header>
				<div
					class="phone-thread"
					class:phone-inbox={isInbox('tg')}
					bind:this={tgThread}
				>
					{#if isInbox('tg')}
						{#each inboxRows('tg') as row}
							<div class="inbox-row" class:is-focus={row.focus}>
								<span class="phone-avatar">
									<KoboyoIcon name={row.icon} class="h-8 w-8" />
								</span>
								<div class="inbox-copy">
									<div class="inbox-top">
										<p class="inbox-name">{row.name}</p>
										<p class="inbox-when">{row.when}</p>
									</div>
									<p class="inbox-preview">{row.preview}</p>
								</div>
							</div>
						{/each}
					{:else if bubbles('tg').length === 0}
						<p class="phone-empty">{tutorial.emptyTg}</p>
					{/if}
					{#each bubbles('tg') as b}
						<div
							class="bubble"
							class:bubble-you={b.who === 'you'}
							class:bubble-them={b.who !== 'you'}
							class:bubble-sys={b.who === 'sys'}
						>
							{#if b.kind === 'sticker'}
								<span class="sticker-chip">
									<KoboyoIcon name="sticker" class="h-8 w-8" />
								</span>
							{:else if b.kind === 'invite'}
								<div class="invite-card">
									<img
										class="invite-qr"
										src="/logos/invite-example.svg"
										alt=""
										width="120"
										height="120"
									/>
									<a class="invite-link" href={b.link} rel="nofollow noopener"
										>{b.link}</a
									>
									<p class="whitespace-pre-wrap">{b.text}</p>
								</div>
							{:else}
								{#if whoLabel(b.who) && b.who !== 'you'}
									<span class="bubble-who">{whoLabel(b.who)}</span>
								{/if}
								<p class="whitespace-pre-wrap">{b.text}</p>
							{/if}
						</div>
					{/each}
				</div>
			</article>

			<article class="phone phone-dc" aria-label={tutorial.delta}>
				<header class="phone-bar">
					<img
						class="app-logo"
						src="/logos/deltachat.svg"
						alt=""
						width="28"
						height="28"
					/>
					<div>
						<p class="phone-app">{tutorial.delta}</p>
						<p class="phone-peer">{scene.dcTitle}</p>
					</div>
				</header>
				<div
					class="phone-thread"
					class:phone-inbox={isInbox('dc')}
					bind:this={dcThread}
				>
					{#if isInbox('dc')}
						{#each inboxRows('dc') as row}
							<div class="inbox-row" class:is-focus={row.focus}>
								<span class="phone-avatar">
									<KoboyoIcon name={row.icon} class="h-8 w-8" />
								</span>
								<div class="inbox-copy">
									<div class="inbox-top">
										<p class="inbox-name">{row.name}</p>
										<p class="inbox-when">{row.when}</p>
									</div>
									<p class="inbox-preview">{row.preview}</p>
								</div>
							</div>
						{/each}
					{:else if bubbles('dc').length === 0}
						<p class="phone-empty">{tutorial.emptyDc}</p>
					{/if}
					{#each bubbles('dc') as b}
						<div
							class="bubble"
							class:bubble-you={b.who === 'you'}
							class:bubble-them={b.who !== 'you' && b.who !== 'sys'}
							class:bubble-sys={b.who === 'sys'}
						>
							{#if b.kind === 'sticker'}
								<span class="sticker-chip">
									<KoboyoIcon name="sticker" class="h-8 w-8" />
								</span>
							{:else if b.kind === 'invite'}
								<div class="invite-card">
									<img
										class="invite-qr"
										src="/logos/invite-example.svg"
										alt=""
										width="120"
										height="120"
									/>
									<a class="invite-link" href={b.link} rel="nofollow noopener"
										>{b.link}</a
									>
									<p class="whitespace-pre-wrap">{b.text}</p>
								</div>
							{:else}
								{#if whoLabel(b.who) && b.who !== 'you'}
									<span class="bubble-who">{whoLabel(b.who)}</span>
								{/if}
								<p class="whitespace-pre-wrap">{b.text}</p>
							{/if}
						</div>
					{/each}
				</div>
			</article>
		</div>

		{#if scene.why}
			<aside class="doodle-card mt-6 max-w-3xl p-5" aria-label={tutorial.whyLabel}>
				<p class="font-mono text-xs uppercase">{tutorial.whyLabel}</p>
				<p class="mt-2 leading-relaxed">{scene.why}</p>
			</aside>
		{/if}

		<div class="mt-6 flex flex-wrap gap-2">
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
</section>
