<script lang="ts">
	import DocLayout from '$lib/DocLayout.svelte';
	import { botUrl, repoUrl } from '$lib/links';
</script>

<svelte:head>
	<title>Pairing — TGPORTAL</title>
</svelte:head>

<div class="mx-auto max-w-3xl px-5 py-16">
	<DocLayout
		title="Personal pairing"
		lede="Nothing is forwarded until a Telegram user is linked to a Delta Chat chat."
	>
		<p>
			This is <code>mode: personal</code> (also available under <code>both</code>). For owner-owned
			bots see <a href="/docs/persona/">persona</a>.
		</p>
		<p>
			Using a hosted bot means you
			<a href="/docs/trust/">trust that runner</a>. Prefer
			<a href="/docs/self-host/">your own instance</a>.
		</p>

		<h2>Telegram first</h2>
		<ol>
			<li>
				Open the bot in a private 1:1 (e.g. <a href={botUrl}>@tgdeltabridgebot</a>) and send
				<code>/start</code> or <code>/pair</code>. Groups are rejected so codes do not leak.
			</li>
			<li>You get a QR of the Delta Chat invite plus an 8-character code (about 30 minutes).</li>
			<li>Open the invite in Delta Chat / ArcaneChat and accept the chat.</li>
			<li>Send <strong>only the code</strong> (case-insensitive).</li>
		</ol>

		<h2>Delta Chat first</h2>
		<ol>
			<li>Message the Delta Chat bot while unpaired (or send <code>/pair</code>).</li>
			<li>It replies with <code>https://t.me/&lt;bot&gt;?start=CODE</code>.</li>
			<li>Open that link on Telegram. Pairing is active immediately.</li>
		</ol>

		<h2>Useful commands (Telegram)</h2>
		<ul>
			<li><code>/status</code> — pairing state</li>
			<li><code>/disconnect</code> — unlink</li>
			<li>
				<code>/send_pack</code> — reply to a sticker to bridge the whole pack (if
				<code>bridge.sticker_packs</code> is on)
			</li>
		</ul>
		<p>
			Full walkthrough in the repo:
			<a href="{repoUrl}/blob/main/docs/pairing.md">docs/pairing.md</a>.
		</p>
	</DocLayout>
</div>
