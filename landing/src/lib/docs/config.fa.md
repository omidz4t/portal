# config.yml

فایل کامل پیکربندی برای مدیر سیستم: شامل تمامی کلیدها، مقادیر پیش‌فرض و توضیحات. اطلاعات حساس و محرمانه در فایل `.env` نگهداری می‌شوند. از روی فایل نمونه یک نسخه کپی تهیه کنید؛ هرگز فایل‌های `config.yml`، `.env` یا مسیر `./data` را به مخزن گیت ارسال (Commit) نکنید.

## هدف از تفکیک دو فایل پیکربندی

نرم‌افزار «پورتال» (Portal) تنظیمات اجرایی و اطلاعات محرمانه را از یکدیگر جدا می‌کند:

- `config.yml` — حالت اجرا، نمایه، فضای ذخیره‌سازی، ثبت وقایع (Logging)، پروکسی‌ها، رابط کاربری تلگرام و کلیدهای فعال‌سازی پل ارتباطی. این فایل در گیت نادیده گرفته می‌شود (Gitignored).
- `config.example.yml` — قالب ردیابی‌شده در مخزن. می‌توانید با دستور `make config` از روی آن کپی بسازید.
- `.env` — شامل `TELEGRAM_BOT_TOKEN`، متغیر اختیاری `PERSONA_ACCOUNT_QR`، `INVITE_URL`، `TGPORTAL_DB_KEY`، فهرست کاربران مجاز و آدرس‌های پروکسی. این فایل نیز در گیت نادیده گرفته می‌شود.
- `.env.example` — قالب ردیابی‌شده در مخزن برای اطلاعات محرمانه.

در صورت عدم وجود فایل‌های محلی، آن‌ها را به روش زیر ایجاد کنید:

```bash
make config
# or:
cp config.example.yml config.yml
cp .env.example .env
```

## فایل کامل config.yml

متن زیر نسخهٔ کامل فایلی است که پورتال آن را می‌خواند. توضیحات (کامنت‌ها) بخشی از این قالب هستند؛ هنگام کپی کردن، آن‌ها را حذف نکنید. مقادیر نمایش‌داده‌شده همان مقادیر پیش‌فرض مستندسازی‌شده هستند.

<!--full-config-->

## متغیرهای محیطی (`.env`)

هرگز توکن‌های دریافتی از BotFather را درون فایل YAML قرار ندهید. توکن‌های متعلق به کاربران که از طریق دستور `/pair-bot` ثبت می‌شوند، تنها در پایگاه دادهٔ SQLite و درون دایرکتوری `folder` ذخیره خواهند شد.

- `TELEGRAM_BOT_TOKEN` — جهت راه‌اندازی پل ارتباطی تلگرام الزامی است.
- `PERSONA_ACCOUNT_QR` — شناسهٔ URI از نوع `dcaccount:` یا `dclogin:` برای ایجاد و راه‌اندازی حساب‌های شبح (در حالت persona یا both).
- `INVITE_URL` — لینک دعوت اختیاری مدیر در «دلتا چت» (Delta Chat) برای نمایش در پیام راه‌اندازی اولیه.
- `TGPORTAL_DB_KEY` — کلید هگز ۳۲ بایتی (`openssl rand -hex 32`) در صورتی که گزینهٔ `database_encrypt` فعال (true) باشد.
- `TELEGRAM_ALLOWED_USER_IDS` — شناسه‌های کاربری تلگرام که با کاما از یکدیگر جدا شده‌اند؛ در صورت تعیین، مقدار `telegram.allowed_user_ids` را بازنویسی می‌کند.
- `PROXY_URL`، `TELEGRAM_PROXY_URL`، `DELTACHAT_PROXY_URL`، `PROXY_ENABLED` — متغیرهای اختیاری برای بازنویسی تنظیمات «پروکسی» (Proxy).

```bash
TELEGRAM_BOT_TOKEN=123456:ABC-DEF...
# PERSONA_ACCOUNT_QR=dcaccount:nine.testrun.org
# INVITE_URL=https://i.delta.chat/#...
# TGPORTAL_DB_KEY=
# TELEGRAM_ALLOWED_USER_IDS=123456789
# PROXY_URL=socks5://127.0.0.1:1080
# PROXY_ENABLED=true
# TELEGRAM_PROXY_URL=socks5://127.0.0.1:1080
# DELTACHAT_PROXY_URL=socks5://127.0.0.1:1080
```

## اولویت اعمال تنظیمات

دایرکتوری داده‌ها: ابتدا آرگومان خط فرمان `--folder` یا `-f`، سپس مقدار `folder` در فایل `config.yml` و در نهایت مسیر پیش‌فرض `./data`.

فهرست مجاز تلگرام: در صورت تنظیم، متغیر `TELEGRAM_ALLOWED_USER_IDS`، سپس مقدار `telegram.allowed_user_ids` در فایل پیکربندی و در غیر این صورت بدون محدودیت.

مسیر فایل پیکربندی: آرگومان خط فرمان `--config` یا `-c` (پیش‌فرض `config.yml`). در Makefile: دستور `make serve CONFIG=path/to.yml`.

## فایل‌های دارایی و زمان اجرا

- `assets/logo.jpg` — نشان تجاری تلگرام و تصویر نمایهٔ پیش‌فرض دلتا چت
- `assets/start_black_hole.mp4` — پویانمایی دستور `/start`
- `folder/accounts/` — پایگاه‌های دادهٔ حساب‌های دلتا چت
- `folder/tgportal.db` — محل ذخیره‌سازی داده‌های جفت‌سازی و حساب‌های پرسونا
- `folder/tg-cache/` — دایرکتوری بارگیری‌های موقت تلگرام

یادداشت‌های راهنمای مدیر در مخزن پروژه: [docs/configuration.md](https://github.com/omidz4t/portal/blob/main/docs/configuration.md) · [config.example.yml](https://github.com/omidz4t/portal/blob/main/config.example.yml) · [میزبانی شخصی](../self-host/) · [پرسونا](../persona/).
