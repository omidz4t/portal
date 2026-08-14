# Contributing to Portal

Thank you for wanting to help. Portal is a **Telegram ↔ Delta Chat** bridge. Keep diffs small, reviewable, and free of secrets.

## Quick start

```bash
git clone https://github.com/omidz4t/portal.git
cd portal
make config          # creates config.yml and .env if missing
make test
make build           # → ./portal
```

You need **Go 1.22+** and [`deltachat-rpc-server`](https://github.com/chatmail/core/tree/main/deltachat-rpc-server) on `PATH` (match the v2.56 line in `go.mod`).

Routine tasks go through the **Makefile** (`make serve`, `make init QR=…`). See [docs/development.md](docs/development.md) and [AGENTS.md](AGENTS.md).

## Key rules

- **The operator is in the path.** The host can read pairing data and every file the bridge touches. Do not add logging of tokens, pairing codes, message bodies, or BotFather tokens.
- **No secrets in git.** Never commit `.env`, local `config.yml`, `./data/`, or real invite URLs with auth material.
- **Media must not sit on disk.** Telegram/DC cache files are deleted after a successful forward. Do not add a “keep stickers/videos in `data/`” feature.
- **Do not block `Bot.Run`.** Delta Chat `OnNewMsg` work must run in a goroutine. Use `dc.Session` for RPC — never call `bot.Rpc` from new code.
- **AI-assisted code:** you must read and understand every line you submit. Blind paste is not allowed.
- **Tests:** new logic under `internal/` needs a test where it is cheap (store, config, parse, limits).
- **Commits:** [Conventional Commits](https://www.conventionalcommits.org/), imperative mood, **signed** (`git commit -S`). One focused commit per change.

Examples: `feat: add unpair command`, `fix: drop blob after send`, `docs: document pairing`.

## How to contribute

1. Open an **issue** first unless the change is a one-line typo.
2. Branch from `main`:
   - `feat/short-name`
   - `fix/short-name`
   - `docs/short-name`
3. Implement under `internal/…`. Keep `cmd/portal` thin.
4. Update `config.example.yml` / `.env.example` if you add settings.
5. `make test` and `make build`.
6. Open a pull request against **`main`**.

**Pull requests never get production SSH secrets.** Deploy runs only on `main` after a version bump, or via a manual workflow with write access. See [SECURITY.md](SECURITY.md).

## What not to send

- Tokens, production `.env`, or copies of `/etc/portal/`
- Exploit PoCs against Telegram, chatmail, or a live host
- Drive-by refactors unrelated to the issue

## License

By contributing you agree your work is licensed under the same [MIT](LICENSE) license as the rest of the repository.
