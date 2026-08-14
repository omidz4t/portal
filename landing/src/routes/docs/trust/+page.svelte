<script lang="ts">
	import DocLayout from '$lib/DocLayout.svelte';
</script>

<svelte:head>
	<title>Trust the operator — TGPORTAL</title>
</svelte:head>

<div class="mx-auto max-w-3xl px-5 py-16">
	<DocLayout
		title="You are trusting whoever runs the bot"
		lede="TGPORTAL is not an end-to-end tunnel past the operator. It is better to host one yourself."
	>
		<p>
			Messages you send a Telegram bot are delivered to <strong>that bot’s process</strong>.
			TGPORTAL then downloads files and calls Delta Chat RPC on the same machine. The person (or
			company) who can read that disk and those logs can read your bridged content.
		</p>
		<p>That includes:</p>
		<ul>
			<li>Telegram user ids, names, and message bodies sent to the bot</li>
			<li>Pairing codes and the SQLite map of Telegram ↔ Delta Chat chats</li>
			<li>Sticker, GIF, image, and video files while they sit in cache</li>
			<li>
				In persona mode: BotFather tokens you register, and the keys of ghost Delta Chat accounts
			</li>
		</ul>
		<p>
			Delta Chat still encrypts traffic <em>on the wire to chatmail</em>. That does not hide
			plaintext from <code>tgportal</code> itself. The runner is a participant.
		</p>
		<p>
			A public instance such as <a href="https://t.me/tgdeltabridgebot">@tgdeltabridgebot</a> is convenient
			for trying the product. Using it means you accept that operator’s machine. Do not bridge anything
			you would not hand to them.
		</p>
		<p>
			<strong>Host your own.</strong>
			Then the only runner is you. See <a href="/docs/self-host/">self-host</a>. Operators who serve
			other people should publish a privacy policy (repo
			<code>docs/privacy.md</code>) and never log tokens.
		</p>
	</DocLayout>
</div>
