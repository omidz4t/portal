# Persona mode

Register your own Telegram bot. Each remote user becomes a stable ghost Delta Chat account that talks to you as themselves.

## Why it exists

Personal mode is you, the portal bot, and one Delta Chat chat. That is enough for “my stickers, my chat.”

Persona mode is for someone who already has (or wants) a Telegram bot of their own. Friends and groups write to **your** bot. On Delta Chat you should see those people, not a labeled dump from a shared bridge account.

## How ghosts work

For each person who messages your bot, Portal creates a real Delta Chat account (name and photo from Telegram). The next time, it is the same account.

That ghost delivers the person’s message to you as a normal 1:1 — no prefix, no “via Portal.” When you reply, the program sends the reply out through your BotFather bot.

Ghost tables are separate from personal-mode pairs. If Alice pairs with the public portal bot, that does not reuse Alice’s persona ghost.

## Owner setup

1. Set `mode: persona` or `mode: both`, and `PERSONA_ACCOUNT_QR` (a `dcaccount:` / `dclogin:` URI used to provision ghosts).
2. Send `/pair` on the **portal** bot and finish on Delta Chat. This stores your Delta Chat identity so ghosts can write to you without another invite.
3. Create a bot with BotFather. Copy the token. Never paste it into issues or chats.
4. In a private chat with the portal bot:

```text
/pair-bot <TOKEN> [https://t.me/YourBot]
```

Limits: `persona.max_ghosts`, `persona.max_bots`. List with `/bots`. Stop with `/unpair-bot [id|@user]`. After unpair, run `/pair-bot` again to start the same bot.

## Groups

Off by default. Set `persona.allow_groups: true`, then:

1. Add the persona bot to the Telegram group.
2. BotFather → your bot → Bot Settings → Group Privacy → Turn off. Otherwise Telegram only delivers mentions and commands.
3. Posts become a Delta Chat group named `TG: …`. Each speaker is their ghost; you are a member.
4. Your replies in that DC group go to Telegram **as the bot**, not as the original users.

## Trust

The instance that stores the BotFather token **is** your Telegram bot. Ghost account keys live under `./data`. If you send `/pair-bot` to someone else’s public Portal, you are handing them the bot and every conversation it will carry. Prefer [self-host](../self-host/). See [trust](../trust/).

Operator notes in the repo: [docs/persona.md](https://github.com/omidz4t/portal/blob/main/docs/persona.md) · [persona-design.md](https://github.com/omidz4t/portal/blob/main/docs/persona-design.md).
