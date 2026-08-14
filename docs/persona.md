# Persona mode (operator guide)

Full design: **[persona-design.md](persona-design.md)**.  
Public write-up: landing site `/docs/persona`. **Trust:** [trust.md](trust.md) — BotFather tokens and ghost keys live on the host; do not `/pair-bot` on a machine you do not run.

## Modes

| `mode` | Meaning |
|--------|---------|
| `personal` | Classic portal pairing only (default product) |
| `persona` | User-owned bots + per-TG ghost DC accounts |
| `both` | Both |

## Setup

```env
TELEGRAM_BOT_TOKEN=...              # portal
PERSONA_ACCOUNT_QR=dcaccount:…      # required to create ghosts
```

```yaml
mode: both
persona:
  max_ghosts: 200
  max_bots: 20
```

1. `/pair` on the **portal** bot and finish on Delta Chat  
   → Portal stores your DC **vcard/public key** for later ghosts.  
2. Create a bot with BotFather.  
3. On the portal bot (private chat):

   ```text
   /pair-bot <TOKEN> [https://t.me/YourBot]
   ```

4. Anyone who DMs your bot gets a **ghost DC account** (name + photo from Telegram). That account messages **you** as a normal 1:1 (no labels). Replies go back via your persona bot.

### Groups

Set `persona.allow_groups: true` (default is off). Then:

1. Add the persona bot to the Telegram group.  
2. **Disable privacy mode** in [@BotFather](https://t.me/BotFather):  
   `/mybots` → your bot → **Bot Settings** → **Group Privacy** → **Turn off**  
   (Otherwise Telegram only delivers @mentions/commands to the bot, not normal chat.)  
3. People post in the group → mirrored as a Delta Chat group `TG: …` with each speaker as their ghost; you are a member.  
4. Your replies in that DC group go to Telegram **as the bot** (Bot API cannot puppet users).

No `/persona-invite`. Key material comes from step 1.

## Commands (portal)

| Command | |
|---------|--|
| `/pair-bot <TOKEN> [url]` | Register persona bot (re-run after `/unpair-bot` to start the same bot again) |
| `/bots` | List |
| `/unpair-bot [id\|@user]` | Stop polling |

## If it fails

- `PERSONA_ACCOUNT_QR` empty → set in `.env` and restart  
- Owner key missing → re-run `/pair` (vcard capture was added for persona) then `/pair-bot` again  
- Telegram shows `Could not bridge…` with the real error  
- Ghost create fails mid-way → the new DC account is deleted (not left counting against chatmail quota)  

## Security

Bot tokens in SQLite under `folder`. Never commit `data/` or `.env`. Never log tokens.
