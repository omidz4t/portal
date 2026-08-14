<script lang="ts">
	import { botUrl, repoUrl } from '$lib/links';

	let { current = '' }: { current?: string } = $props();

	const links = [
		{ href: '/#how', label: 'How it works', id: 'how' },
		{ href: '/#persona', label: 'Persona', id: 'persona' },
		{ href: '/docs/', label: 'Docs', id: 'docs' },
		{ href: repoUrl, label: 'Source', id: 'source', external: true }
	];
</script>

<header class="border-b border-line/80">
	<div class="mx-auto flex max-w-6xl items-center justify-between gap-3 px-5 py-3">
		<a href="/" class="flex min-h-11 items-center gap-3 rounded-lg px-1">
			<img
				src="/avatar.png"
				alt=""
				width="36"
				height="36"
				class="h-9 w-9 rounded-lg"
				aria-hidden="true"
			/>
			<span class="text-sm font-semibold tracking-[0.18em] text-paper">TGPORTAL</span>
		</a>

		<nav class="hidden items-center gap-1 md:flex" aria-label="Primary">
			{#each links as link}
				<a
					href={link.href}
					class="inline-flex min-h-11 items-center rounded-lg px-3 text-sm text-mist hover:text-paper"
					class:font-semibold={current === link.id}
					class:text-paper={current === link.id}
					class:underline={current === link.id}
					class:underline-offset-4={current === link.id}
					aria-current={current === link.id ? 'page' : undefined}
					rel={link.external ? 'noreferrer' : undefined}
				>
					{link.label}
				</a>
			{/each}
		</nav>

		<a
			href={botUrl}
			rel="noreferrer"
			class="hidden min-h-11 items-center rounded-full bg-teal px-4 text-sm font-semibold text-ink hover:brightness-110 md:inline-flex"
		>
			Open Telegram bot
		</a>

		<details class="relative md:hidden">
			<summary
				class="inline-flex min-h-11 min-w-11 cursor-pointer list-none items-center justify-center rounded-lg border border-line px-3 text-sm font-medium text-paper"
			>
				Menu
			</summary>
			<div
				class="absolute right-0 z-20 mt-2 w-56 rounded-xl border border-line bg-panel p-2 shadow-lg"
			>
				<nav class="flex flex-col" aria-label="Mobile">
					{#each links as link}
						<a
							href={link.href}
							class="inline-flex min-h-11 items-center rounded-lg px-3 text-sm text-mist hover:bg-ink hover:text-paper"
							aria-current={current === link.id ? 'page' : undefined}
							rel={link.external ? 'noreferrer' : undefined}
						>
							{link.label}
						</a>
					{/each}
					<a
						href={botUrl}
						rel="noreferrer"
						class="mt-1 inline-flex min-h-11 items-center justify-center rounded-lg bg-teal px-3 text-sm font-semibold text-ink"
					>
						Open Telegram bot
					</a>
				</nav>
			</div>
		</details>
	</div>
</header>
