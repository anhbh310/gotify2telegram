# Telegram 2 Gotify
This Gotify plugin forwards all recieved messages to Telegram through the Telegram bot.

## Prerequisite
- A Telegram bot, bot token, and chat ID from bot conversation. You can get that information by following this [blog](https://medium.com/linux-shots/setup-telegram-bot-to-get-alert-notifications-90be7da4444).
- Golang, Docker, wget (If you want to build the binary from source).

## Installation
* **By shared object**

    1. Get the compatible shared object from [release](https://github.com/anhbh310/gotify2telegram/releases).

    2. Put it into Gotify plugin folder.

    3. Set secrets via environment variables (List of mandatory secrets is in [Appendix](#appendix)).

    4. Restart gotify.

* **Build from source**

    1. Change GOTIFY_VERSION in Makefile.

    2. Build the binary.

    ```
    make build
    ```

    3. Follow instructions from step 2 in the shared object installation.


## Appendix
Mandatory secrets.

```(shell)
GOTIFY_HOST=ws://YOUR_GOTIFY_IP
GOTIFY_CLIENT_TOKEN=YOUR_CLIENT_TOKEN
TELEGRAM_CHAT_ID=YOUR_TELEGRAM_CHAT_ID
TELEGRAM_BOT_TOKEN=YOUR_TELEGRAM_BOT_TOKEN
```
