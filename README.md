# Portal

Bidirectional **Telegram ↔ Delta Chat** bridge for text, stickers, GIFs, images, and short video.

Public bot: [@tgdeltabridgebot](https://t.me/tgdeltabridgebot)  
Site: [https://omidz4t.github.io/portal](https://omidz4t.github.io/portal)  
Image: [`ghcr.io/omidz4t/portal`](https://github.com/users/omidz4t/packages/container/package/portal)  
Binary: `portal`

> **Beta.** This software is unfinished and may lose messages, break pairing, or change without notice. Use it at your own risk. The host can read everything you bridge — prefer running your own instance. See [docs/trust.md](docs/trust.md).

## Quick start

Requires Go 1.22+, [`deltachat-rpc-server`](https://github.com/chatmail/core/tree/main/deltachat-rpc-server) on `PATH`, and a [BotFather](https://t.me/BotFather) token.

```bash
git clone https://github.com/omidz4t/portal.git
cd portal
make config          # then set TELEGRAM_BOT_TOKEN in .env
make init QR=dcaccount:nine.testrun.org
make serve
```

Pair: Telegram `/start` → open the Delta Chat invite → send the code to the bot.

Erase stored data: `/delete_my_data` then `/delete_my_data_approve` (Telegram or paired Delta Chat, private 1:1).

## Docker (GHCR)

Public image (amd64 + arm64). Includes `portal` and `deltachat-rpc-server`. Do not bake tokens into the image.

```bash
docker pull ghcr.io/omidz4t/portal:latest
# or: docker pull ghcr.io/omidz4t/portal:<git-sha>

# first account (once per data volume)
docker run --rm \
  -v "$PWD/config.yml:/etc/portal/config.yml:ro" \
  --env-file .env \
  -v portal-data:/var/lib/portal \
  ghcr.io/omidz4t/portal:latest \
  --config /etc/portal/config.yml --folder /var/lib/portal \
  init dcaccount:nine.testrun.org

docker run --rm \
  -v "$PWD/config.yml:/etc/portal/config.yml:ro" \
  --env-file .env \
  -v portal-data:/var/lib/portal \
  ghcr.io/omidz4t/portal:latest
```

Tags: `latest`, the git SHA, and the `VERSION` file. Details: [docs/docker.md](docs/docker.md).

## Config

| File | Role |
|------|------|
| `.env` | Secrets (`TELEGRAM_BOT_TOKEN`, optional `INVITE_URL`, `TGPORTAL_DB_KEY`) |
| `config.yml` | Mode, profile, proxies, bridge toggles |
| `./data` | Accounts, SQLite, cache — do not commit |

Templates: [`config.example.yml`](config.example.yml), [`.env.example`](.env.example). Reference: [docs/configuration.md](docs/configuration.md).

```bash
make serve
make build-release   # → ./dist/portal
make test
./portal --version
docker pull ghcr.io/omidz4t/portal:latest
```

## Docs

| | |
|---|---|
| [docs/installation.md](docs/installation.md) | Install |
| [docs/docker.md](docs/docker.md) | Container / GHCR |
| [docs/pairing.md](docs/pairing.md) | How to pair Telegram and Delta Chat |
| [docs/persona.md](docs/persona.md) | Persona / ghost accounts |
| [docs/security.md](docs/security.md) | Operator hardening |
| [SECURITY.md](SECURITY.md) | Report a vulnerability |
| [docs/trust.md](docs/trust.md) | Why self-host |
| [CONTRIBUTING.md](CONTRIBUTING.md) | Contributing |

## Disclaimer

The product vision, architecture, and operator docs were defined and reviewed by humans. Most of the Go (and related) source in this repository was written with AI assistance under that direction — not as an unattended dump of generated code, but as an iterative, human-guided process.

Use at your own risk. Portal is **MIT** software in **beta**; it may lose messages, break pairing, or change without notice. The host can read everything you bridge. Run it in production only after you have validated it for your threat model. See [docs/trust.md](docs/trust.md) and [SECURITY.md](SECURITY.md).

We welcome criticism, bug reports, and discussion — please use [GitHub Discussions](https://github.com/omidz4t/portal/discussions) or [Issues](https://github.com/omidz4t/portal/issues).

## License

[MIT](LICENSE)
