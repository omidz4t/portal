<script lang="ts">
	import { botUrl, repoUrl } from '$lib/links';
</script>

<footer class="border-t border-line/80">
	<div
		class="mx-auto flex max-w-6xl flex-wrap items-center justify-between gap-4 px-5 py-8 text-sm text-mist"
	>
		<p>TGPORTAL · MIT · themadorg</p>
		<div class="flex flex-wrap gap-4">
			<a class="hover:text-paper" href="/docs/">Docs</a>
			<a class="hover:text-paper" href="/docs/trust/">Trust</a>
			<a class="hover:text-paper" href="/docs/persona/">Persona</a>
			<a class="hover:text-paper" href={repoUrl}>Source</a>
			<a class="hover:text-paper" href={botUrl}>Telegram</a>
		</div>
	</div>
</footer>
