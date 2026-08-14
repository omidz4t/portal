<script lang="ts">
	import DocLayout from '$lib/DocLayout.svelte';
	import { repoUrl } from '$lib/links';
	import type { DocsCopy, Locale } from '$lib/content';
	import { docsPath } from '$lib/content';

	let { docs, locale }: { docs: DocsCopy; locale: Locale } = $props();

	const releasesUrl = `${repoUrl}/releases`;
	const latestUrl = `${repoUrl}/releases/latest`;
</script>

<svelte:head>
	<title>{docs.selfTitle} — Portal</title>
</svelte:head>

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
		<pre><code
				>git clone https://github.com/themadorg/tgportal.git
cd tgportal
make config
# edit .env: TELEGRAM_BOT_TOKEN=…
# optional persona: PERSONA_ACCOUNT_QR=dcaccount:…
# edit config.yml: mode: personal | persona | both

make build
./tgportal --version

# first account, then run
make init QR=dcaccount:nine.testrun.org
make serve
# same as: ./tgportal --config config.yml serve</code
			></pre>
		<p>{docs.selfWay1Release}</p>
		<pre><code
				>make build-release
./dist/tgportal --version
./dist/tgportal --config config.yml serve</code
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
		<pre><code
				># example: Linux x86_64
curl -fsSL -o tgportal.tgz \
  https://github.com/themadorg/tgportal/releases/latest/download/tgportal_VERSION_linux_amd64.tar.gz
tar -tzf tgportal.tgz
tar -xzf tgportal.tgz
chmod +x tgportal
sha256sum tgportal

./tgportal --version
./tgportal --config config.yml init dcaccount:nine.testrun.org
./tgportal --config config.yml serve</code
			></pre>
		<p>{docs.selfWay2Note}</p>
		<p>{docs.selfPreview}</p>
		<p>
			{docs.selfMore}
			<a href="{repoUrl}/blob/main/docs/installation.md">installation</a>
			·
			<a href="{repoUrl}/blob/main/docs/configuration.md">configuration</a>
			·
			<a href="{repoUrl}/blob/main/docs/security.md">security</a>
			·
			<a href={docsPath(locale, 'trust')}>{docs.pages[0].title}</a>
		</p>
	</DocLayout>
</div>
