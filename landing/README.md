# TGPORTAL landing

SvelteKit (Svelte 5, static adapter, Tailwind) marketing site for the Telegram ↔ Delta Chat bridge.

From the repo root:

```sh
make run-landing
```

Dev server: [http://127.0.0.1:5173](http://127.0.0.1:5173)

Routes: `/` (warning + persona), `/docs`, `/docs/trust`, `/docs/persona`, `/docs/pairing`, `/docs/self-host`.

UI notes (from [ui-ux-pro-max](https://github.com/nextlevelbuilder/ui-ux-pro-max-skill) **markdown only** — not installed): skip link, visible focus, 44px targets, contrast, reduced-motion, mobile menu, breadcrumbs.

```sh
cd landing
npm install
npm run dev      # same as make run-landing
npm run build    # static files in landing/build
npm run preview
```
