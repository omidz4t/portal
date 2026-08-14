# Privacy Policy — Delta ↔️ TG (TGPORTAL)

**Last updated:** 27 July 2026  

| | |
|---|---|
| **Third-Party Service** | Delta ↔️ TG bridge bot (also known as **TGPORTAL**) |
| **Developer (operator)** | [Your name or legal entity] |
| **Contact for privacy requests** | [privacy@example.com] |
| **Telegram bot** | [@tgdeltabridgebot](https://t.me/tgdeltabridgebot) *(update to your bot)* |
| **Policy URL** | [Link where this document is published] |

This is a **separate, service-specific Privacy Policy** for Delta ↔️ TG.  
It is intended to work **together with** Telegram’s platform rules for bots, including:

- [Telegram Privacy Policy](https://telegram.org/privacy) (including [bot messages](https://telegram.org/privacy#6-bot-messages))  
- [Telegram Bot Terms](https://telegram.org/tos/bots)  
- [Standard Bot Privacy Policy](https://telegram.org/privacy-tpa) (Telegram’s default policy for bots that have not published their own)  

By using the Telegram bot, you also accept the Telegram Bot Terms and related platform terms. **Telegram Messenger Inc. is not the operator of this bot** and is not a party to this Policy as “Developer.”

> **Self-hosted copies:** If you run your own instance of the open-source software, **you** are the Developer/controller for that instance. Publish your own contact details and this (or an adapted) policy for your users.

> **Trust:** Using a hosted bot means you trust **that** operator with pairing data and every message or file the bridge touches. Delta Chat encryption does not hide content from `tgportal`. Prefer running your own instance. See [trust.md](trust.md).

---

## 1. Terms and definitions

Aligned with Telegram’s [Standard Bot Privacy Policy](https://telegram.org/privacy-tpa):

| Term | Meaning in this Policy |
|------|-------------------------|
| **Telegram / Platform** | The Telegram messaging platform and Bot API operated by Telegram Messenger Inc. |
| **Developer** | The person or legal entity that operates this Third-Party Service (see header). |
| **Third-Party Service / Service** | The Delta ↔️ TG Telegram bot and its paired Delta Chat bot account (the bridge). |
| **User / you** | A person who accesses the Service via Telegram and/or the linked Delta Chat chat. |
| **Policy** | This document. |
| **Delta Chat side** | The Delta Chat / chatmail account used by the bridge and the User’s paired Delta Chat conversation. |

---

## 2. General provisions

2.1. This Policy governs the relationship between **Developer** and **User** regarding the Service. It does **not** regulate the relationship between Telegram and its users and does **not** replace the [Telegram Privacy Policy](https://telegram.org/privacy).

2.2. The Service is an **independent third-party bot**. It is **not** maintained, endorsed, or affiliated with Telegram Messenger Inc. Developer is solely responsible for the Service.

2.3. Developer aims to follow privacy expectations set by platforms that distribute Telegram apps (including Apple and Google developer policies), to the extent applicable to a messaging bot.

2.4. This Policy describes how Developer **collects, stores, uses, shares, and protects** information when you use the Service.

2.5. **Continued use** of the Service (for example messaging the bot, pairing, or sending media to be bridged) constitutes acceptance of:

- this Policy;  
- the [Telegram Bot Terms](https://telegram.org/tos/bots); and  
- any additional terms Developer may publish for the Service.

2.6. If you do **not** accept these terms, **stop using the Service** immediately (and use `/disconnect` if you are paired).

2.7. Developer is responsible for ensuring this Policy matches the actual deployment and complies with applicable local law. This document is **not legal advice**.

---

## 3. Disclaimers

3.1. The Service is a third-party application on the Telegram Bot Platform. Developer identity is as stated in this Policy, the bot’s profile/about text, and any `/developer_info` (or similar) response if offered.

3.2. Developer may amend this Policy. Material changes will be reflected by updating the **Last updated** date and the published Policy URL. Where practical, Developer may also notify Users via the bot. It is your responsibility to review the current Policy.

3.3. You acknowledge that you have read and agree to the Telegram Bot Terms and Telegram’s privacy rules for bots, including that bots can receive messages and files you send them (see Telegram Privacy Policy, bot messages).

3.4. You warrant that you have the rights and legal capacity to use the Service under local law (including age rules).

3.5. Developer treats information you provide as submitted in good faith and is not obliged to verify every statement for accuracy.

3.6. Information you **choose to make public** (e.g. username, public channel content, content you ask the bridge to forward into a chat others can see) may be visible to others and is not protected solely by this Policy once disclosed that way.

---

## 4. What the Service does

Delta ↔️ TG is a **bridge**: after **pairing**, supported messages and media can be forwarded **between Telegram and Delta Chat** for your linked pair.

It is a **relay**, not an advertising network, not a public social feed, and not a permanent chat archive product for Users.

---

## 5. Collection of personal data

### 5.1. Data Telegram may provide to the bot (platform)

As described in the [Telegram Privacy Policy (bots)](https://telegram.org/privacy#6-bot-messages) and Bot API behaviour, when you interact with the bot, Telegram may deliver to Developer’s servers (via the Bot API) limited account and message data, for example:

- Telegram **user id**  
- **Username**, first/last name fields Telegram provides with updates (if any)  
- **Chat id** and message metadata  
- **Message content** you send to the bot (text, captions, stickers, photos, videos, documents, custom emoji entities, etc.)  
- File identifiers and downloadable media for messages you send the bot  

Developer does **not** receive your Telegram password. Developer does **not** automatically receive your phone number unless you **explicitly share** a contact/phone with the bot (this Service does not require phone number sharing for core pairing).

### 5.2. Additional data you send to the Service

Without limiting §5.1, the Service receives whatever you **voluntarily send** to the bot or to the paired Delta Chat bot account, including:

- Pairing commands and **pairing codes**  
- **Text** and **captions**  
- **Images**, **short videos**, **stickers** (including animated Lottie/TGS and video stickers), **GIFs**/animations  
- **Custom emoji** resolved through Telegram APIs when present in your messages  

### 5.3. Data created to operate pairing

To link Telegram and Delta Chat, Developer stores in a local database (typically SQLite under the bot’s data directory):

| Data | Purpose |
|------|---------|
| Telegram user id | Identify your Telegram account |
| Telegram username (if available) | Optional display (e.g. `/status`) |
| Telegram chat id | Deliver bridged messages to you on Telegram |
| Delta Chat account/chat ids | Deliver bridged messages on Delta Chat |
| Pairing codes & status | Complete and maintain pairing |
| Timestamps | Pairing lifecycle |

### 5.4. Delta Chat side

To deliver and receive Delta Chat messages, the Service operates a **Delta Chat bot account**. That involves:

- Account credentials/keys and mailbox data managed by **Delta Chat core** on the host  
- Message content and attachments in the paired chat  
- Peer **email/address** and display name as available from Delta Chat (e.g. for `/status`)  

Your own Delta Chat / chatmail **provider** also processes mail according to **their** privacy policy.

### 5.5. Technical / operational data

Depending on configuration:

- Application **logs** (if enabled; destination/level configurable; may be off by default)  
- Error messages shown to you (e.g. media rejected for length/size)  
- Network metadata needed to reach Telegram Bot API and mail/chatmail  
- Optional **proxy** settings if the operator routes traffic through a proxy  

### 5.6. Anonymous / non-identifying data

Developer may process limited technical diagnostics (e.g. aggregate error rates) that are **not** intended to identify you. This Service does **not** run third-party advertising trackers.

### 5.7. Mini apps

This Service is a **Telegram bot**, not a Telegram **Mini App**. Mini App–specific data collection described in Telegram’s Mini App Terms does **not** apply unless Developer later ships a mini app (in which case this Policy will be updated).

---

## 6. Processing of personal data (purposes)

6.1. Developer only requests, collects, processes, and stores data **necessary** for the Service to function, including:

| Purpose | Examples |
|---------|----------|
| **Provide the bridge** | Pairing, forwarding text/media Telegram ↔ Delta Chat |
| **User controls** | `/status`, `/disconnect`, pairing links (`t.me/…?start=CODE`) |
| **Security & abuse prevention** | Optional allow-list of Telegram user ids, disconnect, reject oversized media |
| **Operate & debug** | Logs (if enabled), error responses |
| **Legal compliance** | Respond to lawful requests where required |

6.2. **Legal grounds** (where GDPR/UK GDPR or similar laws apply) typically include:

- Processing **necessary to provide the Service you request** (pairing and bridging) — contract / steps at your request  
- **Legitimate interests** in securing and operating the Service, balanced against your rights  
- **Legal obligation** where applicable  

6.3. **No monetization of user data.** Developer does **not** sell personal data and does **not** use bridged message content for advertising or profiling outside the scope of providing the bridge, unless a future version clearly states otherwise and obtains any required consent.

6.4. Processing is limited to purposes stated in this Policy and needed to furnish and improve the **functionality of the bridge**.

---

## 7. Sharing and recipients

7.1. **You / your chats**  
Bridged content is delivered to the paired Telegram chat and/or Delta Chat chat. Anyone with access to those chats can see it.

7.2. **Infrastructure necessary to run the Service**

| Recipient | Role |
|-----------|------|
| **Telegram** | Bot API transport of updates and outbound bot messages |
| **Delta Chat / chatmail provider** | Transport and storage for the bot’s Delta Chat account |
| **Host / VPS** | Runs the binary, database, temporary files, optional logs |
| **Optional proxy provider** | Only if Developer configures a proxy |

These parties process data under **their** policies. Developer does not control Telegram’s or your mail provider’s systems.

7.3. Consistent with Telegram’s Standard Bot Privacy Policy principles: Developer will **not** share user data with unrelated third parties (including other bots or services of Developer) **unless**:

- you **explicitly authorize** it;  
- it is **required by law** (e.g. lawful court order); or  
- it is strictly necessary for the infrastructure in §7.2 to provide the Service you requested.

7.4. **No sale** of personal data.

---

## 8. Data protection and security

8.1. Developer employs reasonable technical and organisational measures appropriate to a small messaging bridge, including for example:

- Keeping bot tokens and secrets outside public repositories  
- Restricting filesystem access to the host data directory  
- Optional Telegram **allow-list** so only approved users can use the bot  
- Short-lived media caches intended to be deleted after processing  
- Pairing codes that should not be posted publicly  
- Transport security as provided by HTTPS/Telegram and your mail stack  

8.2. **No security measure is perfect.** Compromise of the host, bot token, or your devices may expose bridged content. If a Telegram bot token leaks, regenerate it in BotFather immediately.

8.3. Temporary media files are **not** intended as a permanent library of your chats. Pairing records remain until you disconnect or the operator deletes them.

---

## 9. Retention

| Data | Retention |
|------|-----------|
| Active pairing | Until `/disconnect`, operator deletion, or instance wipe |
| Pending pairing codes | Until used, replaced, or invalidated |
| Temporary download caches | Intended deletion after bridge attempt |
| Application logs | Only if enabled; duration depends on operator configuration |
| Delta Chat bot mailbox / blobs | Per Delta Chat core and provider practice |

There is no built-in long-term search product over your private messages beyond live bridging and the pairing database.

---

## 10. International processing

Telegram and many chatmail providers operate **across borders**. Message content and metadata may be processed outside your country subject to those providers’ practices and applicable transfer rules.

---

## 11. Children

The Service is not directed at children under **16** (or a higher age required in your country). Do not use the Service if you are under that age. Developer should not knowingly pair minors.

---

## 12. Your rights and Developer obligations

### 12.1. Rights you may have

Subject to applicable law (including GDPR-style rights where they apply), you may:

| Right | How it maps to this Service |
|-------|-----------------------------|
| **Access** | Request a copy of personal data Developer stores about you (primarily pairing records) |
| **Deletion** | Request deletion of stored personal data (subject to limited legal exceptions) |
| **Rectification** | Ask to correct inaccurate pairing data |
| **Restrict / object** | Object to certain processing where the law allows |
| **Withdraw / stop** | Stop using the Service; use `/disconnect`; revoke consent where processing was based on consent |
| **Complaint** | Lodge a complaint with a competent data protection authority |

Telegram’s Standard Bot Privacy Policy also expects that Users may request a copy of data the bot stored and request deletion (except essential data the law allows the Developer to keep).

### 12.2. Practical in-product controls

| Action | Command / step |
|--------|----------------|
| Stop bridging | `/disconnect` on Telegram and/or Delta Chat |
| See link info | `/status` (may show Telegram id and Delta Chat address) |
| Pair / re-pair | `/pair`, `/connect`, or Delta Chat pairing link |

### 12.3. How to submit a privacy request

Email **[privacy@example.com]** from a contact method that lets Developer verify you are the User (Developer may request reasonable identity verification if misuse is suspected).

Developer will:

- Provide an accessible way to read this Policy (this URL / bot about text)  
- Process lawful requests **within timeframes required by law**, and in any event **no later than 30 days** from a valid request (as contemplated by Telegram’s Standard Bot Privacy Policy for developers)  
- May impose **reasonable limits** on abusive repeated requests without undermining legal rights  

### 12.4. What Developer may retain after deletion requests

Where permitted by law, limited data may be retained (examples Telegram lists in the standard policy: legal obligations, legal claims, public interest, or tax/transactional requirements). For this bridge, that is typically rare; most Users can be fully unpaired and pairing rows deleted.

### 12.5. Telegram’s role in deletion on Platform

Telegram may delete data on **its** servers (messages, chats with the bot, or the bot itself) in response to abuse of the Platform, as described in Telegram’s Standard Bot Privacy Policy and Telegram’s own policies. That does not automatically erase the Developer’s host database unless Developer also deletes it.

---

## 13. What you should know before pairing

1. **Developer can technically access** messages and files that pass through the bridge on the host.  
2. **Telegram** and your **Delta Chat/mail provider** also process content under their policies.  
3. Bridged media often uses **neutral filenames**; **your captions** may still be sent with media.  
4. Oversized or **too-long videos** may be rejected (errors name which item failed).  
5. Do not send highly sensitive material through a bot or host you do not trust.  
6. Do not share live **pairing codes** in public channels.

---

## 14. Cookies and websites

The bot itself is not a website cookie service. If a project website uses cookies, that site needs its own notice.

---

## 15. Changes to this Privacy Policy

Developer may update this Policy when the Service or practices change. The **Last updated** date and published URL will be revised. Please check the Policy URL periodically. Continued use after changes constitutes acceptance where permitted by law.

*(Telegram’s platform-wide [Standard Bot Privacy Policy](https://telegram.org/privacy-tpa) may also change on Telegram’s site; this Service-specific Policy remains the primary description of **this** bot’s practices.)*

---

## 16. Open-source software

The TGPORTAL / Delta ↔️ TG source code may be released under an open-source licence (see `LICENSE`).  
**Publishing code does not make every author the controller of every deployment.** Each operator who runs an instance is Developer for that instance and must provide contact details and a privacy policy to their Users.

---

## 17. Contact

| | |
|---|---|
| **Developer** | [Your name or legal entity] |
| **Privacy requests** | [privacy@example.com] |
| **Telegram bot** | [@tgdeltabridgebot](https://t.me/tgdeltabridgebot) |
| **Source / project** | [https://github.com/omidz4t/portal] *(or your fork)* |

---

## 18. Operator checklist (before going public)

- [ ] Legal name of Developer  
- [ ] Working privacy contact email  
- [ ] Correct bot username in this Policy and BotFather “About”  
- [ ] Publish this Policy at a stable **HTTPS URL** and link it from the bot (BotFather description / commands / pinned message)  
- [ ] State whether logging is on and where logs go  
- [ ] State if `allowed_user_ids` (allow-list) is used, or that this is a **public** instance  
- [ ] Confirm `TGPORTAL_DB_KEY` is set if you advertise encrypted pairing storage  
- [ ] Hosting region (recommended)  
- [ ] Confirm no mini app (or update §5.7 if you add one)  
- [ ] Read [Telegram Bot Terms](https://telegram.org/tos/bots) and [privacy-tpa](https://telegram.org/privacy-tpa)  

---

## 19. Relationship to Telegram’s Standard Bot Privacy Policy

| Topic | How this Service treats it |
|-------|----------------------------|
| Separate policy | **Yes** — this document is Developer’s published Policy for the Service |
| Default TPA applies if none published | Operators **should publish** this (or better) so Users are not left only with the generic default |
| Data only as needed for features | Pairing + bridging only (§6) |
| No monetization of user data | Confirmed (§6.3) |
| No third-party sharing except as needed/legal | Confirmed (§7) |
| Security measures | Confirmed (§8) |
| User access & deletion; 30-day response target | Confirmed (§12) |
| Accept Bot Terms to use Service | Confirmed (§2.5) |
| Not affiliated with Telegram | Confirmed (§2.2, §3.1) |

---

*This Policy is written to align with Telegram’s [Standard Bot Privacy Policy](https://telegram.org/privacy-tpa) and to describe Delta ↔️ TG / TGPORTAL accurately. It is not legal advice. For commercial or regulated use, obtain advice for your jurisdiction.*
