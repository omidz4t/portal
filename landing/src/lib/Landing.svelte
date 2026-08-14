<script lang="ts">
	import { botUrl, repoUrl } from '$lib/links';
	import KoboyoIcon from '$lib/KoboyoIcon.svelte';
	import type { Copy } from '$lib/content';

	let { copy }: { copy: Copy } = $props();

	const flowIcons = ['telegram', 'bot', 'ghost', 'paper-plane'] as const;
</script>

<svelte:head>
	<title>{copy.metaTitle}</title>
	<meta name="description" content={copy.metaDescription} />
</svelte:head>

<div>
	<aside class="doodle-warn" role="note" aria-label={copy.warnTitle}>
		<div class="mx-auto flex max-w-6xl gap-3 px-5 py-4 text-sm leading-relaxed text-warn">
			<KoboyoIcon name="warning" class="mt-0.5 h-6 w-6" />
			<p>
				<strong class="font-semibold text-ink">{copy.warnTitle}</strong>
				{copy.warnBody}
				<a class="font-semibold underline underline-offset-2" href="/docs/self-host/"
					>{copy.ctaHost}</a
				>.
				<a class="underline underline-offset-2" href="/docs/trust/">{copy.warnWhy}</a>
			</p>
		</div>
	</aside>

	<section class="mx-auto grid max-w-6xl items-center gap-12 px-5 py-16 lg:grid-cols-2 lg:py-20">
		<div>
			<p class="font-mono text-xs tracking-[0.18em] uppercase">{copy.heroEyebrow}</p>
			<h1 class="mt-3 text-4xl leading-tight font-semibold text-balance sm:text-5xl">
				{copy.heroTitle}
			</h1>
			<p class="mt-5 max-w-xl text-lg leading-relaxed text-mist">{copy.heroLead}</p>
			<div class="mt-8 flex flex-wrap gap-3">
				<a href="/docs/self-host/" class="doodle-btn doodle-btn-ink">{copy.ctaHost}</a>
				<a href={botUrl} rel="noreferrer" class="doodle-btn">{copy.ctaPublic}</a>
				<a href="/docs/" class="doodle-btn doodle-btn-ghost">{copy.ctaDocs}</a>
			</div>
		</div>
		<figure class="doodle-card doodle-card-tilt overflow-hidden">
			<img
				src="/poster.jpg"
				alt={copy.posterAlt}
				width="960"
				height="640"
				class="aspect-3/2 h-full w-full object-cover"
			/>
		</figure>
	</section>

	<section class="mx-auto max-w-6xl px-5 pb-16">
		<h2 class="text-2xl font-semibold">{copy.bridgeTitle}</h2>
		<p class="mt-2 max-w-2xl text-mist">{copy.bridgeLead}</p>
		<div class="doodle-card mt-8 overflow-x-auto">
			<table class="w-full text-left text-sm">
				<thead class="font-mono text-xs tracking-wide uppercase">
					<tr>
						<th class="px-4 py-3">{copy.colTelegram}</th>
						<th class="px-4 py-3">{copy.colDelta}</th>
					</tr>
				</thead>
				<tbody>
					{#each copy.media as row}
						<tr class="border-t-2 border-ink/15">
							<td class="px-4 py-3">{row.tg}</td>
							<td class="px-4 py-3 text-mist">{row.dc}</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	</section>

	<section id="how" class="mx-auto max-w-6xl px-5 pb-16">
		<h2 class="text-2xl font-semibold">{copy.howTitle}</h2>
		<p class="mt-2 max-w-2xl text-mist">
			{copy.howLead}
			<a class="underline underline-offset-2" href="/docs/pairing/">{copy.pairingDocs}</a>.
		</p>
		<ol class="mt-10 grid gap-5 md:grid-cols-2">
			{#each copy.steps as step, i}
				<li
					class="doodle-card p-6"
					class:doodle-card-tilt={i % 2 === 0}
					class:doodle-card-tilt-r={i % 2 === 1}
				>
					<p class="font-mono text-xs">{step.n}</p>
					<h3 class="mt-2 text-lg font-semibold">{step.title}</h3>
					<p class="mt-2 text-sm leading-relaxed text-mist">{step.body}</p>
				</li>
			{/each}
		</ol>
	</section>

	<section id="persona" class="mx-auto max-w-6xl px-5 pb-16">
		<p class="font-mono text-xs tracking-[0.18em] uppercase">{copy.personaEyebrow}</p>
		<h2 class="mt-2 text-2xl font-semibold text-balance sm:text-3xl">{copy.personaTitle}</h2>
		<p class="mt-4 max-w-3xl text-lg leading-relaxed text-mist">{copy.personaLead}</p>

		<div class="mt-10 grid gap-4 md:grid-cols-2">
			<article class="doodle-card p-6">
				<div class="flex items-center gap-3">
					<KoboyoIcon name="message" class="h-10 w-10" />
					<h3 class="text-lg font-semibold">{copy.personalCardTitle}</h3>
				</div>
				<p class="mt-3 text-sm leading-relaxed text-mist">{copy.personalCardBody}</p>
			</article>
			<article class="doodle-card doodle-card-tilt-r p-6">
				<div class="flex items-center gap-3">
					<KoboyoIcon name="ghost" class="h-10 w-10" />
					<h3 class="text-lg font-semibold">{copy.personaCardTitle}</h3>
				</div>
				<p class="mt-3 text-sm leading-relaxed text-mist">{copy.personaCardBody}</p>
			</article>
		</div>

		<h3 class="mt-14 text-xl font-semibold">{copy.flowTitle}</h3>
		<p class="mt-2 max-w-2xl text-mist">{copy.flowLead}</p>
		<ol class="mt-8 grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
			{#each copy.flow as step, i}
				<li class="doodle-card p-5">
					<KoboyoIcon name={flowIcons[i]} class="h-14 w-14" />
					<p class="mt-4 font-mono text-xs">{step.n}</p>
					<h4 class="mt-1 text-lg font-semibold">{step.title}</h4>
					<p class="mt-2 text-sm leading-relaxed text-mist">{step.body}</p>
				</li>
			{/each}
		</ol>

		<div class="mt-10 grid gap-5 md:grid-cols-3">
			<article class="doodle-card p-6">
				<KoboyoIcon name="key" class="h-10 w-10" />
				<h3 class="mt-4 text-lg font-semibold">{copy.setupTitle}</h3>
				<p class="mt-2 text-sm leading-relaxed text-mist">{copy.setupBody}</p>
			</article>
			<article class="doodle-card p-6">
				<KoboyoIcon name="users" class="h-10 w-10" />
				<h3 class="mt-4 text-lg font-semibold">{copy.groupsTitle}</h3>
				<p class="mt-2 text-sm leading-relaxed text-mist">{copy.groupsBody}</p>
			</article>
			<article class="doodle-card p-6">
				<KoboyoIcon name="lock" class="h-10 w-10" />
				<h3 class="mt-4 text-lg font-semibold">{copy.hostTitle}</h3>
				<p class="mt-2 text-sm leading-relaxed text-mist">
					{copy.hostBody}
					<a class="underline underline-offset-2" href="/docs/self-host/">{copy.ctaHost}</a>.
				</p>
			</article>
		</div>

		<p class="mt-8 text-sm text-mist">
			{copy.personaCommands}
			<a class="underline underline-offset-2" href="/docs/persona/">{copy.personaDocs}</a>.
		</p>
	</section>

	<section id="self-host" class="mx-auto max-w-6xl px-5 pb-20">
		<div class="grid gap-10 lg:grid-cols-2">
			<div>
				<h2 class="text-2xl font-semibold">{copy.selfTitle}</h2>
				<p class="mt-3 leading-relaxed text-mist">{copy.selfLead}</p>
				<ul class="mt-5 list-disc space-y-2 ps-5 text-sm text-mist">
					{#each copy.selfNeeds as item}
						<li>{item}</li>
					{/each}
				</ul>
				<p class="mt-4 text-sm">
					<a class="underline underline-offset-2" href="/docs/self-host/">{copy.selfWalk}</a>
					·
					<a class="underline underline-offset-2" href={repoUrl}>GitHub</a>
				</p>
			</div>
			<pre class="doodle-card overflow-x-auto p-5 font-mono text-sm leading-7"><code
					>make config
# .env → TELEGRAM_BOT_TOKEN=…
#        PERSONA_ACCOUNT_QR=…
make init QR=dcaccount:nine.testrun.org
make serve

make run-landing</code
				></pre>
		</div>
	</section>
</div>
