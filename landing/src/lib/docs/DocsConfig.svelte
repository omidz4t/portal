<script lang="ts">
	import DocLayout from '$lib/DocLayout.svelte';
	import { repoUrl } from '$lib/links';
	import type { DocsCopy, Locale } from '$lib/content';
	import { docsPath } from '$lib/content';
	import fullConfig from './full-config.yml?raw';

	let { docs, locale }: { docs: DocsCopy; locale: Locale } = $props();
</script>

<div class="mx-auto max-w-5xl px-5 py-16">
	<DocLayout title={docs.cfgTitle} lede={docs.cfgLede} {docs} {locale} wide>
		<p>{docs.cfgFilesLead}</p>
		<ul>
			{#each docs.cfgFilesItems as item}
				<li>{item}</li>
			{/each}
		</ul>
		<p>{docs.cfgMake}</p>
		<pre dir="ltr"><code
				>make config
# or:
cp config.example.yml config.yml
cp .env.example .env</code
			></pre>

		<h2>{docs.cfgYamlTitle}</h2>
		<p>{docs.cfgYamlLead}</p>
		<pre class="config-yml" dir="ltr"><code>{fullConfig.trimEnd()}</code></pre>

		<h2>{docs.cfgEnvTitle}</h2>
		<p>{docs.cfgEnvLead}</p>
		<ul>
			{#each docs.cfgEnvItems as item}
				<li>{item}</li>
			{/each}
		</ul>
		<pre dir="ltr"><code
				>TELEGRAM_BOT_TOKEN=123456:ABC-DEF...
# PERSONA_ACCOUNT_QR=dcaccount:nine.testrun.org
# INVITE_URL=https://i.delta.chat/#...
# TGPORTAL_DB_KEY=
# TELEGRAM_ALLOWED_USER_IDS=123456789
# PROXY_URL=socks5://127.0.0.1:1080
# PROXY_ENABLED=true
# TELEGRAM_PROXY_URL=socks5://127.0.0.1:1080
# DELTACHAT_PROXY_URL=socks5://127.0.0.1:1080</code
			></pre>

		<h2>{docs.cfgPrecTitle}</h2>
		<p>{docs.cfgPrecFolder}</p>
		<p>{docs.cfgPrecAllow}</p>
		<p>{docs.cfgPrecPath}</p>

		<h2>{docs.cfgAssetsTitle}</h2>
		<ul>
			{#each docs.cfgAssetsItems as item}
				<li>{item}</li>
			{/each}
		</ul>
		<p>
			{docs.cfgRepo}
			<a href="{repoUrl}/blob/main/docs/configuration.md">docs/configuration.md</a>
			·
			<a href="{repoUrl}/blob/main/config.example.yml">config.example.yml</a>
			·
			<a href={docsPath(locale, 'self-host')}>{docs.trustLinkSelf}</a>
			·
			<a href={docsPath(locale, 'persona')}>{docs.pages[1].title}</a>
		</p>
	</DocLayout>
</div>
