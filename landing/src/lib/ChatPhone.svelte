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
		<img class="app-logo" src={logo} alt="" width="28" height="28" />
		<div>
			<p class="phone-app">{app}</p>
			<p class="phone-peer">{peer}</p>
		</div>
	</header>
	<div class="phone-thread" class:phone-inbox={inbox} bind:this={thread}>
		{#if inbox}
			{#each rows as row}
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
		{:else if bubbles.length === 0}
			<p class="phone-empty">{empty}</p>
		{/if}
		{#each bubbles as b}
			<div
				class="bubble"
				class:bubble-you={b.who === 'you'}
				class:bubble-them={variant === 'tg' ? b.who !== 'you' : b.who !== 'you' && b.who !== 'sys'}
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
