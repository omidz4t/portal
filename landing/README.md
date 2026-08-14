# TGPORTAL landing

SvelteKit (Svelte 5, static adapter, Tailwind) marketing site for the Telegram ↔ Delta Chat bridge.

From the repo root:

```sh
make run-landing
```

Dev server: [http://127.0.0.1:5173](http://127.0.0.1:5173)

Routes: `/` (English), `/fa/` (Farsi, [Arad](https://github.com/MohamadDarvishi/Arad) FD webfonts), `/docs`, `/docs/trust`, `/docs/persona`, `/docs/pairing`, `/docs/self-host`.

Farsi strings were drafted with `llmcli ask -m openrouter/google/gemini-3.5-flash-lite` from `src/lib/content/en.json`.

Look: light paper + ink doodle (Koboyo-like). Icons from [koboyo.com/icons](https://koboyo.com/icons).

```sh
cd landing
npm install
npm run dev      # same as make run-landing
npm run build    # static files in landing/build
npm run preview
```
