# Operator security

Portal is not an end-to-end tunnel past the host. The process that runs `portal` downloads Telegram files and injects them into Delta Chat (and the other way). Treat the box as a participant.

## Files that must stay private

| Path | Why |
|------|-----|
| `.env` / `/etc/portal/env` | `TELEGRAM_BOT_TOKEN`, `TGPORTAL_DB_KEY`, optional `PERSONA_ACCOUNT_QR`, `INVITE_URL` |
| `config.yml` | Local behavior (gitignored) |
| `folder/` (default `./data` or `/var/lib/portal`) | Account DBs, `tgportal.db`, caches |
| Owner BotFather tokens | Only in SQLite after `/pair-bot`, never in YAML |

Mode `0640` on env files. Do not commit them.

## Recommended host settings

- `database_encrypt: true` and `TGPORTAL_DB_KEY` from `openssl rand -hex 32`
- Private host: `telegram.allowed_user_ids` or `TELEGRAM_ALLOWED_USER_IDS`
- Public host: `persona.allow_groups: false`, keep ghost/bot caps
- Log to `stderr` for systemd; never log tokens
- Run as an unprivileged user (`User=portal` in the unit)
- SSH: keys only, no password login, deploy key forced-command if you use GitHub Actions

## Media on disk

Temporary downloads live under `folder/tg-cache` and `folder/dc-cache`. They must be deleted after a successful forward. Delta Chat may briefly write `dc.db-blobs`; the service sets a short device retention and unlinks ephemeral files. If you find leftover `.webm` / `.tgs` / user `.mp4` under the data dir, that is a bug — report it as in [SECURITY.md](../SECURITY.md).

Branding under `/var/lib/portal/assets` (logo, start animation) is not user media.

## Pairing

- Issue codes only in a private Telegram 1:1
- Reject pairing in groups
- Codes expire (`pairing.pending_ttl_sec`)
- Rotate `TELEGRAM_BOT_TOKEN` in BotFather if it leaks

## Deploy

Production SSH secrets live in the GitHub **production** environment. They are not given to pull requests. The deploy key can only run `portal-swap` (atomic binary replace + rollback).

## If something leaks

1. Revoke the Telegram token in BotFather  
2. Rotate `TGPORTAL_DB_KEY` and treat old `tgportal.db` as compromised  
3. Re-pair users  
4. If persona bots were registered, owners must revoke those BotFather tokens too  
