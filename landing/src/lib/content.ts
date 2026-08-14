import en from './content/en.json';
import fa from './content/fa.json';

export type Copy = typeof en;
export type Locale = 'en' | 'fa';

export const copies: Record<Locale, Copy> = { en, fa };

export function localeFromPath(pathname: string): Locale {
	return pathname === '/fa' || pathname.startsWith('/fa/') ? 'fa' : 'en';
}

export function basePath(locale: Locale): string {
	return locale === 'fa' ? '/fa/' : '/';
}
