<script lang="ts">
	import type { Locale } from '$lib/content';
	import versions from '$lib/versions-data';

	type TargetId =
		| 'linux_amd64'
		| 'linux_arm64'
		| 'android_arm64'
		| 'darwin_amd64'
		| 'darwin_arm64'
		| 'windows_amd64';

	type Target = {
		id: TargetId;
		asset: 'linux_amd64' | 'linux_arm64' | 'darwin_amd64' | 'darwin_arm64' | 'windows_amd64';
		os: 'linux' | 'android' | 'darwin' | 'windows';
		label: { en: string; fa: string };
		hint: { en: string; fa: string };
	};

	const targets: Target[] = [
		{
			id: 'linux_amd64',
			asset: 'linux_amd64',
			os: 'linux',
			label: { en: 'Linux · x86_64', fa: 'لینوکس · x86_64' },
			hint: { en: 'Most VPS and desktop PCs', fa: 'اکثر سرورها و رایانه‌های رومیزی' }
		},
		{
			id: 'linux_arm64',
			asset: 'linux_arm64',
			os: 'linux',
			label: { en: 'Linux · ARM64', fa: 'لینوکس · ARM64' },
			hint: { en: 'Pi, ARM servers', fa: 'رزبری‌پای و سرور ARM' }
		},
		{
			id: 'android_arm64',
			asset: 'linux_arm64',
			os: 'android',
			label: { en: 'Phone · ARM64', fa: 'گوشی · ARM64' },
			hint: { en: 'Android / Termux', fa: 'اندروید / ترموکس' }
		},
		{
			id: 'darwin_arm64',
			asset: 'darwin_arm64',
			os: 'darwin',
			label: { en: 'macOS · Apple silicon', fa: 'مک · اپل سیلیکون' },
			hint: { en: 'M1 / M2 / M3 / M4', fa: 'ام۱، ام۲، ام۳، ام۴' }
		},
		{
			id: 'darwin_amd64',
			asset: 'darwin_amd64',
			os: 'darwin',
			label: { en: 'macOS · Intel', fa: 'مک · اینتل' },
			hint: { en: 'Older Macs', fa: 'مک‌های قدیمی‌تر' }
		},
		{
			id: 'windows_amd64',
			asset: 'windows_amd64',
			os: 'windows',
			label: { en: 'Windows · x86_64', fa: 'ویندوز · x86_64' },
			hint: { en: '64-bit Windows', fa: 'ویندوز ۶۴بیت' }
		}
	];

	let { locale }: { locale: Locale } = $props();

	const tag = versions.releases[0]?.tag ?? `v${versions.current}`;
	const version = tag.replace(/^v/, '');

	function guess(): TargetId {
		if (typeof navigator === 'undefined') return 'linux_amd64';
		const ua = navigator.userAgent;
		const plat = navigator.platform || '';
		const arm = /aarch64|arm64|Apple/i.test(ua) || /arm/i.test(plat);
		if (/Win/i.test(plat) || /Windows/i.test(ua)) return 'windows_amd64';
		if (/Mac/i.test(plat) || /Macintosh/i.test(ua)) return arm ? 'darwin_arm64' : 'darwin_amd64';
		if (/Android/i.test(ua)) return 'android_arm64';
		if (arm) return 'linux_arm64';
		return 'linux_amd64';
	}

	let selected = $state<TargetId>(guess());
	const target = $derived(targets.find((t) => t.id === selected) ?? targets[0]);

	const packed = $derived(`portal_${version}_${target.asset}`);
	const archive = $derived(target.os === 'windows' ? `${packed}.zip` : `${packed}.tar.gz`);
	const url = $derived(`https://github.com/omidz4t/portal/releases/download/${tag}/${archive}`);

	const snippet = $derived.by(() => {
		if (target.os === 'windows') {
			return [
				`Invoke-WebRequest -Uri "${url}" -OutFile portal.zip`,
				`Expand-Archive -Force portal.zip .`,
				`Rename-Item ${packed}.exe portal.exe`,
				`.\\portal.exe --version`,
				`.\\portal.exe --config config.yml init dcaccount:nine.testrun.org`,
				`.\\portal.exe --config config.yml serve`
			].join('\n');
		}
		return [
			`curl -fsSL -o portal.tgz "${url}"`,
			`tar -xzf portal.tgz`,
			`mv ${packed} portal`,
			`chmod +x portal`,
			`./portal --version`,
			`./portal --config config.yml init dcaccount:nine.testrun.org`,
			`./portal --config config.yml serve`
		].join('\n');
	});

	const copyLabel = $derived(locale === 'fa' ? 'کپی' : 'Copy');
	const copiedLabel = $derived(locale === 'fa' ? 'کپی شد' : 'Copied');
	let copied = $state(false);

	async function copy() {
		try {
			await navigator.clipboard.writeText(snippet);
			copied = true;
			setTimeout(() => (copied = false), 1600);
		} catch {
			copied = false;
		}
	}

	const pickTitle = $derived(locale === 'fa' ? 'سیستم‌عامل و معماری' : 'Operating system');
</script>

<div class="os-picker" dir={locale === 'fa' ? 'rtl' : 'ltr'}>
	<p class="os-picker-label">{pickTitle}</p>
	<div class="os-picker-row" role="radiogroup" aria-label={pickTitle}>
		{#each targets as t (t.id)}
			<button
				type="button"
				class="os-chip"
				class:is-on={selected === t.id}
				role="radio"
				aria-checked={selected === t.id}
				onclick={() => (selected = t.id)}
			>
				<span class="os-chip-name">{t.label[locale]}</span>
				<span class="os-chip-hint">{t.hint[locale]}</span>
			</button>
		{/each}
	</div>

	<div class="self-host-snippet">
		<div class="os-meta" dir="ltr">
			<a class="doodle-btn doodle-btn-ghost self-host-dl" href={url}>{archive}</a>
			<button type="button" class="doodle-btn" onclick={copy}>{copied ? copiedLabel : copyLabel}</button>
		</div>
		<pre class="code-ltr" dir="ltr"><code>{snippet}</code></pre>
	</div>
</div>

<style>
	.os-picker {
		margin: 1.1rem 0 1.4rem;
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
	}

	.os-picker-label {
		margin: 0;
		font-weight: 700;
		font-size: 0.95rem;
	}

	.os-picker-row {
		display: flex;
		flex-wrap: wrap;
		gap: 0.5rem;
	}

	.os-chip {
		display: flex;
		flex-direction: column;
		align-items: flex-start;
		gap: 0.1rem;
		min-height: 2.75rem;
		border: 2px solid var(--color-ink);
		border-radius: 0.9rem;
		padding: 0.45rem 0.75rem;
		background: #fffdf8;
		text-align: start;
		cursor: pointer;
		box-shadow: 2px 2px 0 var(--color-ink);
	}

	.os-chip:hover {
		transform: translate(1px, 1px);
		box-shadow: 1px 1px 0 var(--color-ink);
	}

	.os-chip.is-on {
		background: var(--color-ink);
		color: #fffdf8;
	}

	.os-chip-name {
		font-weight: 700;
		font-size: 0.9rem;
	}

	.os-chip-hint {
		font-size: 0.75rem;
		opacity: 0.8;
	}

	.os-meta {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 0.5rem;
	}
</style>
