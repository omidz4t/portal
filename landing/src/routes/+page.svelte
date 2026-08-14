<script lang="ts">
	import { botUrl, repoUrl } from '$lib/links';

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
</script>

<div>
	<aside class="border-b border-warn/40 bg-warn-bg" role="note" aria-label="Trust warning">
		<div class="mx-auto flex max-w-6xl gap-3 px-5 py-4 text-sm leading-relaxed text-warn">
			<svg
				class="mt-0.5 h-5 w-5 shrink-0"
				viewBox="0 0 24 24"
				fill="none"
				stroke="currentColor"
				stroke-width="2"
				aria-hidden="true"
			>
				<path
					stroke-linecap="round"
					stroke-linejoin="round"
					d="M12 9v4m0 4h.01M10.29 3.86 1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0Z"
				/>
			</svg>
			<p>
				<strong class="font-semibold text-paper">Warning.</strong>
				A public bot is a relay. Whoever runs
				<code class="font-mono text-paper">tgportal</code>
				can see pairing data and everything you bridge. Delta Chat encryption does not hide messages from
				that host. Prefer
				<a class="font-semibold underline underline-offset-2" href="/docs/self-host/"
					>self-hosting</a
				>
				so you are the only operator.
				<a class="underline underline-offset-2" href="/docs/trust/">Why this matters</a>.
			</p>
		</div>
	</aside>

	<section class="mx-auto grid max-w-6xl items-center gap-12 px-5 py-16 lg:grid-cols-2 lg:py-24">
		<div>
			<p class="font-mono text-xs tracking-[0.22em] text-teal uppercase">Telegram ↔ Delta Chat</p>
			<h1 class="mt-4 text-4xl leading-tight font-semibold tracking-tight text-balance sm:text-5xl">
				Stickers and GIFs that survive the jump to Delta Chat.
			</h1>
			<p class="mt-5 max-w-xl text-lg leading-relaxed text-mist">
				TGPORTAL is a Go bridge. Pair once with a short code. After that, media you send the
				Telegram bot lands in your Delta Chat conversation — and the other way around — without
				noisy filenames or “from Telegram” captions.
			</p>
			<div class="mt-8 flex flex-wrap gap-3">
				<a
					href="/docs/self-host/"
					class="inline-flex min-h-11 items-center rounded-full bg-teal px-5 text-sm font-semibold text-ink hover:brightness-110"
				>
					Host your own
				</a>
				<a
					href={botUrl}
					rel="noreferrer"
					class="inline-flex min-h-11 items-center rounded-full border border-line px-5 text-sm font-medium text-paper hover:border-mist"
				>
					Public bot (trust the runner)
				</a>
				<a
					href="/docs/"
					class="inline-flex min-h-11 items-center rounded-full border border-line px-5 text-sm font-medium text-paper hover:border-mist"
				>
					Docs
				</a>
			</div>
		</div>
		<figure class="overflow-hidden rounded-2xl border border-line bg-panel shadow-2xl">
			<img
				src="/poster.jpg"
				alt="TGPORTAL branding poster: dark void with the product mark"
				width="960"
				height="640"
				class="aspect-3/2 h-full w-full object-cover"
			/>
		</figure>
	</section>

	<section class="border-y border-line/80 bg-panel/40">
		<div class="mx-auto max-w-6xl px-5 py-16">
			<h2 class="text-2xl font-semibold">What it bridges</h2>
			<p class="mt-2 max-w-2xl text-mist">
				Bidirectional, per-user pairing. Telegram types stay useful on Delta Chat.
			</p>
			<div class="mt-8 overflow-x-auto rounded-xl border border-line">
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
		<h2 class="text-2xl font-semibold">Pair once (personal mode)</h2>
		<p class="mt-2 max-w-2xl text-mist">
			Default product: one Telegram user maps to one Delta Chat chat on the portal bot. Live
			instance:
			<a class="text-teal underline-offset-2 hover:underline" href={botUrl}>@tgdeltabridgebot</a>
			— only if you accept
			<a class="text-teal underline-offset-2 hover:underline" href="/docs/trust/"
				>trusting that host</a
			>. You can also start from Delta Chat and open
			<code class="font-mono text-paper">t.me/…?start=CODE</code>.
			<a class="text-teal underline-offset-2 hover:underline" href="/docs/pairing/">Pairing docs</a
			>.
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

	<section id="persona" class="border-y border-line/80 bg-panel/40">
		<div class="mx-auto max-w-6xl px-5 py-16">
			<p class="font-mono text-xs tracking-[0.2em] text-sky uppercase">Persona</p>
			<h2 class="mt-2 text-2xl font-semibold">Your Telegram bot, their faces on Delta Chat</h2>
			<p class="mt-3 max-w-3xl leading-relaxed text-mist">
				Persona is a second product in the same binary. You keep a normal portal pair (so TGPORTAL
				knows <em>your</em> Delta Chat identity), then register a BotFather bot you own. People who
				message <em>that</em> bot are not dumped into one shared chat. Each Telegram user gets a
				<strong class="text-paper">ghost Delta Chat account</strong> — name and photo copied from Telegram
				— that writes to you as a normal 1:1. No “from Telegram” prefix. Replies you send in that 1:1
				go back out through your bot.
			</p>
			<div class="mt-8 grid gap-5 md:grid-cols-3">
				<article class="rounded-2xl border border-line bg-ink/50 p-6">
					<p class="font-mono text-xs text-sky">01 · owner</p>
					<h3 class="mt-2 text-lg font-semibold">Pair, then /pair-bot</h3>
					<p class="mt-2 text-sm leading-relaxed text-mist">
						<code class="font-mono text-paper">/pair</code> on the portal bot first (stores your DC
						vcard/public key). Create a bot with BotFather. In a private chat with the portal:
						<code class="font-mono text-paper">/pair-bot &lt;TOKEN&gt;</code>. Tokens stay in SQLite
						under <code class="font-mono text-paper">./data</code> — never in logs or git.
					</p>
				</article>
				<article class="rounded-2xl border border-line bg-ink/50 p-6">
					<p class="font-mono text-xs text-sky">02 · ghosts</p>
					<h3 class="mt-2 text-lg font-semibold">One account per Telegram id</h3>
					<p class="mt-2 text-sm leading-relaxed text-mist">
						Alice DMs your bot → ghost A messages you. Bob DMs → ghost B, separate chat. Same person
						later? Same ghost, reused forever. These accounts are created with
						<code class="font-mono text-paper">PERSONA_ACCOUNT_QR</code> and are independent of personal-mode
						pairings.
					</p>
				</article>
				<article class="rounded-2xl border border-line bg-ink/50 p-6">
					<p class="font-mono text-xs text-sky">03 · groups</p>
					<h3 class="mt-2 text-lg font-semibold">Optional TG group mirror</h3>
					<p class="mt-2 text-sm leading-relaxed text-mist">
						If <code class="font-mono text-paper">persona.allow_groups</code> is on, add the bot to
						a Telegram group and turn <strong class="text-paper">Group Privacy off</strong> in
						BotFather. Posts become a Delta Chat group
						<code class="font-mono text-paper">TG: …</code> with each speaker as their ghost. Your replies
						leave as the bot (Telegram cannot puppet users).
					</p>
				</article>
			</div>
			<p class="mt-8 text-sm text-mist">
				Config: <code class="font-mono text-paper">mode: persona</code> or
				<code class="font-mono text-paper">both</code>. Full operator steps, commands, and failure
				notes:
				<a class="text-teal underline-offset-2 hover:underline" href="/docs/persona/"
					>persona docs</a
				>.
			</p>
		</div>
	</section>

	<section id="self-host" class="mx-auto max-w-6xl px-5 py-16">
		<div class="grid gap-10 lg:grid-cols-2">
			<div>
				<h2 class="text-2xl font-semibold">Run your own instance</h2>
				<p class="mt-3 leading-relaxed text-mist">
					This is the recommended way to use TGPORTAL. Open source (MIT). One Go binary plus
					<code class="font-mono text-paper">deltachat-rpc-server</code> on
					<code class="font-mono text-paper">PATH</code>. You hold the token, the SQLite DB, and the
					ghost keys.
				</p>
				<ul class="mt-5 list-disc space-y-2 pl-5 text-sm text-mist">
					<li>Go 1.22+ and a Telegram BotFather token</li>
					<li>Chatmail / Delta Chat provider (e.g. nine.testrun.org for tests)</li>
					<li>Optional SOCKS5 / HTTP proxies</li>
					<li>Makefile is the supported entrypoint</li>
				</ul>
				<p class="mt-4 text-sm">
					<a class="text-teal underline-offset-2 hover:underline" href="/docs/self-host/"
						>Self-host walkthrough</a
					>
					·
					<a class="text-teal underline-offset-2 hover:underline" href={repoUrl}>GitHub</a>
				</p>
			</div>
			<pre
				class="overflow-x-auto rounded-2xl border border-line bg-panel p-5 font-mono text-sm leading-7 text-paper"><code
					>make config
# .env → TELEGRAM_BOT_TOKEN=…
#        PERSONA_ACCOUNT_QR=…   # if you want persona
make init QR=dcaccount:nine.testrun.org
make serve

make run-landing</code
				></pre>
		</div>
	</section>
</div>
