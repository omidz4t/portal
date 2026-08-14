# Installation

## Dependencies

### Go

Install a recent Go toolchain (1.22+ recommended):

```bash
go version
```

### deltachat-rpc-server

TGPORTAL talks to Delta Chat through the standalone RPC server.

1. Install following upstream docs:  
   https://github.com/chatmail/core/tree/main/deltachat-rpc-server
2. Ensure it is on your `PATH`:

```bash
which deltachat-rpc-server
deltachat-rpc-server --version
```

**Version note:** Match the major/minor line of the Go libraries when possible.  
This repository depends on `deltabot-cli-go/v2` and `rpc-client-go/v2` around **v2.56**.  
Mismatched server versions can cause cryptic RPC errors.

### Telegram bot

1. Message [@BotFather](https://t.me/BotFather) → `/newbot`
2. Copy the token (looks like `123456:ABC-DEF...`)
3. Optional: set bot name, description, and profile photo in BotFather

### Optional: Lottie converters

Animated TGS stickers are forwarded as files by default. For GIF conversion when available:

- `lottie_to_png` + `ffmpeg`, or  
- `tgs-to-gif`

If neither is installed, TGS is still bridged as a file attachment.

---

## Get TGPORTAL (two ways)

### A) Build the binary from source

```bash
git clone https://github.com/omidz4t/portal.git
cd tgportal
make build          # → ./tgportal
# or
make build-release  # → ./dist/tgportal
```

Needs Go 1.22+. Then continue with First-time setup (`make config`, `make init`, `make serve`).

### B) Download a release binary

No Go toolchain. Assets on each `v*` tag: [releases](https://github.com/omidz4t/portal/releases).

Names: `tgportal_<tag>_<os>_<arch>.tar.gz` (Windows `.zip`) and `checksums.txt`.

```bash
# pick os/arch from the release page, then:
tar -xzf tgportal_<tag>_linux_amd64.tar.gz
chmod +x tgportal
./tgportal --version
```

You still need `deltachat-rpc-server` on `PATH`, plus `config.yml` / `.env` from the example files. Then:

```bash
./tgportal --config config.yml init dcaccount:nine.testrun.org
./tgportal --config config.yml serve
```

---

## First-time setup

```bash
# 1) Local config files (gitignored copies)
make config

# 2) Secrets
$EDITOR .env
# TELEGRAM_BOT_TOKEN=...
# INVITE_URL=...   # optional admin boot notify only

# 3) Optional: edit branding / toggles
$EDITOR config.yml

# 4) Create Delta Chat bot account (chatmail example)
make init QR=dcaccount:nine.testrun.org

# 5) Run
make serve
```

Data is written under `./data` by default (accounts, SQLite, cache).

### Public instance (one process for everyone)

Same binary. Use `mode: both`, leave `allowed_user_ids` empty, set `TGPORTAL_DB_KEY` and `database_encrypt: true`. See [security.md](security.md).

---

## Verify

```bash
make test
make build
./tgportal --help
./tgportal list
./tgportal link    # after init
```

When `serve` is healthy you should see logs similar to:

```text
sqlite ready at data/tgportal.db
Listening at: https://i.delta.chat/#...
authorized as @YourBot; bridging stickers/lottie/gif → Delta Chat
```

---

## Updating

```bash
git pull
make build test
make serve
```

Keep `deltachat-rpc-server` compatible with the module versions in `go.mod`.

---

## Uninstall / reset local data

```bash
make clean
rm -rf data/          # destroys DC account state and pairings
# keep or delete config.yml / .env as you prefer
```

Re-run `make init` after deleting `data/` to create a fresh bot account.
