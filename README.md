# Portal

Bidirectional **Telegram ↔ Delta Chat** bridge for text, stickers, GIFs, images, and short video.

Public bot: [@tgdeltabridgebot](https://t.me/tgdeltabridgebot)  
Site: [https://omidz4t.github.io/portal](https://omidz4t.github.io/portal)  
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
```

## Docs

| | |
|---|---|
| [docs/installation.md](docs/installation.md) | Install |
| [docs/pairing.md](docs/pairing.md) | Pairing |
| [docs/persona.md](docs/persona.md) | Persona / ghost accounts |
| [docs/security.md](docs/security.md) | Operator hardening |
| [SECURITY.md](SECURITY.md) | Report a vulnerability |
| [docs/trust.md](docs/trust.md) | Why self-host |
| [CONTRIBUTING.md](CONTRIBUTING.md) | Contributing |

## License

[MIT](LICENSE)
