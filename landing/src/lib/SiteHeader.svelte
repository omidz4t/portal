<script lang="ts">
	import { botUrl, repoUrl } from '$lib/links';

	let { current = '' }: { current?: string } = $props();
</script>

<header class="border-b border-line/80">
	<div class="mx-auto flex max-w-6xl items-center justify-between gap-4 px-5 py-4">
		<a href="/" class="flex items-center gap-3">
			<img src="/avatar.png" alt="" class="h-9 w-9 rounded-lg" />
			<span class="text-sm font-semibold tracking-[0.18em] text-paper">TGPORTAL</span>
		</a>
		<nav class="hidden items-center gap-6 text-sm text-mist sm:flex">
			<a href="/#how" class="hover:text-paper">How it works</a>
			<a href="/#persona" class="hover:text-paper">Persona</a>
			<a href="/docs/" class:text-paper={current === 'docs'} class="hover:text-paper">Docs</a>
			<a href={repoUrl} class="hover:text-paper">Source</a>
		</nav>
		<a
			href={botUrl}
			class="rounded-full bg-teal px-4 py-2 text-sm font-semibold text-ink hover:brightness-110"
		>
			Open Telegram bot
		</a>
	</div>
</header>
