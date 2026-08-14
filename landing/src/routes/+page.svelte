<script lang="ts">
	const botUrl = 'https://t.me/tgdeltabridgebot';
	const repoUrl = 'https://github.com/themadorg/tgportal';

	const media = [
		{ tg: 'Static WEBP stickers', dc: 'Sticker / image' },
		{ tg: 'Lottie TGS', dc: 'File, or GIF if a converter is installed' },
		{ tg: 'Video stickers (WEBM)', dc: 'Video' },
		{ tg: 'GIF / animation (GIF or MP4)', dc: 'GIF / video' },
		{ tg: 'Custom emoji', dc: 'Sticker via getCustomEmojiStickers' },
		{ tg: 'Text, images, short video', dc: 'Same, both directions' }
	];

	const steps = [
		{
			n: '01',
			title: 'Open the Telegram bot',
			body: 'Message @tgdeltabridgebot and send /start in a private chat.'
		},
		{
			n: '02',
			title: 'Open the Delta Chat invite',
			body: 'The bot replies with an invite QR/link and a short pairing code (about 30 minutes).'
		},
		{
			n: '03',
			title: 'Send the code on Delta Chat',
			body: 'Accept the chat, send only the code. The pair becomes active: your Telegram user ↔ that DC chat.'
		},
		{
			n: '04',
			title: 'Send media either way',
			body: 'Stickers, GIFs, text, images, short video. Filenames stay neutral; source captions are stripped.'
		}
	];

	const modes = [
		{
			name: 'personal',
			title: 'Personal pairing',
			body: 'One Telegram user maps to one Delta Chat conversation on the portal bot. Default product.'
		},
		{
			name: 'persona',
			title: 'Persona bots',
			body: 'Pair, then /pair-bot with your BotFather token. Each remote Telegram user gets a stable ghost Delta Chat account that messages you as a normal 1:1.'
		},
		{
			name: 'both',
			title: 'Both',
			body: 'Run classic pairing and user-owned bots on the same instance.'
		}
	];
</script>

<header class="border-b border-line/80">
	<div class="mx-auto flex max-w-6xl items-center justify-between gap-4 px-5 py-4">
		<a href="/" class="flex items-center gap-3">
			<img src="/avatar.png" alt="" class="h-9 w-9 rounded-lg" />
			<span class="text-sm font-semibold tracking-[0.18em] text-paper">TGPORTAL</span>
		</a>
		<nav class="hidden items-center gap-6 text-sm text-mist sm:flex">
			<a href="#how" class="hover:text-paper">How it works</a>
			<a href="#modes" class="hover:text-paper">Modes</a>
			<a href="#self-host" class="hover:text-paper">Self-host</a>
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

<main>
	<section class="mx-auto grid max-w-6xl items-center gap-12 px-5 py-16 lg:grid-cols-2 lg:py-24">
		<div>
			<p class="font-mono text-xs tracking-[0.22em] text-teal uppercase">Telegram ↔ Delta Chat</p>
			<h1 class="mt-4 text-4xl leading-tight font-semibold tracking-tight sm:text-5xl">
				Stickers and GIFs that survive the jump to Delta Chat.
			</h1>
			<p class="mt-5 max-w-xl text-lg leading-relaxed text-mist">
				TGPORTAL is a Go bridge. Pair once with a short code. After that, media you send the
				Telegram bot lands in your Delta Chat conversation — and the other way around — without
				noisy filenames or “from Telegram” captions.
			</p>
			<div class="mt-8 flex flex-wrap gap-3">
				<a
					href={botUrl}
					class="rounded-full bg-sky px-5 py-2.5 text-sm font-semibold text-ink hover:brightness-110"
				>
					Pair with @tgdeltabridgebot
				</a>
				<a
					href={repoUrl}
					class="rounded-full border border-line px-5 py-2.5 text-sm font-medium text-paper hover:border-mist"
				>
					GitHub
				</a>
			</div>
		</div>
		<figure class="overflow-hidden rounded-2xl border border-line bg-panel shadow-2xl">
			<img src="/poster.jpg" alt="TGPORTAL branding poster" class="h-full w-full object-cover" />
		</figure>
	</section>

	<section class="border-y border-line/80 bg-panel/40">
		<div class="mx-auto max-w-6xl px-5 py-16">
			<h2 class="text-2xl font-semibold">What it bridges</h2>
			<p class="mt-2 max-w-2xl text-mist">
				Bidirectional, per-user pairing. Telegram types stay useful on Delta Chat.
			</p>
			<div class="mt-8 overflow-hidden rounded-xl border border-line">
				<table class="w-full text-left text-sm">
					<thead class="bg-ink/60 font-mono text-xs tracking-wide text-mist uppercase">
						<tr>
							<th class="px-4 py-3">Telegram</th>
							<th class="px-4 py-3">Delta Chat</th>
						</tr>
					</thead>
					<tbody>
						{#each media as row}
							<tr class="border-t border-line">
								<td class="px-4 py-3">{row.tg}</td>
								<td class="px-4 py-3 text-mist">{row.dc}</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		</div>
	</section>

	<section id="how" class="mx-auto max-w-6xl px-5 py-16">
		<h2 class="text-2xl font-semibold">Pair once</h2>
		<p class="mt-2 max-w-2xl text-mist">
			The live bot is
			<a class="text-teal underline-offset-2 hover:underline" href={botUrl}>@tgdeltabridgebot</a>.
			You can also start from Delta Chat: send any message to the bot, then open the
			<code class="font-mono text-paper">t.me/…?start=CODE</code> link it gives you.
		</p>
		<ol class="mt-10 grid gap-5 md:grid-cols-2">
			{#each steps as step}
				<li class="rounded-2xl border border-line bg-panel p-6">
					<p class="font-mono text-xs text-teal">{step.n}</p>
					<h3 class="mt-2 text-lg font-semibold">{step.title}</h3>
					<p class="mt-2 text-sm leading-relaxed text-mist">{step.body}</p>
				</li>
			{/each}
		</ol>
	</section>

	<section id="modes" class="border-y border-line/80 bg-panel/40">
		<div class="mx-auto max-w-6xl px-5 py-16">
			<h2 class="text-2xl font-semibold">Two products, one binary</h2>
			<p class="mt-2 max-w-2xl text-mist">
				<code class="font-mono text-paper">mode</code> in
				<code class="font-mono text-paper">config.yml</code> is
				<code class="font-mono text-paper">personal</code>,
				<code class="font-mono text-paper">persona</code>, or
				<code class="font-mono text-paper">both</code>.
			</p>
			<div class="mt-8 grid gap-5 md:grid-cols-3">
				{#each modes as mode}
					<article class="rounded-2xl border border-line bg-ink/50 p-6">
						<p class="font-mono text-xs text-sky">{mode.name}</p>
						<h3 class="mt-2 text-lg font-semibold">{mode.title}</h3>
						<p class="mt-2 text-sm leading-relaxed text-mist">{mode.body}</p>
					</article>
				{/each}
			</div>
			<p class="mt-6 text-sm text-mist">
				Persona groups can be mirrored as Delta Chat groups named
				<code class="font-mono text-paper">TG: …</code> when
				<code class="font-mono text-paper">persona.allow_groups</code> is on. Replies go out as the bot
				(Telegram cannot puppet other users).
			</p>
		</div>
	</section>

	<section id="self-host" class="mx-auto max-w-6xl px-5 py-16">
		<div class="grid gap-10 lg:grid-cols-2">
			<div>
				<h2 class="text-2xl font-semibold">Run your own instance</h2>
				<p class="mt-3 leading-relaxed text-mist">
					Open source (MIT). Runtime is a single Go binary plus
					<code class="font-mono text-paper">deltachat-rpc-server</code> on
					<code class="font-mono text-paper">PATH</code>. Pairing lives in SQLite under
					<code class="font-mono text-paper">./data</code>. Secrets stay in
					<code class="font-mono text-paper">.env</code> — never commit tokens.
				</p>
				<ul class="mt-5 list-disc space-y-2 pl-5 text-sm text-mist">
					<li>Go 1.22+ and a Telegram BotFather token</li>
					<li>Chatmail / Delta Chat provider (e.g. nine.testrun.org for tests)</li>
					<li>Optional SOCKS5 / HTTP proxies for Telegram or Delta Chat</li>
					<li>Makefile is the supported build and serve entrypoint</li>
				</ul>
			</div>
			<pre
				class="overflow-x-auto rounded-2xl border border-line bg-panel p-5 font-mono text-sm leading-7 text-paper"><code
					>make config
# .env → TELEGRAM_BOT_TOKEN=…
make init QR=dcaccount:nine.testrun.org
make serve

# this site
make run-landing</code
				></pre>
		</div>
	</section>
</main>

<footer class="border-t border-line/80">
	<div
		class="mx-auto flex max-w-6xl flex-wrap items-center justify-between gap-4 px-5 py-8 text-sm text-mist"
	>
		<p>TGPORTAL · MIT · themadorg</p>
		<div class="flex gap-4">
			<a class="hover:text-paper" href={repoUrl}>Source</a>
			<a
				class="hover:text-paper"
				href="https://github.com/themadorg/tgportal/blob/main/docs/privacy.md">Privacy</a
			>
			<a class="hover:text-paper" href={botUrl}>Telegram</a>
		</div>
	</div>
</footer>
