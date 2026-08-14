<script lang="ts">
	import DocLayout from '$lib/DocLayout.svelte';
	import { repoUrl } from '$lib/links';
	import type { DocsCopy, Locale } from '$lib/content';
	import { docsPath } from '$lib/content';

	let { docs, locale }: { docs: DocsCopy; locale: Locale } = $props();

	const releasesUrl = `${repoUrl}/releases`;
	const latestUrl = `${repoUrl}/releases/latest`;
</script>

<div class="mx-auto max-w-3xl px-5 py-16">
	<DocLayout title={docs.selfTitle} lede={docs.selfLede} {docs} {locale}>
		<p>{docs.selfNeedLead}</p>
		<ul>
			{#each docs.selfNeedItems as item}
				<li>{item}</li>
			{/each}
		</ul>
		<p>{docs.selfConfig}</p>

		<h2>{docs.selfWay1}</h2>
		<p>{docs.selfWay1P}</p>
		<pre dir="ltr"><code
				>git clone https://github.com/omidz4t/portal.git
cd portal
make config
# edit .env: TELEGRAM_BOT_TOKEN=…
# optional persona: PERSONA_ACCOUNT_QR=dcaccount:…
# edit config.yml: mode: personal | persona | both

make build
./portal --version

# first account, then run
make init QR=dcaccount:nine.testrun.org
make serve
# same as: ./portal --config config.yml serve</code
			></pre>
		<p>{docs.selfWay1Release}</p>
		<pre dir="ltr"><code
				>make build-release
./dist/portal --version
./dist/portal --config config.yml serve</code
			></pre>

		<h2>{docs.selfWay2}</h2>
		<p>{docs.selfWay2P}</p>
		<ol>
			<li>
				<a href={latestUrl}>{docs.selfLatest}</a>
				·
				<a href={releasesUrl}>{docs.selfList}</a>
			</li>
			{#each docs.selfWay2Items.slice(1) as item}
				<li>{item}</li>
			{/each}
		</ol>
		<pre dir="ltr"><code
				># example: Linux x86_64
curl -fsSL -o portal.tgz \
  https://github.com/omidz4t/portal/releases/latest/download/portal_VERSION_linux_amd64.tar.gz
tar -tzf portal.tgz
tar -xzf portal.tgz
chmod +x portal
sha256sum portal

./portal --version
./portal --config config.yml init dcaccount:nine.testrun.org
./portal --config config.yml serve</code
			></pre>
		<p>{docs.selfWay2Note}</p>
		<p>{docs.selfPreview}</p>
		<p>
			{docs.selfMore}
			<a href="{repoUrl}/blob/main/docs/installation.md">installation</a>
			·
			<a href={docsPath(locale, 'config')}>{docs.selfLinkConfig}</a>
			·
			<a href="{repoUrl}/blob/main/docs/security.md">security</a>
			·
			<a href={docsPath(locale, 'trust')}>{docs.pages[0].title}</a>
		</p>
	</DocLayout>
</div>
