<script lang="ts">
	import DocLayout from '$lib/DocLayout.svelte';
	import { repoUrl } from '$lib/links';
	import type { DocsCopy, Locale } from '$lib/content';
	import { docsPath } from '$lib/content';

	let { docs, locale }: { docs: DocsCopy; locale: Locale } = $props();
</script>

<div class="mx-auto max-w-3xl px-5 py-16">
	<DocLayout title={docs.cfgTitle} lede={docs.cfgLede} {docs} {locale}>
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
# PROXY_URL=socks5://127.0.0.1:1080</code
			></pre>

		<h2>{docs.cfgYamlTitle}</h2>
		<h3>{docs.cfgProfileTitle}</h3>
		<ul>
			{#each docs.cfgProfileItems as item}
				<li>{item}</li>
			{/each}
		</ul>

		<h3>{docs.cfgModeTitle}</h3>
		<p>
			{docs.cfgModeLead}
			<a href={docsPath(locale, 'persona')}>{docs.pages[1].title}</a>
		</p>
		<ul>
			{#each docs.cfgPersonaItems as item}
				<li>{item}</li>
			{/each}
		</ul>

		<h3>{docs.cfgPairTitle}</h3>
		<ul>
			{#each docs.cfgPairItems as item}
				<li>{item}</li>
			{/each}
		</ul>

		<h3>{docs.cfgStoreTitle}</h3>
		<ul>
			{#each docs.cfgStoreItems as item}
				<li>{item}</li>
			{/each}
		</ul>

		<h3>{docs.cfgLogTitle}</h3>
		<p>{docs.cfgLogLead}</p>
		<pre dir="ltr"><code
				>log: false
# log: stderr
# log: ./data/tgportal.log
log_level: info</code
			></pre>

		<h3>{docs.cfgProxyTitle}</h3>
		<p>{docs.cfgProxyLead}</p>
		<pre dir="ltr"><code
				>proxy:
  enabled: false
  url: ""   # socks5://127.0.0.1:1080

telegram:
  proxy:
    enabled: false
    url: ""

deltachat:
  proxy:
    enabled: false
    url: ""</code
			></pre>

		<h3>{docs.cfgTgTitle}</h3>
		<ul>
			{#each docs.cfgTgItems as item}
				<li>{item}</li>
			{/each}
		</ul>

		<h3>{docs.cfgBridgeTitle}</h3>
		<p>{docs.cfgBridgeLead}</p>
		<ul>
			{#each docs.cfgBridgeItems as item}
				<li>{item}</li>
			{/each}
		</ul>
		<h3>{docs.cfgLimitsTitle}</h3>
		<p>{docs.cfgLimitsLead}</p>
		<ul>
			{#each docs.cfgLimitsItems as item}
				<li>{item}</li>
			{/each}
		</ul>

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
		</p>
	</DocLayout>
</div>
