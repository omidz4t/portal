<script lang="ts">
	import { botUrl, repoUrl } from '$lib/links';
	import KoboyoIcon from '$lib/KoboyoIcon.svelte';

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
			<KoboyoIcon name="warning" class="mt-0.5 h-6 w-6" />
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
			<h2 class="mt-2 text-2xl font-semibold text-balance sm:text-3xl">
				Persona: they message your Telegram bot. You see them as people on Delta Chat.
			</h2>
			<p class="mt-4 max-w-3xl text-lg leading-relaxed text-mist">
				Personal mode is one public portal bot and one chat. Persona is different: you plug in a bot <em
					>you</em
				>
				created with BotFather. Friends never join a shared dump. Each Telegram user becomes a
				<strong class="text-paper">ghost</strong> — a real Delta Chat account with their name and photo
				— that opens a normal 1:1 with you. Replies you type there leave as your bot.
			</p>

			<div class="mt-10 grid gap-4 md:grid-cols-2">
				<article class="rounded-2xl border border-line bg-ink/40 p-6">
					<div class="flex items-center gap-3 text-mist">
						<KoboyoIcon name="message" class="h-10 w-10" />
						<h3 class="text-lg font-semibold text-paper">Personal pairing</h3>
					</div>
					<p class="mt-3 text-sm leading-relaxed text-mist">
						You talk to the portal bot. Stickers and text land in <em>one</em> Delta Chat conversation
						with that bot. Fine for “my media, my chat.”
					</p>
				</article>
				<article class="rounded-2xl border border-sky/40 bg-ink/60 p-6">
					<div class="flex items-center gap-3 text-sky">
						<KoboyoIcon name="ghost" class="h-10 w-10" />
						<h3 class="text-lg font-semibold text-paper">Persona</h3>
					</div>
					<p class="mt-3 text-sm leading-relaxed text-mist">
						They talk to <em>your</em> bot. On Delta Chat, Alice looks like Alice, Bob looks like Bob
						— separate chats, reused forever for the same Telegram id.
					</p>
				</article>
			</div>

			<h3 class="mt-14 text-xl font-semibold">How a message travels</h3>
			<p class="mt-2 max-w-2xl text-mist">
				Read left to right on a wide screen, or top to bottom on a phone. Same story.
			</p>
			<ol class="mt-8 grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
				<li class="rounded-2xl border border-line bg-ink/50 p-5">
					<KoboyoIcon name="telegram" class="h-14 w-14 text-sky" />
					<p class="mt-4 font-mono text-xs text-sky">1 · they write</p>
					<h4 class="mt-1 text-lg font-semibold">Alice DMs your bot</h4>
					<p class="mt-2 text-sm leading-relaxed text-mist">
						She never opens Delta Chat. She only knows the Telegram bot you published.
					</p>
				</li>
				<li class="rounded-2xl border border-line bg-ink/50 p-5">
					<KoboyoIcon name="bot" class="h-14 w-14 text-teal" />
					<p class="mt-4 font-mono text-xs text-sky">2 · your bot</p>
					<h4 class="mt-1 text-lg font-semibold">TGPORTAL is polling it</h4>
					<p class="mt-2 text-sm leading-relaxed text-mist">
						You registered the BotFather token with
						<code class="font-mono text-paper">/pair-bot</code> on the portal (after you
						<code class="font-mono text-paper">/pair</code>ed yourself).
					</p>
				</li>
				<li class="rounded-2xl border border-line bg-ink/50 p-5">
					<KoboyoIcon name="ghost" class="h-14 w-14 text-paper" />
					<p class="mt-4 font-mono text-xs text-sky">3 · a ghost is born</p>
					<h4 class="mt-1 text-lg font-semibold">One DC account per Telegram id</h4>
					<p class="mt-2 text-sm leading-relaxed text-mist">
						Missing? Create with <code class="font-mono text-paper">PERSONA_ACCOUNT_QR</code>, copy
						name and photo. Exists? Reuse. Not the same table as personal pairs.
					</p>
				</li>
				<li class="rounded-2xl border border-line bg-ink/50 p-5">
					<KoboyoIcon name="paper-plane" class="h-14 w-14 text-teal" />
					<p class="mt-4 font-mono text-xs text-sky">4 · you read and reply</p>
					<h4 class="mt-1 text-lg font-semibold">Normal 1:1 on Delta Chat</h4>
					<p class="mt-2 text-sm leading-relaxed text-mist">
						No “via Telegram” label. Your reply goes back out through the bot. Telegram cannot
						puppet Alice as Alice.
					</p>
				</li>
			</ol>

			<div class="mt-10 grid gap-5 md:grid-cols-3">
				<article class="rounded-2xl border border-line bg-ink/50 p-6">
					<KoboyoIcon name="key" class="h-10 w-10 text-sky" />
					<h3 class="mt-4 text-lg font-semibold">You set it up once</h3>
					<p class="mt-2 text-sm leading-relaxed text-mist">
						<code class="font-mono text-paper">mode: persona</code> or
						<code class="font-mono text-paper">both</code>. Pair the portal so TGPORTAL has your DC
						public key. Then
						<code class="font-mono text-paper">/pair-bot &lt;TOKEN&gt;</code> in a private chat.
						Token stays in SQLite under <code class="font-mono text-paper">./data</code>.
					</p>
				</article>
				<article class="rounded-2xl border border-line bg-ink/50 p-6">
					<KoboyoIcon name="users" class="h-10 w-10 text-sky" />
					<h3 class="mt-4 text-lg font-semibold">Groups are optional</h3>
					<p class="mt-2 text-sm leading-relaxed text-mist">
						Turn on <code class="font-mono text-paper">persona.allow_groups</code>, add the bot,
						disable BotFather group privacy. Posts become a Delta Chat group
						<code class="font-mono text-paper">TG: …</code> with each speaker as their ghost. You are
						a member; your replies are the bot.
					</p>
				</article>
				<article class="rounded-2xl border border-line bg-ink/50 p-6">
					<KoboyoIcon name="lock" class="h-10 w-10 text-warn" />
					<h3 class="mt-4 text-lg font-semibold">The host is the bot</h3>
					<p class="mt-2 text-sm leading-relaxed text-mist">
						Whoever runs <code class="font-mono text-paper">tgportal</code> holds the token and the
						ghost keys. Do not <code class="font-mono text-paper">/pair-bot</code> on a public
						instance you do not operate.
						<a class="text-teal underline-offset-2 hover:underline" href="/docs/self-host/"
							>Host your own</a
						>.
					</p>
				</article>
			</div>

			<p class="mt-8 text-sm text-mist">
				Commands: <code class="font-mono text-paper">/bots</code>,
				<code class="font-mono text-paper">/unpair-bot</code>. Full notes:
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
