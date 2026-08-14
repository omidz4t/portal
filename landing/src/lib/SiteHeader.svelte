<script lang="ts">
	import { botUrl, repoUrl } from '$lib/links';
	import type { Copy, Locale } from '$lib/content';
	import { basePath, docsPath, twinPath } from '$lib/content';

	let {
		copy,
		locale,
		pathname = '/',
		current = ''
	}: {
		copy: Copy;
		locale: Locale;
		pathname?: string;
		current?: string;
	} = $props();

	const base = $derived(basePath(locale));
	const links = $derived([
		{ href: `${base}#how`, label: copy.navHow, id: 'how' },
		{ href: `${base}#try`, label: copy.navTry, id: 'try' },
		{ href: `${base}#persona`, label: copy.navPersona, id: 'persona' },
		{ href: docsPath(locale), label: copy.navDocs, id: 'docs' },
		{ href: repoUrl, label: copy.navSource, id: 'source', external: true }
	]);
</script>

<header class="doodle-bar">
	<div class="mx-auto flex max-w-6xl items-center justify-between gap-3 px-5 py-3">
		<a href={base} class="flex min-h-11 items-center gap-3 rounded-lg px-1">
			<img
				src="/avatar.png"
				alt=""
				width="36"
				height="36"
				class="h-9 w-9 rounded-full border-2 border-ink"
				aria-hidden="true"
			/>
			<span class="text-sm font-semibold tracking-wide">{copy.brandName}</span>
		</a>

		<nav class="hidden items-center gap-1 md:flex" aria-label="Primary">
			{#each links as link}
				<a
					href={link.href}
					class="doodle-nav"
					class:is-current={current === link.id}
					aria-current={current === link.id ? 'page' : undefined}
					rel={link.external ? 'noreferrer' : undefined}
				>
					{link.label}
				</a>
			{/each}
		</nav>

		<div class="hidden items-center gap-2 md:flex">
			<nav class="flex items-center gap-1 text-sm" aria-label="Language">
				<a
					class="doodle-lang"
					class:is-current={locale === 'en'}
					href={twinPath(pathname, 'en')}
					hreflang="en">EN</a
				>
				<span aria-hidden="true">·</span>
				<a
					class="doodle-lang"
					class:is-current={locale === 'fa'}
					href={twinPath(pathname, 'fa')}
					hreflang="fa">فا</a
				>
			</nav>
			<a href={botUrl} rel="noreferrer" class="doodle-btn doodle-btn-ink">{copy.openBot}</a>
		</div>

		<details class="relative md:hidden">
			<summary class="doodle-btn min-w-11">{copy.navMenu}</summary>
			<div class="doodle-card absolute end-0 z-20 mt-2 w-56 p-2">
				<nav class="flex flex-col" aria-label="Mobile">
					{#each links as link}
						<a
							href={link.href}
							class="inline-flex min-h-11 items-center rounded-lg px-3 text-sm"
							aria-current={current === link.id ? 'page' : undefined}
							rel={link.external ? 'noreferrer' : undefined}
						>
							{link.label}
						</a>
					{/each}
					<div class="flex gap-2 px-3 py-2 text-sm">
						<a href={twinPath(pathname, 'en')} hreflang="en">EN</a>
						<a href={twinPath(pathname, 'fa')} hreflang="fa">فا</a>
					</div>
					<a href={botUrl} rel="noreferrer" class="doodle-btn doodle-btn-ink mt-1 justify-center">
						{copy.openBot}
					</a>
				</nav>
			</div>
		</details>
	</div>
</header>
