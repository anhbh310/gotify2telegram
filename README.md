# Telegram 2 Gotify
This Gotify plugin fowards all recived messages to Telegram through Telegram bot.

## Prerequisite
- A Telegram bot, bot token, and chat ID from bot conversation. You can get those information by following this [blog](https://medium.com/linux-shots/setup-telegram-bot-to-get-alert-notifications-90be7da4444).
- Golang, Docker, wget (If you want build the binary from source).

## Installation
* **By shared object**

    1. Get compatible shared object from [release](https://github.com/anhbh310/gotify2telegram/releases).

    2. Put it into Gotify plugin folder.

    3. Set secrets via environment variables (List of mandatory secrect is in [Appendix](#appendix)).

    4. Restart gotify.

* **Build from source**

    1. Change GOTIFY_VERSION in Makefile.

    2. Build the binary.

    ```
    make build
    ```

    3. Follow instruction from step 2 in shared object installation.


## Appendix
Mandatory secret.

```(shell)
GOTIFY_HOST=ws://YOUR_GOTIFY_IP
GOTIFY_CLIENT_TOKEN=YOUR_CLIENT_TOKEN
TELEGRAM_CHAT_ID=YOUR_TELEGRAM_CHAT_ID
TELEGRAM_BOT_TOKEN=YOUR_TELEGRAM_BOT_TOKEN
```
