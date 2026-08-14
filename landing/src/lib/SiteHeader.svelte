<script lang="ts">
	import { botUrl, repoUrl } from '$lib/links';
	import KoboyoIcon from '$lib/KoboyoIcon.svelte';
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

	let menu: HTMLDetailsElement | undefined = $state();

	function setMenuLock(open: boolean) {
		if (typeof document === 'undefined') return;
		document.body.style.overflow = open ? 'hidden' : '';
	}

	function onMenuToggle(event: Event) {
		setMenuLock((event.currentTarget as HTMLDetailsElement).open);
	}

	function closeMenu() {
		if (menu) menu.open = false;
		setMenuLock(false);
	}

	function onMenuNavClick(event: MouseEvent) {
		const target = event.target;
		if (target instanceof Element && target.closest('a')) closeMenu();
	}

	$effect(() => {
		return () => setMenuLock(false);
	});
</script>

<header class="doodle-bar">
	<div class="mx-auto flex max-w-6xl items-center justify-between gap-3 px-5 py-3">
		<a href={base} class="flex min-h-11 items-center gap-1.5 rounded-lg px-1">
			<KoboyoIcon name="cartoon-portal-render" class="h-8 w-8" />
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

		<details
			bind:this={menu}
			class="site-menu md:hidden"
			ontoggle={onMenuToggle}
		>
			<summary class="doodle-btn min-w-11 site-menu-toggle">
				<span class="site-menu-label-open">{copy.navMenu}</span>
				<span class="site-menu-label-close">{copy.navClose}</span>
			</summary>
			<div class="site-menu-panel">
				<nav class="site-menu-nav" aria-label="Mobile" onclick={onMenuNavClick}>
					{#each links as link}
						<a
							href={link.href}
							class="site-menu-link"
							class:is-current={current === link.id}
							aria-current={current === link.id ? 'page' : undefined}
							rel={link.external ? 'noreferrer' : undefined}
						>
							{link.label}
						</a>
					{/each}
					<div class="site-menu-langs">
						<a
							class="doodle-lang"
							class:is-current={locale === 'en'}
							href={twinPath(pathname, 'en')}
							hreflang="en">{copy.langEn}</a
						>
						<span aria-hidden="true">·</span>
						<a
							class="doodle-lang"
							class:is-current={locale === 'fa'}
							href={twinPath(pathname, 'fa')}
							hreflang="fa">{copy.langFa}</a
						>
					</div>
					<a href={botUrl} rel="noreferrer" class="doodle-btn doodle-btn-ink site-menu-bot">
						{copy.openBot}
					</a>
				</nav>
			</div>
		</details>
	</div>
</header>
