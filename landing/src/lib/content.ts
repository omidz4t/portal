import { base } from '$app/paths';
import en from './content/en.json';
import fa from './content/fa.json';
import docsEn from './content/docs-en.json';
import docsFa from './content/docs-fa.json';
import tutorialEn from './content/tutorial-en.json';
import tutorialFa from './content/tutorial-fa.json';

export type Copy = typeof en;
export type DocsCopy = typeof docsEn;

export type InboxRow = {
	name: string;
	preview: string;
	when: string;
	icon: string;
	focus?: boolean;
};

export type TutorialScene = {
	caption: string;
	why: string;
	tgTitle: string;
	dcTitle: string;
	tgIcon?: string;
	dcIcon?: string;
	view?: string;
	inboxTg?: InboxRow[];
	inboxDc?: InboxRow[];
	bubbles: { side: string; who: string; text: string; kind?: string; link?: string }[];
};

export type TutorialStory = {
	id: string;
	label: string;
	group: string;
	intro: string;
	telegram?: string;
	scenes: TutorialScene[];
};

export type TutorialCopy = {
	title: string;
	homeTitle: string;
	lead: string;
	open: string;
	close: string;
	homeBack: string;
	next: string;
	back: string;
	restart: string;
	stepOf: string;
	whyLabel: string;
	telegram: string;
	delta: string;
	you: string;
	emptyTg: string;
	emptyDc: string;
	stories: TutorialStory[];
};

export type Locale = 'en' | 'fa';

export const copies: Record<Locale, Copy> = { en, fa };
export const docsCopies: Record<Locale, DocsCopy> = { en: docsEn, fa: docsFa };
export const tutorialCopies: Record<Locale, TutorialCopy> = {
	en: tutorialEn as TutorialCopy,
	fa: tutorialFa as TutorialCopy
};

/** Pathname without SvelteKit `paths.base` (e.g. `/portal`). */
export function appPathname(pathname: string): string {
	if (base && (pathname === base || pathname.startsWith(`${base}/`))) {
		return pathname.slice(base.length) || '/';
	}
	return pathname;
}

export function localeFromPath(pathname: string): Locale {
	const path = appPathname(pathname);
	return path === '/fa' || path.startsWith('/fa/') ? 'fa' : 'en';
}

export function basePath(locale: Locale): string {
	return locale === 'fa' ? `${base}/fa/` : `${base}/`;
}

export function showPath(locale: Locale): string {
	return locale === 'fa' ? `${base}/fa/show/` : `${base}/show/`;
}

export function docsPath(locale: Locale, slug = ''): string {
	const root = locale === 'fa' ? `${base}/fa/docs/` : `${base}/docs/`;
	return slug ? `${root}${slug}/` : root;
}

export function versionsPath(locale: Locale): string {
	return locale === 'fa' ? `${base}/fa/versions/` : `${base}/versions/`;
}

/** Same page in the other language (for the EN / فا switch). */
export function twinPath(pathname: string, target: Locale): string {
	const rest = appPathname(pathname).replace(/^\/fa(?=\/|$)/, '') || '/';
	const withSlash = rest.endsWith('/') || rest === '/' ? rest : `${rest}/`;
	if (target === 'en') {
		return `${base}${withSlash === '/' ? '/' : withSlash}`;
	}
	if (withSlash === '/') return `${base}/fa/`;
	return `${base}/fa${withSlash}`;
}
