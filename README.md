# Gotify 2 Telegram
This Gotify plugin forwards all received messages to Telegram through the Telegram bot with support for MarkdownV2 formatting.

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

## Troubleshooting
1. When only the Gotify dashboard receives your message, but not Telegram:

    If, when making the API call to get your bot's chat ID, no data is returned, you may need to change the bot's privacy settings.

    - In the BotFather chat, list your created bots and select the respective bot for which you want to change the Group Privacy setting.
    - Turn off the Group Privacy setting.

## Please note
I am not a developer. In fact, the implementation for MarkdownV2 formatting support was entirely written by Claude AI (as ChatGPT was unable to figure it out). I have tested it and it works perfectly for my use case: forwarding notifications from my Proxmox instance to Telegram.

The code properly handles large messages by intelligently splitting them while preserving MarkdownV2 formatting. It also splits messages at natural boundaries (such as newlines and words) whenever possible to avoid exceeding Telegram's message size limit. Each part of a split message is labeled with "(n/N)" at the top, making it easier to track the message parts. Additionally, the code includes a small delay between messages to prevent hitting Telegram's rate limits.

I have built the plugin for Gotify V2.6.1 for AMD64, ARM7, and ARM64 platforms and added them as a release to this fork.

## Appendix
Mandatory secrets.

```(shell)
GOTIFY_HOST=ws://YOUR_GOTIFY_IP (depending on your setup, "ws://localhost:80" will likely work by default)
GOTIFY_CLIENT_TOKEN=YOUR_CLIENT_TOKEN (create a new Client in Gotify and use the Token from there, or you can use an existing client)
TELEGRAM_CHAT_ID=YOUR_TELEGRAM_CHAT_ID (conversation ID from the Telegram API call above)
TELEGRAM_BOT_TOKEN=YOUR_TELEGRAM_BOT_TOKEN (API token provided by BotFather)
```

