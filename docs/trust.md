# Trust the operator

TGPORTAL is a **relay**. Messages and media you send the Telegram bot (or that arrive on the paired Delta Chat account) are **downloaded and re-sent by the machine that runs `tgportal`**.

That host can see:

- Telegram user ids, names, and whatever you send the bot
- Pairing codes and the SQLite map of Telegram ↔ Delta Chat
- Bridged files (stickers, GIFs, images, video) while they sit in cache
- In **persona** mode: BotFather tokens you register, plus ghost Delta Chat account keys created for remote users

Delta Chat transport encryption does **not** hide content from the bridge. The operator of the process is on the path. Telegram already delivered the update to that process.

## Prefer self-hosting

Using a public instance (for example [@tgdeltabridgebot](https://t.me/tgdeltabridgebot)) means you **trust that operator** with the traffic you bridge.

It is **better to host your own** copy:

1. `make config` and put **your** BotFather token in `.env`
2. `make init QR=dcaccount:…` on a provider you accept
3. `make serve` on a machine you control

Then the only “runner” is you. Still treat `./data` and `.env` as secrets.

Self-hosted operators are the controller for their users. Publish a privacy policy (see [privacy.md](privacy.md)) and do not log BotFather tokens.

Related: [security.md](security.md), [persona.md](persona.md).
