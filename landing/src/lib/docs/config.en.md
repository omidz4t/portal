# config.yml

The whole operator file: every key, default, and comment. Secrets stay in `.env`. Copy the example; never commit `config.yml`, `.env`, or `./data`.

## Two files on purpose

Portal splits settings and secrets:

- `config.yml` — mode, profile, storage, logging, proxies, Telegram UI, bridge toggles. Gitignored.
- `config.example.yml` — tracked template. Copy it with `make config`.
- `.env` — `TELEGRAM_BOT_TOKEN`, optional `PERSONA_ACCOUNT_QR`, `INVITE_URL`, `TGPORTAL_DB_KEY`, allow-list, proxy URLs. Gitignored.
- `.env.example` — tracked template for secrets.

Create locals if they are missing:

```bash
make config
# or:
cp config.example.yml config.yml
cp .env.example .env
```

## Full config.yml

This is the complete file Portal reads. Comments are part of the template — leave them in when you copy. Values shown are the documented defaults.

<!--full-config-->

## Environment (`.env`)

Never put BotFather tokens in YAML. User-owned tokens from `/pair-bot` live only in SQLite under `folder`.

- `TELEGRAM_BOT_TOKEN` — required for the Telegram bridge
- `PERSONA_ACCOUNT_QR` — `dcaccount:` / `dclogin:` URI to provision ghost accounts (persona / both)
- `INVITE_URL` — optional admin Delta Chat invite for the boot message
- `TGPORTAL_DB_KEY` — 32-byte hex key (`openssl rand -hex 32`) when `database_encrypt` is true
- `TELEGRAM_ALLOWED_USER_IDS` — comma-separated IDs; overrides `telegram.allowed_user_ids` when set
- `PROXY_URL`, `TELEGRAM_PROXY_URL`, `DELTACHAT_PROXY_URL`, `PROXY_ENABLED` — optional proxy overrides

```bash
TELEGRAM_BOT_TOKEN=123456:ABC-DEF...
# PERSONA_ACCOUNT_QR=dcaccount:nine.testrun.org
# INVITE_URL=https://i.delta.chat/#...
# TGPORTAL_DB_KEY=
# TELEGRAM_ALLOWED_USER_IDS=123456789
# PROXY_URL=socks5://127.0.0.1:1080
# PROXY_ENABLED=true
# TELEGRAM_PROXY_URL=socks5://127.0.0.1:1080
# DELTACHAT_PROXY_URL=socks5://127.0.0.1:1080
```

## What wins

Data directory: CLI `--folder` / `-f`, then `folder` in `config.yml`, then `./data`.

Telegram allow-list: `TELEGRAM_ALLOWED_USER_IDS` if set, then `telegram.allowed_user_ids`, then unrestricted.

Config path: CLI `--config` / `-c` (default `config.yml`). Makefile: `make serve CONFIG=path/to.yml`.

## Assets and runtime files

- `assets/logo.jpg` — Telegram branding and default Delta Chat avatar
- `assets/start_black_hole.mp4` — `/start` animation
- `folder/accounts/` — Delta Chat account databases
- `folder/tgportal.db` — pairing and persona store
- `folder/tg-cache/` — temporary Telegram downloads

Operator notes in the repo: [docs/configuration.md](https://github.com/omidz4t/portal/blob/main/docs/configuration.md) · [config.example.yml](https://github.com/omidz4t/portal/blob/main/config.example.yml) · [self-host](../self-host/) · [persona](../persona/).
