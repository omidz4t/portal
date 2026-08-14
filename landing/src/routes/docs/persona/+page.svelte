<script lang="ts">
	import DocLayout from '$lib/DocLayout.svelte';
	import { repoUrl } from '$lib/links';
</script>

<svelte:head>
	<title>Persona mode — TGPORTAL</title>
</svelte:head>

<div class="mx-auto max-w-3xl px-5 py-16">
	<DocLayout
		title="Persona mode"
		lede="Register your own Telegram bot. Each remote user becomes a stable ghost Delta Chat account that talks to you as themselves."
	>
		<h2>Why it exists</h2>
		<p>
			<strong>Personal mode</strong> is one portal bot and one pair:
			<code>telegram_user_id ↔ dc_chat_id</code>. Fine for “my stickers land in my DC chat.”
		</p>
		<p>
			<strong>Persona mode</strong> is for people who already have (or want) a Telegram bot of their
			own. Friends and groups message <em>your</em> bot. On Delta Chat you should see
			<em>them</em>, not a labeled dump from a shared bridge account.
		</p>

		<h2>How ghosts work</h2>
		<p>
			TGPORTAL creates a real Delta Chat account per remote Telegram user (under each persona bot).
			Display name and avatar are copied from Telegram. Lookup key is
			<code>(persona_bot_id, telegram_user_id)</code>: missing → create; exists → reuse forever.
		</p>
		<p>
			That ghost delivers the person’s message to <strong>you</strong> as a normal 1:1 — no prefix, no
			“via TGPORTAL.” When you reply in that chat, the persona poller sends the reply out through your
			BotFather bot.
		</p>
		<p>
			Ghost tables are separate from personal-mode pairs. Alice pairing the public portal bot does
			not reuse Alice’s persona ghost.
		</p>

		<h2>Owner setup</h2>
		<ol>
			<li>
				Set <code>mode: persona</code> or <code>both</code>, and
				<code>PERSONA_ACCOUNT_QR</code> (a <code>dcaccount:</code> / <code>dclogin:</code> URI used to
				provision ghosts).
			</li>
			<li>
				<code>/pair</code> on the <strong>portal</strong> bot and finish on Delta Chat. This stores your
				DC vcard / public key so ghosts can message you without a second invite dance.
			</li>
			<li>Create a bot with BotFather. Copy the token. Never paste it into issues or chats.</li>
			<li>
				In a <strong>private</strong> chat with the portal bot:
				<pre><code>/pair-bot &lt;TOKEN&gt; [https://t.me/YourBot]</code></pre>
			</li>
		</ol>
		<p>
			Limits: <code>persona.max_ghosts</code>, <code>persona.max_bots</code>. List with
			<code>/bots</code>. Stop with <code>/unpair-bot [id|@user]</code>. Re-run
			<code>/pair-bot</code> after unpair to start the same bot again.
		</p>

		<h2>Groups</h2>
		<p>Off by default. Enable <code>persona.allow_groups: true</code>, then:</p>
		<ol>
			<li>Add the persona bot to the Telegram group.</li>
			<li>
				BotFather → your bot → Bot Settings → Group Privacy → <strong>Turn off</strong>. Otherwise
				Telegram only delivers mentions and commands.
			</li>
			<li>
				Posts become a Delta Chat group named <code>TG: …</code>. Each speaker is their ghost; you
				are a member.
			</li>
			<li>Your replies in that DC group go to Telegram as the bot, not as the original users.</li>
		</ol>

		<h2>Trust</h2>
		<p>
			The instance that stores the BotFather token <em>is</em> your Telegram bot. Ghost account keys
			live under <code>./data</code>. If you give <code>/pair-bot</code> to someone else’s public
			TGPORTAL, you are handing them the bot and every conversation it will carry. Prefer
			<a href="/docs/self-host/">self-hosting</a>. See <a href="/docs/trust/">trust</a>.
		</p>
		<p>
			Repo operator notes:
			<a href="{repoUrl}/blob/main/docs/persona.md">docs/persona.md</a>, design:
			<a href="{repoUrl}/blob/main/docs/persona-design.md">persona-design.md</a>.
		</p>
	</DocLayout>
</div>
