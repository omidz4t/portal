# TGPORTAL

**Telegram → Delta Chat media bridge** for stickers, Lottie (TGS), video stickers, and GIFs.

Pair once with a short code, then anything you send the Telegram bot is delivered into your Delta Chat conversation—without noisy filenames or “from Telegram” captions.

| | |
|---|---|
| **Telegram bot** | [@tgdeltabridgebot](https://t.me/tgdeltabridgebot) |
| **Direction** | **Bidirectional** Telegram ↔ Delta Chat (per-user pairing) |
| **Language** | Go |
| **Storage** | SQLite (`./data/tgportal.db`) |
| **Runtime data** | `./data` (gitignored) |

---

## Features

- **Sticker bridge** — static WEBP stickers
- **Lottie bridge** — animated TGS stickers (optional conversion to GIF if tools are installed)
- **Video sticker bridge** — WEBM stickers
- **GIF / animation bridge** — GIF documents and Telegram animations (often MP4)
- **Custom emoji bridge** — Telegram custom/premium emojis (via `getCustomEmojiStickers`)
- **Bidirectional** — text, images, short videos, stickers, GIFs **Telegram ↔ Delta Chat**
- **Short video limits** — configurable max duration/size in `bridge.limits`
- **`/start` onboarding** — logo + animation branding, Delta Chat invite link, pairing code
- **SQLite pairing** — maps each Telegram user to a Delta Chat chat
- **Persona mode** — register your own BotFather bot via `/pair-bot`; each person who messages it gets a **stable ghost Delta Chat account** (unique TG id bind, reused)
- **Group mirror (persona)** — TG groups mirrored as DC groups (`TG: …`)
- **Anonymous media** — neutral filenames, no descriptive captions on bridged files
- **Concurrent-safe DC access** — session lock + async handlers (see [docs/architecture.md](docs/architecture.md))

Persona guide: [docs/persona.md](docs/persona.md).

---

## Quick start

### Requirements

- Go **1.22+** (module targets current toolchain)
- [`deltachat-rpc-server`](https://github.com/chatmail/core/tree/main/deltachat-rpc-server) on `PATH`  
  Prefer a version matching the Go packages (this project uses **v2.56.x** client APIs)
- A [Telegram bot token](https://t.me/BotFather)
- Network access to a chatmail / Delta Chat provider (e.g. `nine.testrun.org` for testing)

### 1. Clone and configure

```bash
git clone https://github.com/themadorg/tgportal.git
cd tgportal

make config
# edits:
#   .env           → TELEGRAM_BOT_TOKEN=...
#   config.yml     → branding / bridge toggles (optional)
```

### 2. Create the Delta Chat bot account

```bash
make init QR=dcaccount:nine.testrun.org
```

Use any valid `dcaccount:` / `dclogin:` configuration URI for your provider.

### 3. Run

```bash
make serve
```

### 4. Pair and bridge

1. Open the Telegram bot → send **`/start`**
2. Open the **Delta Chat invite** from the message
3. Send the **pairing code** to the DC bot
4. Send a sticker or GIF on Telegram → it appears in Delta Chat

Full walkthrough: [docs/pairing.md](docs/pairing.md).

---

## Configuration overview

| Source | Purpose |
|--------|---------|
| `.env` | Secrets (`TELEGRAM_BOT_TOKEN`, optional `INVITE_URL`, proxy URLs) |
| `config.yml` | Name, avatar, folder, branding, bridge toggles, **proxies** |
| `./data` | Accounts DB, SQLite pairing DB, TG download cache |

**Proxies (SOCKS5 / HTTP):** set `proxy.url` and/or `telegram.proxy` / `deltachat.proxy` — see [docs/proxy.md](docs/proxy.md).

Examples (committed):

- [`config.example.yml`](config.example.yml)
- [`.env.example`](.env.example)

Details: [docs/configuration.md](docs/configuration.md).

---

## CI & releases

GitHub Actions:

| Workflow | Trigger | What it does |
|----------|---------|----------------|
| **CI** | push/PR to `main` | `go vet`, tests, multi-OS build artifacts |
| **Release** | tag `v*` (e.g. `v0.1.0`) | tests + binaries + GitHub Release |

Create a release (auto version from conventional commits — **no extra deps**):

```bash
make version-dry    # preview: fix→0.0.X, feat→0.X.0
make patch          # force patch  (0.0.X)
make minor          # force minor  (0.X.0)
make major          # force major  (X.0.0)
make release-tag    # auto from commits, commit VERSION, tag vX.Y.Z
git push origin HEAD --tags
```

Or tag manually:

```bash
git tag -s v0.1.0 -m "v0.1.0"
git push origin v0.1.0
```

Assets: `tgportal_<tag>_<os>_<arch>.tar.gz` (Windows `.zip`) plus `checksums.txt`.

See [.github/workflows/](.github/workflows/) and `scripts/bump-version.sh`.

---

## Makefile

The Makefile is the supported entrypoint for build and lifecycle:

| Target | Description |
|--------|-------------|
| `make config` | Create `config.yml` / `.env` from examples if missing |
| `make build` | Tidy + compile `./tgportal` |
| `make build-release` | All-in-one static binary in `dist/` (assets embedded) |
| `make test` | `go test ./...` |
| `make init QR=…` | Configure a new Delta Chat account |
| `make serve` | Build and run `serve` |
| `make run ARGS='…'` | Build and run arbitrary CLI args |
| `make help` | CLI help |
| `make clean` | Remove binary |
| `make run-landing` | Dev server for the SvelteKit site in `./landing` |

```bash
make serve
# equivalent:
#   go build -o tgportal ./cmd/tgportal
#   ./tgportal --config config.yml serve
```

---

## Project layout

```
cmd/tgportal/           # main binary
internal/
  bot/                  # deltabot-cli wiring, pairing on DC, profile
  bridge/               # TG → DC media forward
  config/               # YAML + env
  dc/                   # Session (serialized RPC), invite helpers
  store/                # SQLite pairs
  telegram/             # Bot API, /start, downloads
assets/                 # logo, start animation, default avatar
docs/                   # extended documentation
landing/                # SvelteKit marketing site (`make run-landing`)
config.example.yml
.env.example
Makefile
AGENTS.md               # contributor / coding-agent rules
```

Architecture notes: [docs/architecture.md](docs/architecture.md).

---

## Security

- **Never commit** `.env`, `config.yml` (local), or `./data/`
- Treat `TELEGRAM_BOT_TOKEN` like a password; rotate in BotFather if leaked
- Prefer `telegram.allowed_user_ids` or `TELEGRAM_ALLOWED_USER_IDS` in production
- Bridged media is sent with **anonymous** names and **no** source captions

See [docs/security.md](docs/security.md).

---

## Documentation

| Doc | Contents |
|-----|----------|
| [docs/installation.md](docs/installation.md) | Dependencies, install, first run |
| [docs/configuration.md](docs/configuration.md) | Full config and env reference |
| [docs/pairing.md](docs/pairing.md) | User pairing flow |
| [docs/architecture.md](docs/architecture.md) | Design, concurrency, packages |
| [docs/development.md](docs/development.md) | Hacking, tests, commits |
| [docs/security.md](docs/security.md) | Threat model and hard rules |
| [docs/trust.md](docs/trust.md) | Why you should self-host (operator sees bridged data) |
| [CONTRIBUTING.md](CONTRIBUTING.md) | How to contribute |
| [AGENTS.md](AGENTS.md) | Rules for automated coding agents |

---

## CLI

```bash
./tgportal --help
./tgportal init dcaccount:nine.testrun.org
./tgportal serve
./tgportal link          # print bot invite
./tgportal list          # accounts
```

Global flags:

- `-c, --config` — YAML path (default `config.yml`)
- `-f, --folder` — data directory (default `./data`, overridable in YAML)
- `-a, --account` — single account id when supported by the subcommand

---

## Roadmap (ideas)

- Delta Chat → Telegram reverse bridge  
- Better Lottie → GIF conversion by default  
- Admin commands to unpair / list pairs  
- Metrics (queue depth, send latency)

---

## License

[MIT](LICENSE) — see `LICENSE` for details.

## Acknowledgments

- [deltabot-cli-go](https://github.com/deltachat-bot/deltabot-cli-go)
- [rpc-client-go](https://github.com/chatmail/rpc-client-go) / [Delta Chat core](https://github.com/chatmail/core)
- [telegram-bot-api](https://github.com/go-telegram-bot-api/telegram-bot-api)
