<script lang="ts">
	import DocLayout from '$lib/DocLayout.svelte';
	import { repoDocs } from '$lib/links';

	const pages = [
		{
			href: '/docs/trust/',
			title: 'Trust the operator',
			body: 'The host sees bridged traffic. Prefer running your own instance.'
		},
		{
			href: '/docs/persona/',
			title: 'Persona mode',
			body: 'Your BotFather bot, one ghost Delta Chat account per Telegram user, optional groups.'
		},
		{
			href: '/docs/pairing/',
			title: 'Pairing (personal)',
			body: 'Telegram-first and Delta Chat-first pairing, commands, sticker packs.'
		},
		{
			href: '/docs/self-host/',
			title: 'Self-host',
			body: 'Config, init, serve, and what must stay off git.'
		}
	];
</script>

<svelte:head>
	<title>Docs — TGPORTAL</title>
</svelte:head>

<div class="mx-auto max-w-3xl px-5 py-16">
	<DocLayout
		title="Documentation"
		lede="How the bridge works, why you should host it yourself, and how persona ghosts differ from a single pair."
	>
		<ul class="not-prose mt-8 grid gap-4">
			{#each pages as p}
				<li>
					<a href={p.href} class="doodle-card block p-5">
						<h2 class="text-lg font-semibold text-ink">{p.title}</h2>
						<p class="mt-1 text-sm text-mist">{p.body}</p>
					</a>
				</li>
			{/each}
		</ul>
		<p class="mt-8">
			Maintainer-level notes (architecture, config keys, security hardening) live in the repo
			<a href={repoDocs}>docs/</a> directory.
		</p>
	</DocLayout>
</div>
