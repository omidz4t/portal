# Configuration

TGPORTAL splits **secrets** (environment) from **behavior** (YAML).

| File | Tracked in git? | Role |
|------|-----------------|------|
| `.env` | no | Tokens and optional admin invite |
| `.env.example` | yes | Template |
| `config.yml` | no | Local settings |
| `config.example.yml` | yes | Template |
| `./data/**` | no | Runtime state |

Create locals with:

```bash
make config
# or:
cp config.example.yml config.yml
cp .env.example .env
```

---

## Environment (`.env`)

| Variable | Required | Description |
|----------|----------|-------------|
| `TELEGRAM_BOT_TOKEN` | **yes** (for TG bridge) | Portal BotFather token |
| `PERSONA_ACCOUNT_QR` | for persona mode | `dcaccount:` / `dclogin:` URI to create ghost DC accounts |
| `INVITE_URL` | no | Admin Delta Chat invite for optional boot message |
| `TGPORTAL_DB_KEY` | when `database_encrypt: true` | 32-byte hex key (`openssl rand -hex 32`) |
| `TELEGRAM_ALLOWED_USER_IDS` | no | Comma-separated Telegram user IDs; overrides YAML allow-list when set |

Example:

```env
TELEGRAM_BOT_TOKEN=123456:ABC-DEF...
# PERSONA_ACCOUNT_QR=dcaccount:nine.testrun.org
# INVITE_URL=https://i.delta.chat/#...
# TELEGRAM_ALLOWED_USER_IDS=123456789
```

Never commit real tokens. Rotate in BotFather if a token leaks. User-owned tokens from `/pair-bot` are stored in SQLite under `folder` (not in YAML).

---

## YAML (`config.yml`)

### Mode

```yaml
mode: personal   # personal | persona | both

persona:
  account_qr: ""           # override with PERSONA_ACCOUNT_QR
  max_ghosts: 200
  max_ghosts_per_bot: 200
  max_bots: 20
  max_bots_per_owner: 3
  allow_register_from_tg: true
  allow_groups: false
```

See [persona-design.md](persona-design.md).

### Pairing

| Key | Default | Description |
|-----|---------|-------------|
| `pairing.code_length` | `8` | Code length (4–12) |
| `pairing.pending_ttl_sec` | `1800` | Unused code lifetime (seconds) |

### Database encryption

| Key | Default | Description |
|-----|---------|-------------|
| `database_encrypt` | `true` (example) | Refuse to open the store without a key |
| `database_key` | empty | Do not set in YAML; use `TGPORTAL_DB_KEY` |

### Proxy (SOCKS5 / HTTP)

See **[proxy.md](proxy.md)** for full details.

```yaml
proxy:
  enabled: false
  url: ""   # socks5://127.0.0.1:1080  or  http://127.0.0.1:8080

telegram:
  proxy:
    enabled: false
    url: ""

deltachat:
  proxy:
    enabled: false
    url: ""
```

Env: `PROXY_URL`, `TELEGRAM_PROXY_URL`, `DELTACHAT_PROXY_URL`, `PROXY_ENABLED`.

### Profile (Delta Chat)

| Key | Default | Description |
|-----|---------|-------------|
| `name` | `TGPORTAL` | Display name (`displayname`) |
| `image` | `./assets/logo.jpg` | Avatar path (`selfavatar`); empty skips avatar update |
| `reply` | `hi` | Text reply for unpaired DC messages |
| `boot_message` | … | Optional message to `INVITE_URL` on serve |

### Logging

| Key | Default | Description |
|-----|---------|-------------|
| `log` | `false` | Off. Set to `stderr`, `stdout`, or a file path (e.g. `./data/tgportal.log`) |
| `log_level` | `info` | `debug` · `info` · `warn` · `error` (when logging is enabled) |

```yaml
log: false                 # default — no app logs
# log: stderr
# log: ./data/tgportal.log
log_level: info
```

Note: Delta Chat core may still print its own lines to the process console; `log` controls TGPORTAL / bridge / botcli zap logs.

### Storage

| Key | Default | Description |
|-----|---------|-------------|
| `folder` | `./data` | Data directory (same as `--folder`) |
| `database` | `tgportal.db` | SQLite file (relative to `folder` unless absolute) |
| `account` | `0` | Limit CLI ops to account id; `0` = all |

### Telegram

| Key | Default | Description |
|-----|---------|-------------|
| `telegram.enabled` | `true` | Start Telegram long-poll when token is set |
| `telegram.bot_url` | `https://t.me/tgdeltabridgebot` | Shown in messages |
| `telegram.logo` | `./assets/logo.jpg` | Photo on `/start` / `/help` |
| `telegram.start_animation` | `./assets/start_black_hole.mp4` | Animation on `/start` (caption = instructions) |
| `telegram.allowed_user_ids` | `[]` | Empty = allow any user who DMs the bot |
| `telegram.reaction` | `✅` | Emoji reaction after successful bridge; `off` / `none` / `-` to disable |

### Bridge toggles

| Key | Default | Description |
|-----|---------|-------------|
| `bridge.stickers` | `true` | Static WEBP stickers |
| `bridge.lottie` | `true` | Animated TGS |
| `bridge.video_stickers` | `true` | WEBM stickers |
| `bridge.text` | `true` | Plain text both ways |
| `bridge.images` | `true` | Photos |
| `bridge.videos` | `true` | Short videos (see limits) |
| `bridge.gif` | `true` | GIF + animations |
| `bridge.custom_emoji` | `true` | Telegram custom/premium emojis |
| `bridge.sticker_packs` | `true` | `/send_pack` (full pack from a quoted sticker) |
| `bridge.limits.video_max_duration_sec` | `60` | Max video length (seconds); `0` after load = use default unless set `-1` for unlimited |
| `bridge.limits.video_max_bytes` | `20971520` | Max video size (20 MiB) |
| `bridge.limits.image_max_bytes` | `10485760` | Max image size (10 MiB) |
| `bridge.limits.file_max_bytes` | `20971520` | Max other files (20 MiB) |
| `bridge.limits.sticker_pack_max` | `120` | Max stickers per `/send_pack` |

---

## Precedence

**Data directory (`folder`):**

1. CLI `--folder` / `-f` if set  
2. `folder` in `config.yml`  
3. Default `./data`

**Telegram allow-list:**

1. `TELEGRAM_ALLOWED_USER_IDS` if non-empty  
2. `telegram.allowed_user_ids` in YAML  
3. Empty = unrestricted

**Config file path:**

- CLI `--config` / `-c` (default `config.yml`)
- Makefile: `make serve CONFIG=path/to.yml`

---

## Assets

| Path | Use |
|------|-----|
| `assets/logo.jpg` | Telegram branding + default DC avatar |
| `assets/start_black_hole.mp4` | `/start` animation |
| `assets/avatar.png` | Legacy default (optional) |

Paths in config may be relative (resolved from process working directory).

---

## Runtime files under `folder`

| Path | Description |
|------|-------------|
| `accounts/` | Delta Chat account databases (core) |
| `tgportal.db` | Pairing SQLite DB |
| `tg-cache/` | Temporary Telegram downloads |

All of the above should stay out of version control.
