import en from './content/en.json';
import fa from './content/fa.json';
import docsEn from './content/docs-en.json';
import docsFa from './content/docs-fa.json';
import tutorialEn from './content/tutorial-en.json';
import tutorialFa from './content/tutorial-fa.json';

export type Copy = typeof en;
export type DocsCopy = typeof docsEn;
export type TutorialCopy = typeof tutorialEn;

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
export type Locale = 'en' | 'fa';

export const copies: Record<Locale, Copy> = { en, fa };
export const docsCopies: Record<Locale, DocsCopy> = { en: docsEn, fa: docsFa };
export const tutorialCopies: Record<Locale, TutorialCopy> = { en: tutorialEn, fa: tutorialFa };

export function localeFromPath(pathname: string): Locale {
	return pathname === '/fa' || pathname.startsWith('/fa/') ? 'fa' : 'en';
}

export function basePath(locale: Locale): string {
	return locale === 'fa' ? '/fa/' : '/';
}

export function showPath(locale: Locale): string {
	return locale === 'fa' ? '/fa/show/' : '/show/';
}

export function docsPath(locale: Locale, slug = ''): string {
	const root = locale === 'fa' ? '/fa/docs/' : '/docs/';
	return slug ? `${root}${slug}/` : root;
}

/** Same page in the other language (for the EN / فا switch). */
export function twinPath(pathname: string, target: Locale): string {
	const rest = pathname.replace(/^\/fa(?=\/|$)/, '') || '/';
	if (target === 'en') {
		return rest.endsWith('/') || rest === '/' ? rest : `${rest}/`;
	}
	if (rest === '/') return '/fa/';
	return `/fa${rest.endsWith('/') ? rest : `${rest}/`}`;
}
