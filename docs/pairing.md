# Pairing

Nothing is bridged until one **Telegram user** is linked to one **Delta Chat 1:1 chat**. That link is a *pair*. It is stored in `folder/tgportal.db` (default `./data/tgportal.db`).

This page is **personal** mode (`mode: personal` or `both`). Owner-owned Telegram bots are [persona.md](persona.md).

Using a hosted bot means you trust that operator. Prefer your own instance. See [trust.md](trust.md).

## What you need

- Telegram: a private 1:1 with the portal bot (example: [@tgdeltabridgebot](https://t.me/tgdeltabridgebot))
- Delta Chat or ArcaneChat
- The portal process running (`portal serve`) with a configured Delta Chat account (`portal init …`)

Pairing **only** works in a private chat. Groups are rejected so codes do not leak.

## Telegram first (usual)

1. Open the bot in a **private** chat. Send `/start` or `/pair` (same as `/connect`).
2. The bot replies with:
   - a **QR** of the Delta Chat invite
   - the invite **link**
   - an **8-character code** (length and lifetime are configurable)
3. Scan the QR or open the invite in Delta Chat / ArcaneChat and accept the chat.
4. Send **only the code** to the Delta Chat bot (letters are case-insensitive).
5. Both sides confirm. Stickers, GIFs, text, photos, and short video now go both ways.

If you are already paired, `/pair` says so and does not issue a new code.

## Delta Chat first

1. Message the **Delta Chat** bot while you are unpaired (or send `/pair` there).
2. It replies with a Telegram deep link: `https://t.me/<bot>?start=CODE`.
3. Open that link. Telegram sends `/start CODE` to the portal bot.
4. If the code is valid, pairing becomes **active immediately**.

`/start` without a code is the same as `/pair` (Telegram-first).

## After you are paired

| Direction | What moves |
|-----------|------------|
| Telegram → Delta Chat | Text, photos, short videos, stickers, Lottie, video stickers, GIFs, custom emoji |
| Delta Chat → Telegram | Same kinds, when the chat is the paired 1:1 |

Limits (duration / file size) are in `bridge.limits` — see [configuration.md](configuration.md).

Reply to a sticker on Telegram with `/send_pack` to push the **whole pack** (if `bridge.sticker_packs` is on).

## Telegram commands

| Command | |
|---------|--|
| `/start` | Welcome; or `/start CODE` from a Delta Chat link |
| `/pair` / `/connect` | Invite QR + link + new code |
| `/status` | Whether this Telegram user is paired |
| `/disconnect` | Unlink this Telegram user (bridge stops) |
| `/send_pack` | Reply to a sticker → full pack |
| `/help` | Command list |

Persona (only if `mode` is `persona` or `both`):

| Command | |
|---------|--|
| `/pair-bot <TOKEN> [https://t.me/YourBot]` | Register your BotFather bot |
| `/bots` | List your persona bots |
| `/unpair-bot [id\|@user]` | Stop a persona bot |

You must finish **personal** `/pair` before `/pair-bot`. Details: [persona.md](persona.md).

## Codes and database

| Setting | Default | Meaning |
|---------|---------|---------|
| `pairing.code_length` | `8` | 4–12 characters |
| `pairing.pending_ttl_sec` | `1800` | Unused code lifetime (30 minutes) |

A pending row is created on `/pair`. It becomes `active` when the other side sends the code. `/disconnect` clears the link for that Telegram user. A new `/pair` issues a new code.

Allow-list: if `telegram.allowed_user_ids` or `TELEGRAM_ALLOWED_USER_IDS` is set, only those Telegram IDs may pair.

## What can go wrong

| Symptom | Check |
|---------|--------|
| “Pairing only works in a private chat” | Use a 1:1, not a group |
| Code rejected / expired | `/pair` again; wait if you hit the rate limit |
| No invite / QR | `portal init` ran; DC account is configured; `portal serve` is up |
| Media does not appear | Pairing is active (`/status`); `bridge.*` flags; file size/duration limits |
| “Too many pairing attempts” | Wait a few minutes |

Do not paste codes in public chats or screenshots you will post.

## Operators

```bash
make serve
# or
./portal --config config.yml serve
```

Pairing state: `folder/tgportal.db` table `pairs` (do not commit `folder/`). Encrypted stores need `TGPORTAL_DB_KEY`. See [security.md](security.md).
