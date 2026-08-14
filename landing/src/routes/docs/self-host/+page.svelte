<script lang="ts">
	import DocLayout from '$lib/DocLayout.svelte';
	import { repoUrl } from '$lib/links';
</script>

<svelte:head>
	<title>Self-host — TGPORTAL</title>
</svelte:head>

<div class="mx-auto max-w-3xl px-5 py-16">
	<DocLayout
		title="Self-host"
		lede="Recommended. You keep the token, the database, and the ghost keys."
	>
		<p>
			Requirements: Go 1.22+,
			<a href="https://github.com/chatmail/core/tree/main/deltachat-rpc-server"
				>deltachat-rpc-server</a
			>
			on <code>PATH</code> (match the project’s rpc-client major version), a BotFather token, and a chatmail
			/ Delta Chat provider.
		</p>
		<pre><code
				>git clone https://github.com/themadorg/tgportal.git
cd tgportal
make config
# edit .env: TELEGRAM_BOT_TOKEN=…
# optional persona: PERSONA_ACCOUNT_QR=dcaccount:…
# edit config.yml: mode: personal | persona | both
make init QR=dcaccount:nine.testrun.org
make serve</code
			></pre>
		<p>
			Data directory default is <code>./data</code> (pairing DB, Delta Chat accounts, TG cache).
			Never commit <code>.env</code>, <code>config.yml</code>, or <code>data/</code>.
		</p>
		<p>
			Landing site for local preview: <code>make run-landing</code> →
			<a href="http://127.0.0.1:5173">http://127.0.0.1:5173</a>.
		</p>
		<p>
			More: <a href="{repoUrl}/blob/main/docs/installation.md">installation</a>,
			<a href="{repoUrl}/blob/main/docs/configuration.md">configuration</a>,
			<a href="{repoUrl}/blob/main/docs/security.md">security</a>,
			<a href="/docs/trust/">trust</a>.
		</p>
	</DocLayout>
</div>
