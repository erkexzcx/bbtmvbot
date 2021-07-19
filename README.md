# bbtmvbot

[![License](https://img.shields.io/github/license/erkexzcx/bbtmvbot)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/erkexzcx/bbtmvbot)](https://goreportcard.com/report/github.com/erkexzcx/bbtmvbot)

This bot scans the most popular flat rent portals for latest posts in Vilnius, which will be sent to subscribed users using Telegram app.

Hardware requirements are so low that you can even run this completelly fine on a lowest-end SBC. On RPI0W, RAM usage is only about 8mb and CPU load is only few percent, so you can run this on any _potato_ you want :)

# Usage

Feel free to use my hosted instance in cloud: http://t.me/bbtmvbot

Otherwise see below steps on how to host it yourself.

1. Setup Telegram bot

Using [BotFather](https://t.me/BotFather) create your own Telegram bot.

Also using BotFather use command `/setcommands` and update your bot commands:
```
info - Information about BBTMV Bot
enable - Enable notifications
disable - Disable notifications
config - Configure bot settings
```
Once you set-up bot, you should have your bot's Telegram **API key**.

2. [Install Golang](https://golang.org/doc/install).

The most popular distros usually ships a very old Golang version in their official repositories, which might not work at all. Make sure to remove any existing Golang installations and install the latest version using [official upstream guide](https://golang.org/doc/install) for your operating system.

3. Install build dependencies
```
gcc
g++
```

Examples:
```bash
# Ubuntu/Debian
apt install gcc g++

# Fedora/RHEL
dnf install gcc g++
```

4. Build it
```
git clone https://github.com/erkexzcx/bbtmvbot.git
cd bbtmvbot
go build -ldflags="-s -w" -o bbtmvbot ./cmd/bbtmvbot/bbtmvbot.go
```

5. Create configuration file. Simply copy `config.example.yml` to a new file `config.yml` and edit accordingly.

6. Run it
```
cd <any_working_dir>
./bbtmvbot
```

**Tip**: Run it under SystemD script. :)
```
[Unit]
Description=BBTMVBOT service
After=network-online.target

[Service]
User=erikas
Group=erikas
WorkingDirectory=/home/erikas/bbtmvbot
ExecStart=/home/erikas/bbtmvbot/bbtmvbot
Restart=always
RestartSec=3

[Install]
WantedBy=multi-user.target
```
