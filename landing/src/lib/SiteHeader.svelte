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

	let menuOpen = $state(false);

	function setMenuLock(open: boolean) {
		if (typeof document === 'undefined') return;
		document.body.style.overflow = open ? 'hidden' : '';
	}

	function openMenu() {
		menuOpen = true;
		setMenuLock(true);
	}

	function closeMenu() {
		menuOpen = false;
		setMenuLock(false);
	}

	function onMenuNavClick(event: MouseEvent) {
		const target = event.target;
		if (target instanceof Element && target.closest('a')) closeMenu();
	}

	$effect(() => {
		if (typeof window === 'undefined') return;
		const desktop = window.matchMedia('(min-width: 768px)');
		function onDesktopChange() {
			if (desktop.matches) closeMenu();
		}
		desktop.addEventListener('change', onDesktopChange);
		return () => desktop.removeEventListener('change', onDesktopChange);
	});

	$effect(() => {
		if (!menuOpen) return;
		function onKey(event: KeyboardEvent) {
			if (event.key === 'Escape') closeMenu();
		}
		window.addEventListener('keydown', onKey);
		return () => {
			window.removeEventListener('keydown', onKey);
			setMenuLock(false);
		};
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

		<div class="site-menu-open">
			<button type="button" class="doodle-btn min-w-11" onclick={openMenu}>
				{copy.navMenu}
			</button>
		</div>
	</div>
</header>

{#if menuOpen}
	<div class="site-menu" role="dialog" aria-modal="true" aria-label={copy.navMenu}>
		<div class="site-menu-toolbar">
			<button type="button" class="doodle-btn min-w-11 site-menu-toggle" onclick={closeMenu}>
				{copy.navClose}
			</button>
		</div>
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
	</div>
{/if}
