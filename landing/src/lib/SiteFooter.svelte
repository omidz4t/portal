<script lang="ts">
	import { botUrl, repoUrl } from '$lib/links';
</script>

<footer class="border-t border-line/80">
	<div
		class="mx-auto flex max-w-6xl flex-wrap items-center justify-between gap-4 px-5 py-8 text-sm text-mist"
	>
		<p>
			TGPORTAL · MIT · themadorg · icons
			<a class="underline-offset-2 hover:underline" href="https://koboyo.com/icons" rel="noreferrer"
				>Koboyo</a
			>
		</p>
		<nav class="flex flex-wrap gap-1" aria-label="Footer">
			<a class="inline-flex min-h-11 items-center px-2 hover:text-paper" href="/docs/">Docs</a>
			<a class="inline-flex min-h-11 items-center px-2 hover:text-paper" href="/docs/trust/"
				>Trust</a
			>
			<a class="inline-flex min-h-11 items-center px-2 hover:text-paper" href="/docs/persona/"
				>Persona</a
			>
			<a
				class="inline-flex min-h-11 items-center px-2 hover:text-paper"
				href={repoUrl}
				rel="noreferrer">Source</a
			>
			<a
				class="inline-flex min-h-11 items-center px-2 hover:text-paper"
				href={botUrl}
				rel="noreferrer">Telegram</a
			>
		</nav>
	</div>
</footer>
