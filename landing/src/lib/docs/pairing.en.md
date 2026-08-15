# Personal pairing

Nothing is forwarded until a Telegram user is linked to a Delta Chat chat.

This is `mode: personal` (also available under `both`). For owner-owned bots see [persona](../persona/).

Using a hosted bot means you trust that runner. Prefer your own instance. See [trust](../trust/) · [self-host](../self-host/).

## Telegram first

1. Open the bot in a private 1:1 (e.g. [@tgdeltabridgebot](https://t.me/tgdeltabridgebot)) and send `/start` or `/pair`. Groups are rejected so codes do not leak.
2. You get a QR of the Delta Chat invite plus an 8-character code (about 30 minutes).
3. Open the invite in Delta Chat / ArcaneChat and accept the chat.
4. Send only the code (case-insensitive).

## Delta Chat first

1. Message the Delta Chat bot while unpaired (or send `/pair`).
2. It replies with `https://t.me/<bot>?start=CODE`.
3. Open that link on Telegram. Pairing is active immediately.

## Useful commands (Telegram)

- `/status` — pairing state
- `/disconnect` — unlink
- `/send_pack` — reply to a sticker to bridge the whole pack (if `bridge.sticker_packs` is on)

Full walkthrough in the repo: [docs/pairing.md](https://github.com/omidz4t/portal/blob/main/docs/pairing.md).
