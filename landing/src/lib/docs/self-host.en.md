# Self-host

Recommended. You keep the token, the database, and the ghost keys. Pick one path: build the binary, or download a release.

## What you need

Both ways need the same pieces after you have a `portal` binary:

- `deltachat-rpc-server` on `PATH` (match the project’s rpc-client major; this repo uses v2.56.x APIs)
- A BotFather token in `.env`
- A chatmail / Delta Chat provider for `make init` / `portal init`

Then copy `config.example.yml` → `config.yml` and `.env.example` → `.env`. Data default is `./data`. Never commit `.env`, `config.yml`, or `data/`.

## Way 1 — build the binary

Use this when you have Go 1.22+ and want to compile from git. The Makefile is the supported entrypoint.

```bash
git clone https://github.com/omidz4t/portal.git
cd portal
make config
# edit .env: TELEGRAM_BOT_TOKEN=…
# optional persona: PERSONA_ACCOUNT_QR=dcaccount:…
# edit config.yml: mode: personal | persona | both

make build
./portal --version

# first account, then run
make init QR=dcaccount:nine.testrun.org
make serve
# same as: ./portal --config config.yml serve
```

All-in-one static binary (stripped, version stamped) goes to `dist/portal`:

```bash
make build-release
./dist/portal --version
./dist/portal --config config.yml serve
```

## Way 2 — download a release

Use this when you do not want a Go toolchain. GitHub Actions publish archives on each tag `v*`: `portal_<tag>_<os>_<arch>.tar.gz` (Windows `.zip`) plus `checksums.txt`.

1. Open the [latest release](https://github.com/omidz4t/portal/releases/latest) or the [releases list](https://github.com/omidz4t/portal/releases).
2. Download the archive for your OS/CPU (examples: `linux_amd64`, `linux_arm64`, `darwin_arm64`, `windows_amd64`).
3. Verify the SHA-256 against `checksums.txt`.
4. Unpack, mark executable, put it next to `config.yml` and `.env`.

<!--os-picker-->

Pick your OS above. The archive name matches the GitHub release asset. You still need `deltachat-rpc-server` on `PATH`; the release is only the Go bot.

## This website, not the bot

That is how you run Portal. To edit this website (not the bot), use `make run-landing` and open http://127.0.0.1:5173.

More: [installation](https://github.com/omidz4t/portal/blob/main/docs/installation.md) · [configuration](../config/) · [security](https://github.com/omidz4t/portal/blob/main/docs/security.md) · [trust](../trust/).
