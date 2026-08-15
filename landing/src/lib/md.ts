/** Docs markdown: title/lede split + mdsvex-style HTML via marked (GFM). */

import { marked, type Tokens } from 'marked';

marked.use({
	gfm: true,
	breaks: false,
	renderer: {
		code({ text, lang }: Tokens.Code) {
			const cls = lang ? ` class="language-${esc(lang)}"` : '';
			return `<pre class="code-ltr" dir="ltr"><code${cls}>${esc(text)}</code></pre>\n`;
		}
	}
});

export function splitMarkdown(src: string): { title: string; lede: string; body: string } {
	const lines = src.replace(/^\uFEFF/, '').split(/\r?\n/);
	let i = 0;
	while (i < lines.length && lines[i].trim() === '') i++;
	let title = '';
	if (lines[i]?.startsWith('# ')) {
		title = lines[i].slice(2).trim();
		i++;
	}
	while (i < lines.length && lines[i].trim() === '') i++;
	const ledeParts: string[] = [];
	while (i < lines.length && lines[i].trim() !== '' && !lines[i].startsWith('#')) {
		ledeParts.push(lines[i]);
		i++;
	}
	return { title, lede: ledeParts.join(' ').trim(), body: lines.slice(i).join('\n').trim() };
}

export function renderMarkdown(src: string): string {
	if (!src.trim()) return '';
	return marked.parse(src, { async: false }) as string;
}

function esc(s: string): string {
	return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');
}
