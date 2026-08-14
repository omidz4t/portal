<script lang="ts">
	import KoboyoIcon from '$lib/KoboyoIcon.svelte';
	import type { InboxRow, TutorialScene } from '$lib/content';

	type Bubble = TutorialScene['bubbles'][number];

	let {
		variant,
		app,
		peer,
		logo,
		inbox,
		rows,
		bubbles,
		empty,
		you
	}: {
		variant: 'tg' | 'dc';
		app: string;
		peer: string;
		logo: string;
		inbox: boolean;
		rows: InboxRow[];
		bubbles: Bubble[];
		empty: string;
		you: string;
	} = $props();

	let thread = $state<HTMLElement | null>(null);

	function whoLabel(who: string) {
		if (who === 'you') return you;
		if (who === 'alice') return 'Alice';
		return '';
	}

	function isPortal(name: string) {
		const n = name.toLowerCase();
		return n.includes('portal') || name.includes('پورتال') || n.includes('tgdeltabridge');
	}

	function isBotFather(name: string) {
		return name.toLowerCase().includes('botfather');
	}

	function isAlice(name: string) {
		return (
			name === 'Alice' ||
			name.startsWith('creating Alice') ||
			name.includes('ساخت Alice') ||
			name === 'آلیس'
		);
	}

	const aliceFace = 'https://koboyo.com/icons/svg/face-sunglasses-head.svg';

	const showAliceFace = $derived(
		variant === 'dc' &&
			!inbox &&
			(isAlice(peer) || peer.includes('Alice') || peer.includes('آلیس'))
	);

	function isOutgoing(b: Bubble) {
		if (b.who === 'you') return true;
		if (variant === 'tg' && b.kind === 'sticker' && b.who === 'bot') return true;
		return false;
	}

	function arrowParts(text: string) {
		const parts: { t?: string; dir?: 'left' | 'right' }[] = [];
		const re = /(←|→)/g;
		let last = 0;
		for (const m of text.matchAll(re)) {
			const i = m.index ?? 0;
			if (i > last) parts.push({ t: text.slice(last, i) });
			parts.push({ dir: m[0] === '←' ? 'left' : 'right' });
			last = i + m[0].length;
		}
		if (last < text.length) parts.push({ t: text.slice(last) });
		return parts;
	}

	$effect(() => {
		void bubbles.length;
		void inbox;
		queueMicrotask(() => {
			if (thread) thread.scrollTop = thread.scrollHeight;
		});
	});
</script>

<article class="phone {variant === 'tg' ? 'phone-tg' : 'phone-dc'}" aria-label={app}>
	<header class="phone-bar">
		{#if showAliceFace}
			<span class="phone-avatar">
				<img class="app-logo" src={aliceFace} alt="" width="28" height="28" />
			</span>
		{:else}
			<img
				class="app-logo"
				class:logo-tg-dark={!inbox && isPortal(peer)}
				src={logo}
				alt=""
				width="28"
				height="28"
			/>
		{/if}
		<div>
			<p class="phone-app">{app}</p>
			<p class="phone-peer">
				{#each arrowParts(peer) as p}
					{#if p.dir}
						<KoboyoIcon
							name="move-left"
							class="phone-peer-arrow{p.dir === 'right' ? ' phone-arrow-right' : ''}"
						/>
					{:else}
						{p.t}
					{/if}
				{/each}
			</p>
		</div>
	</header>
	<div class="phone-thread h-full" class:phone-inbox={inbox} bind:this={thread}>
		{#if inbox}
			{#each rows as row}
				<div class="inbox-row" class:is-focus={row.focus}>
					<span class="phone-avatar">
						{#if isAlice(row.name)}
							<img class="app-logo" src={aliceFace} alt="" width="28" height="28" />
						{:else if isPortal(row.name) || isBotFather(row.name)}
							<img
								class="app-logo"
								class:logo-tg-dark={isPortal(row.name)}
								src="/logos/telegram.svg"
								alt=""
								width="28"
								height="28"
							/>
						{:else}
							<KoboyoIcon name={row.icon} class="h-8 w-8" />
						{/if}
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
		{:else if bubbles.length === 0}
			<p class="phone-empty">{empty}</p>
		{:else}
		{#each bubbles as b}
			<div
				class="bubble"
				class:bubble-you={isOutgoing(b)}
				class:bubble-them={variant === 'tg' ? !isOutgoing(b) : !isOutgoing(b) && b.who !== 'sys'}
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
						<a class="invite-link" href={b.link} rel="nofollow noopener">{b.link}</a>
						<p class="whitespace-pre-wrap">
							{#each arrowParts(b.text) as p}
								{#if p.dir}
									<KoboyoIcon
										name="move-left"
										class="phone-peer-arrow{p.dir === 'right' ? ' phone-arrow-right' : ''}"
									/>
								{:else}
									{p.t}
								{/if}
							{/each}
						</p>
					</div>
				{:else}
					{#if whoLabel(b.who) && b.who !== 'you'}
						<span class="bubble-who">{whoLabel(b.who)}</span>
					{/if}
					<p class="whitespace-pre-wrap">
						{#each arrowParts(b.text) as p}
							{#if p.dir}
								<KoboyoIcon
									name="move-left"
									class="phone-peer-arrow{p.dir === 'right' ? ' phone-arrow-right' : ''}"
								/>
							{:else}
								{p.t}
							{/if}
						{/each}
					</p>
				{/if}
			</div>
		{/each}
		{/if}
	</div>
</article>
