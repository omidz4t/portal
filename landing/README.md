# Portal landing

SvelteKit (Svelte 5, static adapter, Tailwind) marketing site for the Telegram ↔ Delta Chat bridge.

From the repo root:

```sh
make run-landing
```

Dev server: [http://127.0.0.1:5173](http://127.0.0.1:5173)

Routes: `/` (English), `/fa/` (Farsi, [Arad](https://github.com/MohamadDarvishi/Arad) FD webfonts), `/docs/…` and `/fa/docs/…` (trust, persona, pairing, self-host).

Farsi strings were drafted with `llmcli ask -m openrouter/google/gemini-3.5-flash-lite` from `src/lib/content/en.json`.

Look: light paper + ink doodle (Koboyo-like). Icons from [koboyo.com/icons](https://koboyo.com/icons).

```sh
cd landing
npm install
npm run dev      # same as make run-landing
npm run build    # static files in landing/build
npm run preview
```

Production URL: [https://omidz4t.github.io/portal](https://omidz4t.github.io/portal)

CI (`.github/workflows/landing.yml`) rebuilds and deploys to GitHub Pages when `landing/` changes on `main`. Local / webxdc builds leave `BASE_PATH` empty. The Pages job sets `BASE_PATH=/portal`.

Webxdc (share the site in Delta Chat):

```sh
make landing-xdc
```

Writes `dist/portal.xdc` (zip of the static build + `manifest.toml` + `icon.png`). Drop that file in a chat and tap Start. Do not commit the `.xdc`.
