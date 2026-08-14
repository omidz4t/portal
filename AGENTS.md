# Agent instructions — TGPORTAL

These rules apply to every coding agent and human working in this repository.

**User-facing docs:** [README.md](README.md) and [docs/](docs/). Keep them updated when behavior or config changes.

**Project name:** TGPORTAL  
**Module:** `github.com/omidz4t/portal`  
**Binary / CLI app name:** `tgportal`  
**Data directory default:** `./data` (set via `folder` in `config.yml` or `--folder`)

### Concurrency (important)

The upstream `rpc-client-go` `Bot.Run()` processes DC events **one at a time** on a single goroutine; `OnNewMsg` runs **inline** on that loop. Blocking there freezes receive.

TGPORTAL mitigations:

- `internal/dc.Session` **serializes** app-level DC RPC (`SendMsg`, `GetMessage`, …); mutex released between send retries
- DC `OnNewMsg` work runs in a **goroutine** (never block `Bot.Run`)
- Telegram updates use a **bounded worker pool** (8) for concurrent download + bridge
- SQLite uses `MaxOpenConns(1)` (safe concurrent access)

Do not call `bot.Rpc` directly from new code — use `dc.Session`.

### Purpose

TGPORTAL bridges **Telegram ↔ Delta Chat** for media:

| Kind | Telegram | Delta Chat |
|------|----------|------------|
| Stickers | static WEBP | sticker/image |
| Lottie | animated TGS | file or GIF if converter installed |
| Video stickers | WEBM | video |
| GIF / animation | GIF or MP4 animation | gif/video |

**Direction:** Bidirectional Telegram ↔ Delta Chat (per-user pairing).

### Pairing flow

1. User opens https://t.me/tgdeltabridgebot and sends `/start`
2. Bot replies with the **Delta Chat invite link** + a **pairing code** (stored in SQLite)
3. User opens the DC bot and sends the code
4. SQLite row becomes `active` (`telegram_user_id` ↔ `dc_chat_id`)
5. Stickers/GIFs from that TG user go to their DC chat

DB: `./data/tgportal.db` (config `database`).  
Secrets: `TELEGRAM_BOT_TOKEN`, optional `INVITE_URL`, `TGPORTAL_DB_KEY` in `.env` only — never commit tokens.

## Commits (required)

- **Commit after each logical change.** Do not leave completed work uncommitted.
- **Use [Conventional Commits](https://www.conventionalcommits.org/):**

  ```
  <type>(optional scope): <short description>

  [optional body]
  ```

  Common types: `feat`, `fix`, `docs`, `chore`, `refactor`, `test`, `build`, `ci`, `style`, `perf`.

  Examples:

  - `feat: reply with configured greeting`
  - `docs: add AGENTS.md commit policy`
  - `chore: ignore build artifacts and secrets`

- **Always sign commits** (`git commit -S` or ensure `commit.gpgsign=true`).
- Prefer **one focused commit per change**; avoid bundling unrelated edits.
- Write clear subjects in the imperative mood (e.g. “add”, “fix”, not “added”).

### Never commit

- **Credentials / secrets:** `.env`, private keys, tokens, account QR secrets, real invite URLs with auth material
- **Local machine config:** `config.yml` (use `config.example.yml` / `.env.example` only)
- **Build products:** compiled binaries, object files, coverage outputs, vendor caches you did not intentionally track

If a secret is committed by mistake, rotate it and remove it from history; do not only delete it in a later commit without rotation.

## Makefile (required workflow)

The **`Makefile` is the canonical entrypoint** for build, config bootstrap, and bot lifecycle. Prefer `make …` over ad-hoc `go` / binary invocations so flags (`--config`), binary name, and local paths stay consistent.

| Target | Purpose |
|--------|---------|
| `make config` | Create local `config.yml` and `.env` from examples if missing |
| `make tidy` | `go mod tidy` |
| `make build` | Tidy + compile `./tgportal` |
| `make build-release` | All-in-one static binary → `dist/tgportal` (embedded assets) |
| `make build-release-all` | Cross-platform archives under `dist/` |
| `make run-landing` | Dev server for the SvelteKit site (`./landing`) |
| `make landing-xdc` | Build the site and pack `dist/portal.xdc` (webxdc) |
| `make run ARGS='…'` | Build and run with `CONFIG` (default `config.yml`) |
| `make init QR=dcaccount:…` | Build and configure a new account |
| `make serve` | Build and start the bot (`serve`) |
| `make help` | CLI help |
| `make clean` | Remove the built binary |
| `make version-dry` | Preview conventional-commit version bump |
| `make patch` | Force patch: VERSION + changelog + commit + tag |
| `make minor` | Force minor: VERSION + changelog + commit + tag |
| `make major` | Force major: VERSION + changelog + commit + tag |
| `make release-tag` | Same, bump chosen from conventional commits |
| `make run-landing` | SvelteKit landing in `./landing` (`npm run dev` on port 5173) |

**Agent rules for the Makefile:**

- **Use Makefile targets** for build, init, serve, and local config setup unless a task truly needs a one-off command.
- **Keep the Makefile accurate** when you change how the bot is built or run (new flags, binary name, config path, setup steps). Update targets and this section together.
- **Do not commit** the binary produced by `make build` (see `.gitignore`).
- After code changes, prefer `make build` (and `go test ./…`) before committing.
- Optional overrides: `CONFIG=…`, `QR=…`, `ARGS=…`.

## Project layout

Do **not** dump application logic in the repo root. Keep root for module metadata, Makefile, docs, and config templates.

```
cmd/tgportal/          # main package only — thin entrypoint
internal/config/       # YAML + env loading (mode: personal|persona|both)
internal/bot/          # Delta CLI wiring, profile, pairing on DC
internal/dc/           # Delta Chat helpers (open chat, send file/text, multi-account)
internal/store/        # SQLite pairing + persona_bots + ghost_accounts
internal/bridge/       # TG → DC media forward (personal mode)
internal/telegram/     # Portal Telegram bot, /start, /pair-bot
internal/persona/      # User-owned bots, ghost DC accounts, group mirror
```

| Path | Role |
|------|------|
| `cmd/tgportal/main.go` | `main`; calls `bot.Run()` |
| `internal/config` | `Config`, `Load`, `mode` / `persona` |
| `internal/bot` | deltabot-cli setup, DC pairing handler, start Telegram + persona |
| `internal/store` | SQLite pairs + persona_bots + ghost_* tables |
| `internal/dc` | Session RPC (incl. AddAccount / ConfigureAccountFromQR) |
| `internal/bridge` | media kinds + ForwardToDelta (personal pairing) |
| `internal/telegram` | Portal Bot API, `/pair`, `/pair-bot`, media |
| `internal/persona` | Multi-bot pollers, GetOrCreateGhost, group mirror |
| `Makefile` | build/run entrypoint (`go build ./cmd/tgportal` → `./tgportal`) |
| `config.example.yml`, `.env.example` | tracked templates |

### Persona mode (ghost accounts)

- Config: `mode: personal|persona|both` and `persona.account_qr` / `PERSONA_ACCOUNT_QR`
- Owners `/pair` then `/pair-bot <TOKEN>` on the **portal** bot (private chat)
- Each remote TG user → unique DC account (`ghost_accounts`), reused forever
- **Never log BotFather tokens**; they live only in SQLite under `folder`
- Docs: [docs/persona.md](docs/persona.md)

Put new features under `internal/…`. Only add code under `cmd/` when introducing another binary.

## Working style

- Keep diffs small and reviewable.
- Match existing Go style and project layout (`cmd/` + `internal/`).
- Update examples (`.example` files) when adding required config or env vars.
- Prefer **Makefile** workflows over raw `go build` / `./tgportal` for routine tasks.
- Run `make build` / `make test` before committing when code changes.

## Config layout

| File | Tracked? | Purpose |
|------|----------|---------|
| `config.example.yml` | yes | Template for bot settings |
| `config.yml` | no | Local settings |
| `.env.example` | yes | Template for secrets |
| `.env` | no | Secrets (e.g. `INVITE_URL`) |
| `assets/avatar.png` | yes | Default profile image |
| `data/` | no | Default bot data dir (`folder: ./data`) |

### Profile (name & image)

Set in `config.yml` and applied on every `serve`:

```yaml
name: TGPORTAL
image: ./assets/avatar.png   # empty = leave avatar unchanged
```

Maps to Delta Chat `displayname` and `selfavatar`.

### Data directory

- **Default:** `./data` (not `~/.config/…`)
- **Configure in** `config.yml`:

  ```yaml
  folder: ./data
  ```

- **CLI override:** `./tgportal --folder /other/path serve` (wins over config.yml)
- Precedence: `--folder` → `config.yml` `folder` → `./data`
- Never commit `data/`; it holds accounts and runtime state.

## Delta Chat bot notes

- Depends on `deltachat-rpc-server` on `PATH` (match rpc-client / deltabot-cli major version).
- Runtime data lives under `folder` (default `./data`), not elsewhere in the repo.
