# Security Policy

Portal is a **relay**. Anyone who can read the machine that runs `portal` can read pairing rows and bridged media. That is by design. See [docs/trust.md](docs/trust.md).

## Supported versions

Portal is **beta**. Only the **latest GitHub release** on [omidz4t/portal](https://github.com/omidz4t/portal/releases) gets security fixes. Older tags are unsupported.

## Reporting a vulnerability

**Do not** open a public issue for anything that could leak tokens, pairing codes, or give access to a host.

Email: **omid.zamani.4t@gmail.com**  
Subject: `[portal security]`

Include:

- Affected version or commit
- What you can do (read traffic, steal `TELEGRAM_BOT_TOKEN`, write the SQLite store, SSH, etc.)
- Steps to reproduce
- Whether a public instance (for example the documented Telegram bot) is affected

You should hear back within a few days. Please wait for a fix or coordinated disclosure before posting details.

Low-impact documentation mistakes or “I wish it logged less” notes can be public issues.

## What we treat as a security bug

- Secrets committed or written to world-readable files
- Pairing codes guessable or leaked in groups
- BotFather tokens logged
- Bridged stickers/videos left on disk after a successful forward
- Unauthenticated access to another user’s pairing
- SQL injection or store encryption bypass when `TGPORTAL_DB_KEY` is set
- A change that lets a pull request read GitHub Environment deploy secrets

## What we do not treat as a product bug

- The **operator** being able to read bridged messages (see [docs/trust.md](docs/trust.md))
- Telegram or chatmail outages
- A self-hoster using an empty allow-list on a public internet box

## Operators

Hardening checklist: [docs/security.md](docs/security.md).
