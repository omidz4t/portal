<script lang="ts">
	import DocLayout from '$lib/DocLayout.svelte';
	import { repoUrl } from '$lib/links';

	const releasesUrl = `${repoUrl}/releases`;
	const latestUrl = `${repoUrl}/releases/latest`;
</script>

<svelte:head>
	<title>Self-host — TGPORTAL</title>
</svelte:head>

<div class="mx-auto max-w-3xl px-5 py-16">
	<DocLayout
		title="Self-host"
		lede="Recommended. You keep the token, the database, and the ghost keys. Pick one path: build the binary, or download a release."
	>
		<p>
			Both ways need the same runtime pieces after you have a <code>tgportal</code> binary:
		</p>
		<ul>
			<li>
				<a href="https://github.com/chatmail/core/tree/main/deltachat-rpc-server"
					>deltachat-rpc-server</a
				>
				on <code>PATH</code> (match the project’s rpc-client major; this repo uses v2.56.x APIs)
			</li>
			<li>A BotFather token in <code>.env</code></li>
			<li>
				A chatmail / Delta Chat provider for <code>make init</code> / <code>tgportal init</code>
			</li>
		</ul>
		<p>
			Then copy <code>config.example.yml</code> → <code>config.yml</code> and
			<code>.env.example</code> → <code>.env</code>. Data default is <code>./data</code>. Never
			commit <code>.env</code>, <code>config.yml</code>, or <code>data/</code>.
		</p>

		<h2>Way 1 — build the binary</h2>
		<p>
			Use this when you have Go 1.22+ and want to compile from git. The Makefile is the supported
			entrypoint.
		</p>
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
		<p>
			All-in-one static binary (stripped, version stamped) goes to <code>dist/tgportal</code>:
		</p>
		<pre><code
				>make build-release
./dist/tgportal --version
./dist/tgportal --config config.yml serve</code
			></pre>

		<h2>Way 2 — download the binary from a release</h2>
		<p>
			Use this when you do not want a Go toolchain. GitHub Actions publish archives on each tag
			<code>v*</code>:
			<code>tgportal_&lt;tag&gt;_&lt;os&gt;_&lt;arch&gt;.tar.gz</code> (Windows
			<code>.zip</code>) plus <code>checksums.txt</code>.
		</p>
		<ol>
			<li>
				Open <a href={latestUrl}>latest release</a> or the
				<a href={releasesUrl}>releases list</a>.
			</li>
			<li>
				Download the archive for your OS/CPU (examples:
				<code>linux_amd64</code>, <code>linux_arm64</code>, <code>darwin_arm64</code>,
				<code>windows_amd64</code>).
			</li>
			<li>Verify the SHA-256 against <code>checksums.txt</code>.</li>
			<li>
				Unpack, mark executable, put it next to <code>config.yml</code> and <code>.env</code>.
			</li>
		</ol>
		<pre><code
				># example: Linux x86_64
curl -fsSL -o tgportal.tgz \
  https://github.com/themadorg/tgportal/releases/latest/download/tgportal_VERSION_linux_amd64.tar.gz
tar -tzf tgportal.tgz   # see the filename
tar -xzf tgportal.tgz
chmod +x tgportal
sha256sum tgportal      # compare to checksums.txt

# config (from the repo examples if the archive did not include them)
#   config.example.yml → config.yml
#   .env.example → .env

./tgportal --version
./tgportal --config config.yml init dcaccount:nine.testrun.org
./tgportal --config config.yml serve</code
			></pre>
		<p>
			Replace <code>VERSION</code> with the tag without the leading <code>v</code> if the asset name
			uses that form (see the release notes). You still need
			<code>deltachat-rpc-server</code> on <code>PATH</code>; the release is only the Go bot.
		</p>

		<p>
			Landing preview while developing: <code>make run-landing</code> →
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
