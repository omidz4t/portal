# You are trusting whoever runs the bot

Portal is not an end-to-end tunnel that hides you from the operator. The safest setup is to run your own instance.

## What the host can see

Anything you send to a Telegram bot arrives on the computer that is running Portal. The program downloads the file there and then sends it into Delta Chat. Anyone who can read that machine can read what you bridged.

That includes:

- Telegram user ids, display names, and the text of messages sent to the bot
- Pairing codes, and which Telegram account is tied to which Delta Chat (in the database)
- Stickers, GIFs, images, and videos while they sit in cache
- In persona mode: BotFather tokens you register, and the keys of ghost Delta Chat accounts

## What Delta Chat encryption does *not* hide

Delta Chat still encrypts traffic on the wire to chatmail. That protects the path between Portal and the mail server. It does **not** hide plaintext from Portal itself. The process that builds and sends the message is a participant, not a sealed pipe.

## The public bot

A public instance such as [@tgdeltabridgebot](https://t.me/tgdeltabridgebot) is fine for trying the product. Using it means you accept that operator’s machine. Do not bridge anything you would not hand to them.

## Host your own

Then the only runner is you. See [self-host](../self-host/).

If you run Portal for other people, publish a privacy policy ([docs/privacy.md](https://github.com/omidz4t/portal/blob/main/docs/privacy.md)) and never log BotFather tokens.
